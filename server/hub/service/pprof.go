package service

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof"

	"wsnet2/log"
)

func (sv *HubService) servePprof(ctx context.Context) <-chan error {
	if sv.conf.PprofPort == 0 {
		return nil
	}

	errCh := make(chan error)

	sv.preparation.Add(1)
	go func() {
		laddr := fmt.Sprintf(":%d", sv.conf.PprofPort)
		log.Infof("hub pprof: %#v", laddr)

		sv.preparation.Done()
		errCh <- http.ListenAndServe(laddr, nil)
	}()

	return errCh
}
