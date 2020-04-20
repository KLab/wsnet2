package service

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"golang.org/x/xerrors"

	"wsnet2/game"
	"wsnet2/log"
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
	*GameService
}

func (sv *GameService) serveWebSocket(ctx context.Context) <-chan error {
	errCh := make(chan error)

	go func() {
		laddr := sv.conf.WebsocketAddr
		log.Infof("game websocket: %#v", laddr)

		lc := net.ListenConfig{}
		listener, err := lc.Listen(ctx, "tcp", laddr)
		if err != nil {
			errCh <- xerrors.Errorf("listen failed: %w", err)
			return
		}

		if cert, key := sv.conf.TLSCert, sv.conf.TLSKey; cert != "" {
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

		svr := &http.Server{
			Handler:      r,
			ReadTimeout:  WebsocketRWTimeout,
			WriteTimeout: WebsocketRWTimeout,
		}
		errCh <- svr.Serve(listener)
	}()

	return errCh
}

func (s *WSHandler) HandleRoom(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomId := vars["id"]
	appId := r.Header.Get("X-Wsnet-App")
	clientId := r.Header.Get("X-Wsnet-User")

	repo, ok := s.repos[appId]
	if !ok {
		log.Debugf("WSHandler.handleRoom: invalid app id: %v", appId)
		http.Error(w, "Bad Request", 400)
		return
	}
	// TODO: authentication

	cli, err := repo.GetClient(roomId, clientId)
	if err != nil {
		log.Debugf("GetClient error: %v", err)
		// TODO: error format
		return
	}
	log.Debugf("client: %v", cli)

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		breq, _ := httputil.DumpRequest(r, false)
		log.Errorf("upgrade error: room=%v, client=%v, remote=%v: %v\nrequest=%v",
			roomId, clientId, err, r.RemoteAddr, string(breq))
		return
	}

	peer := game.NewPeer(ctx, cli, conn)
	<-peer.Done()
	log.Debugf("HandleRoom finished: room=%v, client=%v", roomId, clientId)
}
