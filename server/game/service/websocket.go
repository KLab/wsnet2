package service

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"golang.org/x/xerrors"

	"wsnet2/log"
)

const (
	WebsocketRWTimeout = 5 * time.Minute
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

		svr := &http.Server{
			Handler:      &WSHandler{sv},
			ReadTimeout:  WebsocketRWTimeout,
			WriteTimeout: WebsocketRWTimeout,
		}
		errCh <- svr.Serve(listener)
	}()

	return errCh
}

func (sv *WSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}
