package service

import (
	"context"
	_ "expvar"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"time"

	"wsnet2/log"
)

func (sv *GameService) servePprof(ctx context.Context) <-chan error {
	if sv.conf.PprofPort == 0 {
		return nil
	}

	http.HandleFunc("/debug/stop-the-db", func(w http.ResponseWriter, r *http.Request) {
		d := time.Second * 10
		if p := r.URL.Query().Get("d"); p != "" {
			var err error
			d, err = time.ParseDuration(p)
			if err != nil {
				w.WriteHeader(400)
				w.Write([]byte("bad duration"))
				return
			}
		}
		sv.db.SetMaxOpenConns(1)
		conn, err := sv.db.Conn(context.Background())
		if err != nil {
			time.Sleep(d)
			conn.Close()
		}
		cs := sv.conf.DbMaxConns
		sv.db.SetMaxOpenConns(cs)
		sv.db.SetMaxIdleConns(cs)
		_, _ = w.Write([]byte(fmt.Sprintf("%+v", sv.db.Stats())))
	})

	errCh := make(chan error)

	sv.preparation.Add(1)
	go func() {
		laddr := fmt.Sprintf(":%d", sv.conf.PprofPort)
		log.Infof("game pprof: %#v", laddr)

		sv.preparation.Done()
		errCh <- http.ListenAndServe(laddr, nil)
	}()

	return errCh
}
