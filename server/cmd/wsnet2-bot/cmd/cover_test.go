//go:build scenariocoverage

package cmd

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"

	"wsnet2/config"
	gameserv "wsnet2/game/service"
	hubserv "wsnet2/hub/service"
	lobbyserv "wsnet2/lobby/service"
	"wsnet2/log"
)

/*
docker run -d --rm \
  -v `pwd`/../../../sql/10-schema.sql:/docker-entrypoint-initdb.d/10-schema.sql \
  -v `pwd`/../../../sql/90-docker.sql:/docker-entrypoint-initdb.d/90-docker.sql \
  -e MYSQL_ROOT_PASSWORD=root \
  -e MYSQL_DATABASE=wsnet2 \
  -e MYSQL_USER=wsnet2 \
  -e MYSQL_PASSWORD=wsnet2pass \
  -p 3306:3306 mysql:8.0

go test . -cover -coverprofile=cover.out -tags scenariocoverage \
  -coverpkg=wsnet2/auth,wsnet2/binary,wsnet2/common,wsnet2/config,wsnet2/game,wsnet2/game/service,wsnet2/hub,wsnet2/hub/service,wsnet2/lobby,wsnet2/lobby/service,wsnet2/log,wsnet2/pb

go tool cover -html=cover.out -o cover.html
*/

func TestMain(t *testing.T) {
	conf := config.Config{
		Db: config.DbConf{
			Host:            "localhost",
			Port:            3306,
			DBName:          "wsnet2",
			User:            "wsnet2",
			Password:        "wsnet2pass",
			ConnMaxLifetime: config.Duration(3 * time.Minute),
		},
		Game: config.GameConf{
			Hostname:          "localhost",
			PublicName:        "localhost",
			GRPCPort:          19000,
			WebsocketPort:     8000,
			PprofPort:         0,
			TLSCert:           "",
			TLSKey:            "",
			RetryCount:        5,
			MaxRoomNum:        999,
			MaxRooms:          100,
			MaxClients:        100,
			DefaultMaxPlayers: 10,
			DefaultDeadline:   5,
			DefaultLoglevel:   3,
			HeartBeatInterval: config.Duration(2 * time.Second),
			DbMaxConns:        0,
			ClientConf: config.ClientConf{
				EventBufSize:   128,
				WaitAfterClose: config.Duration(30 * time.Second),
				AuthKeyLen:     32,
			},
			LogConf: config.LogConf{
				LogStdoutConsole: true,
				LogStdoutLevel:   3,
				LogPath:          "",
				LogMaxSize:       0,
				LogMaxBackups:    0,
				LogMaxAge:        0,
				LogCompress:      false,
			},
		},
		Hub: config.HubConf{
			Hostname:          "localhost",
			PublicName:        "localhost",
			GRPCPort:          19001,
			WebsocketPort:     8001,
			PprofPort:         0,
			TLSCert:           "",
			TLSKey:            "",
			MaxClients:        100,
			DefaultLoglevel:   3,
			ValidHeartBeat:    config.Duration(5 * time.Second),
			HeartBeatInterval: config.Duration(2 * time.Second),
			NodeCountInterval: config.Duration(1 * time.Second),
			DbMaxConns:        0,
			ClientConf: config.ClientConf{
				EventBufSize:   128,
				WaitAfterClose: config.Duration(30 * time.Second),
				AuthKeyLen:     32,
			},
			LogConf: config.LogConf{
				LogStdoutConsole: true,
				LogStdoutLevel:   3,
				LogPath:          "",
				LogMaxSize:       0,
				LogMaxBackups:    0,
				LogMaxAge:        0,
				LogCompress:      false,
			},
		},
		Lobby: config.LobbyConf{
			Hostname:       "localhost",
			UnixPath:       "",
			Net:            "tcp",
			Port:           8080,
			PprofPort:      0,
			Loglevel:       3,
			ValidHeartBeat: config.Duration(5 * time.Second),
			AuthDataExpire: config.Duration(time.Minute),
			ApiTimeout:     config.Duration(5 * time.Second),
			HubMaxWatchers: 100,
			DbMaxConns:     0,
			LogConf: config.LogConf{
				LogStdoutConsole: true,
				LogStdoutLevel:   3,
				LogPath:          "",
				LogMaxSize:       0,
				LogMaxBackups:    0,
				LogMaxAge:        0,
				LogCompress:      false,
			},
		},
	}

	defer log.InitLogger(&conf.Lobby.LogConf)()
	log.SetLevel(log.Level(conf.Lobby.Loglevel))
	logger = log.Get(log.INFO)

	db := sqlx.MustOpen("mysql", conf.Db.DSN())
	db.SetConnMaxLifetime(time.Duration(conf.Db.ConnMaxLifetime))

	lobby := must(lobbyserv.New(db, &conf.Lobby))
	game := must(gameserv.New(db, &conf.Game))
	hub := must(hubserv.New(db, &conf.Hub))

	ctx := context.Background()
	cerr := make(chan error, 4)
	var wg sync.WaitGroup
	wg.Add(3)

	lbctx, lbcancel := context.WithCancel(ctx)
	go func() {
		err := lobby.Serve(lbctx)
		t.Log("lobby", err)
		cerr <- err
		wg.Done()
	}()
	go func() {
		err := game.Serve(ctx)
		t.Log("game", err)
		cerr <- err
		wg.Done()
	}()
	go func() {
		err := hub.Serve(ctx)
		t.Log("hub", err)
		cerr <- err
		wg.Done()
	}()

	time.Sleep(time.Second * 5)

	// do scenario
	lobbyURL = "http://localhost:8080"
	appId = "testapp"
	appKey = "testapppkey"

	for n, scenario := range scenarios {
		err := scenario(ctx)
		t.Log(n, err)
		if err != nil {
			cerr <- err
			break
		}
	}

	time.Sleep(time.Second * 5)

	lbcancel()
	game.Shutdown(ctx)
	hub.Shutdown(ctx)

	wg.Wait()
	close(cerr)

	errs := make([]error, 0, 4)
	for e := range cerr {
		errs = append(errs, e)
	}
	if err := errors.Join(errs...); err != nil {
		t.Fatal(err)
	}
}

func must[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}
	return t
}
