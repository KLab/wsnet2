package main

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"

	"wsnet2/config"
	"wsnet2/game/service"
)

func main() {

	conf, err := config.Load("local.toml")
	if err != nil {
		panic(err)
	}

	db := sqlx.MustOpen("mysql", conf.Db.DSN())

	service, err := service.New(db, &conf.Game)
	if err != nil {
		panic(err)
	}

	err = service.Serve()
	if err != nil {
		panic(err)
	}
}
