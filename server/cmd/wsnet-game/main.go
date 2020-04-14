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

func main() {
	log.SetLevel(log.ALL)

	if len(os.Args) < 2 {
		panic(fmt.Errorf("no config.toml specified"))
	}
	conf, err := config.Load(os.Args[1])
	if err != nil {
		panic(fmt.Errorf("%+v\n", err))
	}

	db := sqlx.MustOpen("mysql", conf.Db.DSN())

	service, err := service.New(db, &conf.Game)
	if err != nil {
		panic(fmt.Errorf("%+v\n", err))
	}

	ctx := context.Background()

	err = service.Serve(ctx)
	if err != nil {
		panic(fmt.Errorf("%+v\n", err))
	}
}
