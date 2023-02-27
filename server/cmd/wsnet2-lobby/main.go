package main

import (
	"context"
	"fmt"
	"os"
	"runtime/debug"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"

	"wsnet2"
	"wsnet2/config"
	"wsnet2/lobby/service"
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

	defer log.InitLogger(&conf.Lobby.LogConf)()
	log.SetLevel(log.Level(conf.Lobby.Loglevel))
	log.Infof("WSNet2-Lobby")
	log.Infof("WSNet2Version: %v", wsnet2.Version)
	if bi, ok := debug.ReadBuildInfo(); ok {
		for _, s := range bi.Settings {
			if strings.HasPrefix(s.Key, "vcs.") {
				log.Infof("%v: %v", s.Key, s.Value)
			}
		}
	}

	db := sqlx.MustOpen("mysql", conf.Db.DSN())
	maxConns := conf.Lobby.DbMaxConns
	if maxConns > 0 {
		db.SetMaxOpenConns(maxConns)
		db.SetMaxIdleConns(maxConns)
		log.Infof("DbMaxConns: %v", maxConns)
	}
	db.SetConnMaxLifetime(time.Duration(conf.Db.ConnMaxLifetime))

	service, err := service.New(db, &conf.Lobby)
	if err != nil {
		panic(fmt.Errorf("%+v\n", err))
	}

	ctx := context.Background()

	err = service.Serve(ctx)
	if err != nil {
		panic(fmt.Errorf("%+v\n", err))
	}
}
