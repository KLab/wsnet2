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

	"github.com/go-chi/chi/v5"
	"github.com/shiguredo/websocket"
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
	*GameService
}

func (sv *GameService) serveWebSocket(ctx context.Context) <-chan error {
	errCh := make(chan error)

	sv.preparation.Add(1)
	go func() {
		laddr := fmt.Sprintf(":%d", sv.conf.WebsocketPort)
		log.Infof("game websocket: %#v", laddr)

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
		r := chi.NewMux()
		r.Get("/room/{id:[0-9a-f]+}", ws.HandleRoom)

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
	roomId := chi.URLParam(r, "id")
	appId := r.Header.Get("Wsnet2-App")
	clientId := r.Header.Get("Wsnet2-User")
	logger := log.GetLoggerWith(
		log.KeyHandler, "ws:room",
		log.KeyRoom, roomId,
		log.KeyApp, appId,
		log.KeyClient, clientId,
		log.KeyRequestedAt, float64(time.Now().UnixNano()/1000000)/1000,
	)
	lastEvSeq, err := strconv.Atoi(r.Header.Get("Wsnet2-LastEventSeq"))
	if err != nil {
		logger.Infof("websocket: invalid header: LastEventSeq=%v, %+v", r.Header.Get("Wsnet2-LastEventSeq"), err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	repo, ok := s.repos[appId]
	if !ok {
		logger.Infof("websocket: invalid appId: %v", appId)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	cli, err := repo.GetClient(roomId, clientId)
	if err != nil {
		logger.Infof("websocket: repo.GetClient: %v", err)
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	logger.Infof("websocket: room=%v client=%v lastEv=%v", roomId, clientId, lastEvSeq)

	var authData string
	if ad := r.Header.Get("Authorization"); strings.HasPrefix(ad, "Bearer ") {
		authData = ad[len("Bearer "):]
	}
	if err := cli.ValidAuthData(authData); err != nil {
		logger.Infof("websocket: Authorization: %+v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		breq, _ := httputil.DumpRequest(r, false)
		logger.Errorf("websocket: upgrade: %+v\nrequest: %v", err, string(breq))
		return
	}
	metrics.Conns.Add(1)
	defer metrics.Conns.Add(-1)

	peer, err := game.NewPeer(ctx, cli, conn, lastEvSeq)
	if err != nil {
		logger.Warnf("websocket: NewPeer: %+v", err)
		return
	}
	<-peer.Done()
	logger.Debugf("websocket: finish: room=%v client=%v peer=%p", roomId, clientId, peer)
}
