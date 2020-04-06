package config

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestLoad(t *testing.T) {
	filename := "testdata/test.toml"

	c, err := Load(filename)
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}

	db := DbConf{
		Host:     "localhost",
		Port:     3306,
		DBName:   "wsnet2",
		AuthFile: "dbauth",
		User:     "wsnetuser",
		Password: "wsnetpass",
	}
	if diff := cmp.Diff(c.Db, db); diff != "" {
		t.Fatalf("c.Db differs: (-got +want)\n%s", diff)
	}

	game := GameConf{
		Hostname:   "wsnetgame.localhost",
		RetryCount: 3,
		MaxRoomNum: 999999,
	}
	if diff := cmp.Diff(c.Game, game); diff != "" {
		t.Fatalf("c.Db differs: (-got +want)\n%s", diff)
	}
}

func TestDbConf_DSN(t *testing.T) {
	db := DbConf{
		Host:     "localhost",
		Port:     3306,
		DBName:   "wsnet2",
		AuthFile: "dbauth",
		User:     "wsnetuser",
		Password: "wsnetpass",
	}
	want := "wsnetuser:wsnetpass@tcp(localhost:3306)/wsnet2"
	if dsn := db.DSN(); dsn != want {
		t.Fatalf("DSN = %s, wants %s", dsn, want)
	}
}
