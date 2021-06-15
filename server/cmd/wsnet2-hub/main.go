package main

import (
	"context"
	"fmt"
	"os"
	"time"

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
	log.SetLevel(log.Level(conf.Hub.DefaultLoglevel))

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

	ctx := context.Background()

	err = service.Serve(ctx)
	if err != nil {
		panic(fmt.Errorf("%+v\n", err))
	}
}
