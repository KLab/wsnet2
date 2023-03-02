package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"

	"wsnet2"
	"wsnet2/config"
	"wsnet2/hub/service"
	"wsnet2/log"
)

func main() {
	if len(os.Args) < 2 {
		panic(fmt.Errorf("no config.toml specified"))
	}
	conf, err := config.Load(os.Args[1])
	if err != nil {
		panic(fmt.Errorf("%+v\n", err))
	}

	defer log.InitLogger(&conf.Hub.LogConf)()
	log.SetLevel(log.Level(conf.Hub.DefaultLoglevel))
	log.Infof("WSNet2-Hub")
	log.Infof("WSNet2Version: %v", wsnet2.Version)
	if bi, ok := debug.ReadBuildInfo(); ok {
		for _, s := range bi.Settings {
			if strings.HasPrefix(s.Key, "vcs.") {
				log.Infof("%v: %v", s.Key, s.Value)
			}
		}
	}

	db := sqlx.MustOpen("mysql", conf.Db.DSN())
	maxConns := conf.Hub.DbMaxConns
	if maxConns > 0 {
		db.SetMaxOpenConns(maxConns)
		db.SetMaxIdleConns(maxConns)
		log.Infof("DbMaxConns: %v", maxConns)
	}
	db.SetConnMaxLifetime(time.Duration(conf.Db.ConnMaxLifetime))

	service, err := service.New(db, &conf.Hub)
	if err != nil {
		panic(fmt.Errorf("%+v\n", err))
	}
	log.Infof("HostID: %v", service.HostId)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGTERM)
		select {
		case <-ctx.Done():
		case sig := <-ch:
			log.Infof("got signal: %v", sig)
			service.Shutdown(ctx)
		}
	}()

	err = service.Serve(ctx)
	if err != nil {
		panic(fmt.Errorf("%+v\n", err))
	}
}
