package service

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof"

	"wsnet2/log"
)

func (sv *LobbyService) servePprof(ctx context.Context) <-chan error {
	if sv.conf.PprofPort == 0 {
		return nil
	}

	errCh := make(chan error)

	go func() {
		laddr := fmt.Sprintf(":%d", sv.conf.PprofPort)
		log.Infof("lobby pprof: %#v", laddr)

		errCh <- http.ListenAndServe(laddr, nil)
	}()

	return errCh
}
