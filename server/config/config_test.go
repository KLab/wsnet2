package config

import (
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestLoad(t *testing.T) {
	filename := "testdata/test.toml"

	c, err := Load(filename)
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}

	db := DbConf{
		Host:            "localhost",
		Port:            3306,
		DBName:          "wsnet2",
		AuthFile:        "dbauth",
		User:            "wsnetuser",
		Password:        "wsnetpass",
		ConnMaxLifetime: Duration(time.Minute * 3),
	}
	if diff := cmp.Diff(c.Db, db); diff != "" {
		t.Fatalf("c.Db differs: (-got +want)\n%s", diff)
	}

	hostname, _ := os.Hostname()
	game := GameConf{
		Hostname:   "wsnetgame.localhost",
		PublicName: hostname,
		RetryCount: 3,
		MaxRoomNum: 999999,

		MaxRooms:   123,
		MaxClients: 1234,

		DefaultMaxPlayers: 10,
		DefaultDeadline:   5,
		DefaultLoglevel:   2,

		HeartBeatInterval: Duration(time.Second * 10),

		ClientConf: ClientConf{
			EventBufSize:   512,
			WaitAfterClose: Duration(time.Second * 60),
			AuthKeyLen:     32,
		},

		LogConf: LogConf{
			LogStdoutConsole: true,
			LogStdoutLevel:   3,
			LogPath:          "/tmp/wsnet2-game.log",
			LogMaxSize:       1,
			LogMaxBackups:    2,
			LogMaxAge:        3,
			LogCompress:      true,
		},
	}
	if diff := cmp.Diff(c.Game, game); diff != "" {
		t.Fatalf("c.Game differs: (-got +want)\n%s", diff)
	}

	lobby := LobbyConf{
		Hostname:       "wsnetlobby.localhost",
		UnixPath:       "/tmp/sock",
		Net:            "tcp",
		Port:           8080,
		Loglevel:       2,
		ValidHeartBeat: Duration(time.Second * 30),
		AuthDataExpire: Duration(time.Second * 10),
		ApiTimeout:     Duration(time.Second * 5),
		HubMaxWatchers: 10000,
		LogConf: LogConf{
			LogStdoutConsole: false,
			LogStdoutLevel:   4,
			LogPath:          "/tmp/wsnet2-lobby.log",
			LogMaxSize:       500,
			LogMaxBackups:    0,
			LogMaxAge:        0,
			LogCompress:      false,
		},
	}
	if diff := cmp.Diff(c.Lobby, lobby); diff != "" {
		t.Fatalf("c.Lobby differs: (-got +want)\n%s", diff)
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
	want := "wsnetuser:wsnetpass@tcp(localhost:3306)/wsnet2?parseTime=true"
	if dsn := db.DSN(); dsn != want {
		t.Fatalf("DSN = %s, wants %s", dsn, want)
	}
}
