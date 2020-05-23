package service

import (
	"context"
	"net/http"
	_ "net/http/pprof"

	"wsnet2/log"
)

func (sv *GameService) servePprof(ctx context.Context) <-chan error {
	if sv.conf.PprofAddr == "" {
		return nil
	}

	errCh := make(chan error)

	sv.preparation.Add(1)
	go func() {
		laddr := sv.conf.PprofAddr
		log.Infof("game pprof: %#v", laddr)

		sv.preparation.Done()
		errCh <- http.ListenAndServe(laddr, nil)
	}()

	return errCh
}
