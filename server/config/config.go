package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"golang.org/x/xerrors"
)

type Config struct {
	Db    DbConf `toml:"Database"`
	Game  GameConf
	Lobby LobbyConf
}

type DbConf struct {
	Host     string
	Port     int
	DBName   string
	AuthFile string
	User     string
	Password string
}

type GameConf struct {
	// Hostname : Lobbyなどからのアクセス名. see GameConf.setHost()
	Hostname string
	// Hostname : クライアントからのアクセス名. see GameConf.setHost()
	PublicName string `toml:"public_name"`

	GRPCPort      int `toml:"grpc_port"`
	WebsocketPort int `toml:"websocket_port"`
	PprofPort     int `toml:"pprof_port"`

	TLSCert string `toml:"tls_cert"`
	TLSKey  string `toml:"tls_key"`

	RetryCount int `toml:"retry_count"`
	MaxRoomNum int `toml:"max_room_num"`

	// TODO: MaxRooms 最大部屋数

	DefaultMaxPlayers uint32 `toml:"default_max_players"`
	DefaultDeadline   uint32 `toml:"default_deadline"`
	DefaultLoglevel   uint32 `toml:"default_loglevel"`

	HeartBeatInterval Duration `toml:"heartbeat_interval"`
}

type LobbyConf struct {
	Hostname  string
	Net       string
	Port      int
	PprofPort int `toml:"pprof_port"`

	Loglevel uint32 `toml:"loglevel"`

	// ValidHeartBeat : HeartBeatの有効期間
	ValidHeartBeat Duration `toml:"valid_heartbeat"`
}

type Duration time.Duration

func (d *Duration) UnmarshalText(text []byte) error {
	td, err := time.ParseDuration(string(text))
	*d = Duration(td)
	return err
}

func Load(conffile string) (*Config, error) {
	c := &Config{
		// set default values before decode file.
		Game: GameConf{
			RetryCount: 5,
			MaxRoomNum: 999999,

			DefaultMaxPlayers: 10,
			DefaultDeadline:   5,
			DefaultLoglevel:   2,

			HeartBeatInterval: Duration(2 * time.Second),
		},
		Lobby: LobbyConf{
			ValidHeartBeat: Duration(5 * time.Second),
			Loglevel:       2,
		},
	}

	_, err := toml.DecodeFile(conffile, c)
	if err != nil {
		return nil, err
	}

	err = c.Db.loadAuthfile(conffile)
	if err != nil {
		return nil, err
	}

	c.Game.setHost()

	return c, nil
}

func (db *DbConf) loadAuthfile(conffile string) error {
	if db.AuthFile == "" {
		return nil
	}
	authfile := db.AuthFile
	if authfile[0] != '/' {
		authfile = path.Join(path.Dir(conffile), authfile)
	}
	content, err := ioutil.ReadFile(authfile)
	if err != nil {
		return err
	}
	ss := strings.SplitN(string(content), ":", 2)
	if len(ss) != 2 {
		return xerrors.Errorf("Db authfile format error: %q", string(content))
	}

	db.User = ss[0]
	db.Password = ss[1]
	return nil
}

func (db *DbConf) DSN() string {
	user := db.User
	if db.Password != "" {
		user = fmt.Sprintf("%s:%s", db.User, db.Password)
	}
	return fmt.Sprintf("%s@tcp(%s:%d)/%s?parseTime=true", user, db.Host, db.Port, db.DBName)
}

// setHost : Hostname/PublicNameを設定する
// 優先順位
//  1: Configファイル
//  2: 環境変数
//     - WSNET_GAME_HOSTNAME
//     - WSNET_GAME_PUBLICNAME
//  3: os.Hostname()
//  4: "localhost"
//
func (game *GameConf) setHost() {
	if game.Hostname == "" {
		if h := os.Getenv("WSNET_GAME_HOSTNAME"); h != "" {
			game.Hostname = h
		} else if h, err := os.Hostname(); err == nil {
			game.Hostname = h
		} else {
			game.Hostname = ""
		}
	}
	if game.PublicName == "" {
		if h := os.Getenv("WSNET_GAME_PUBLICNAME"); h != "" {
			game.PublicName = h
		} else if h, err := os.Hostname(); err == nil {
			game.PublicName = h
		} else {
			game.PublicName = ""
		}
	}
}
