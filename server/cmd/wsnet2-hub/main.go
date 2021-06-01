package main

import (
	"context"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"

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
	log.SetLevel(log.Level(conf.Game.DefaultLoglevel))

	db := sqlx.MustOpen("mysql", conf.Db.DSN())
	db.SetMaxOpenConns(conf.Hub.DbConnConf.DbMaxOpenConns)
	db.SetMaxIdleConns(conf.Hub.DbConnConf.DbMaxIdleConns)
	log.Infof("MaxOpenConns: %v, MaxIdleConns: %v", conf.Hub.DbConnConf.DbMaxOpenConns, conf.Hub.DbConnConf.DbMaxIdleConns)

	service, err := service.New(db, &conf.Hub)
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
