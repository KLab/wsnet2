package service

import (
	"context"
	"net/http"
	_ "net/http/pprof"

	"wsnet2/log"
)

func (sv *LobbyService) servePprof(ctx context.Context) <-chan error {
	if sv.conf.PprofAddr == "" {
		return nil
	}

	errCh := make(chan error)

	go func() {
		laddr := sv.conf.PprofAddr
		log.Infof("lobby pprof: %#v", laddr)

		errCh <- http.ListenAndServe(laddr, nil)
	}()

	return errCh
}
