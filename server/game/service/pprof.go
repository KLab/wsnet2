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

	// DB接続を擬似的に詰まった状態にする
	http.HandleFunc("/debug/stop-the-db", func(w http.ResponseWriter, r *http.Request) {
		d := time.Second * 10
		if p := r.URL.Query().Get("d"); p != "" {
			var err error
			d, err = time.ParseDuration(p)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(fmt.Sprintf("invalid duration: %q; %v", p, err)))
				return
			}
		}

		// SetMaxOpenConns(0) は無制限にDB接続することになる。
		// 1本の接続を握って SetMaxOpenConns(1) することでDB接続が詰まった状況を作る。
		conn, err := sv.db.Conn(context.Background())
		if err != nil {
			log.Errorf("/debug/stop-the-db: failed to get db conn: %+v", err)

			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(fmt.Sprintf("failed to get db conn: %v\n", err)))
			return
		}

		sv.db.SetMaxOpenConns(1)
		time.Sleep(d)

		conn.Close()
		cs := sv.conf.DbMaxConns
		sv.db.SetMaxOpenConns(cs)
		sv.db.SetMaxIdleConns(cs)

		_, _ = w.Write([]byte(fmt.Sprintf("%+v\n", sv.db.Stats())))
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
