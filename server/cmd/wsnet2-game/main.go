package main

import (
	"context"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"

	"wsnet2/config"
	"wsnet2/game/service"
	"wsnet2/log"
)

var (
	WSNet2Version string = "LOCAL"
	WSNet2Commit  string = "LOCAL"
)

func main() {
	if len(os.Args) < 2 {
		panic(fmt.Errorf("no config.toml specified"))
	}
	conf, err := config.Load(os.Args[1])
	if err != nil {
		panic(fmt.Errorf("%+v\n", err))
	}

	defer log.InitLogger(&conf.Game.LogConf)()
	log.SetLevel(log.Level(conf.Game.DefaultLoglevel))
	log.Infof("WSNet2-Game")
	log.Infof("WSNet2Version: %v", WSNet2Version)
	log.Infof("WSNet2Commit: %v", WSNet2Commit)

	db := sqlx.MustOpen("mysql", conf.Db.DSN())
	maxConns := conf.Game.DbMaxConns
	if maxConns > 0 {
		db.SetMaxOpenConns(maxConns)
		db.SetMaxIdleConns(maxConns)
		log.Infof("DbMaxConns: %v", maxConns)
	}

	service, err := service.New(db, &conf.Game)
	if err != nil {
		panic(fmt.Errorf("%+v\n", err))
	}
	log.Infof("HostID: %v", service.HostId)

	ctx := context.Background()

	err = service.Serve(ctx)
	if err != nil {
		panic(fmt.Errorf("%+v\n", err))
	}
}
