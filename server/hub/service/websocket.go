package service

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"golang.org/x/xerrors"

	"wsnet2/game"
	"wsnet2/log"
	"wsnet2/metrics"
)

const (
	WebsocketRWTimeout = 5 * time.Minute
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  4000,
		WriteBufferSize: 4000,
		Subprotocols:    []string{"wsnet2"},
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
)

type WSHandler struct {
	*HubService
}

func (sv *HubService) serveWebSocket(ctx context.Context) <-chan error {
	errCh := make(chan error)

	sv.preparation.Add(1)
	go func() {
		laddr := fmt.Sprintf(":%d", sv.conf.WebsocketPort)
		log.Infof("hub websocket: %#v", laddr)

		lc := net.ListenConfig{}
		listener, err := lc.Listen(ctx, "tcp", laddr)
		if err != nil {
			errCh <- xerrors.Errorf("listen failed: %w", err)
			return
		}

		scheme := "ws"
		if cert, key := sv.conf.TLSCert, sv.conf.TLSKey; cert != "" {
			scheme = "wss"
			log.Infof("loading tls key: %#v", cert)
			cert, err := tls.LoadX509KeyPair(cert, key)
			if err != nil {
				errCh <- xerrors.Errorf("x509 load error: %w", err)
				return
			}
			tlsConf := &tls.Config{
				Certificates: []tls.Certificate{cert},
			}
			listener = tls.NewListener(listener, tlsConf)
		}

		ws := &WSHandler{sv}
		r := mux.NewRouter()
		r.HandleFunc("/room/{id:[0-9a-f]+}", ws.HandleRoom).Methods("GET")

		sv.wsURLFormat = fmt.Sprintf("%s://%s:%d/room/%%s",
			scheme, sv.conf.PublicName, sv.conf.WebsocketPort)

		svr := &http.Server{
			Handler:      r,
			ReadTimeout:  WebsocketRWTimeout,
			WriteTimeout: WebsocketRWTimeout,
		}
		sv.preparation.Done()
		errCh <- svr.Serve(listener)
	}()

	return errCh
}

func (s *WSHandler) HandleRoom(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomId := vars["id"]
	clientId := r.Header.Get("Wsnet2-User")
	lastEvSeq, err := strconv.Atoi(r.Header.Get("Wsnet2-LastEventSeq"))
	if err != nil {
		log.Debugf("WSHandler.HandleRoom invalid header: Wsnet2-LastEventSeq", r.Header.Get("Wsnet2-LastEventSeq"))
		http.Error(w, "Bad Request", 400)
		return
	}

	cli, err := s.repo.GetClient(roomId, clientId)
	if err != nil {
		log.Debugf("WSHandler.HandleRoom: GetClient error: %v", err)
		// TODO: error format
		http.Error(w, "Bad Request", 400)
		return
	}
	log.Debugf("client: %v", cli)

	var authData string
	if ad := r.Header.Get("Authorization"); strings.HasPrefix(ad, "Bearer ") {
		authData = ad[len("Bearer "):]
	}
	if err := cli.ValidAuthData(authData); err != nil {
		log.Debugf("WSHandler.HandleRoom: Authenticate failure: room=%v, client=%v, authdata=%v", roomId, clientId, authData)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		breq, _ := httputil.DumpRequest(r, false)
		log.Errorf("WSHandler.HandleRoom upgrade error: room=%v, client=%v, remote=%v: %v\nrequest=%v",
			roomId, clientId, err, r.RemoteAddr, string(breq))
		return
	}

	peer, err := game.NewPeer(ctx, cli, conn, lastEvSeq)
	if err != nil {
		log.Errorf("WSHandler.HandleRoom new peer error: %v", err)
		return
	}
	metrics.Conns.Add(1)
	<-peer.Done()
	metrics.Conns.Add(-1)
	log.Debugf("HandleRoom finished: room=%v client=%v peer=%p", roomId, clientId, peer)
}
