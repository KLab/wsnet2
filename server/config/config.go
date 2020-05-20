package config

import (
	"fmt"
	"io/ioutil"
	"path"
	"strings"

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
	Hostname string

	GRPCAddr      string `toml:"grpc_addr"`
	WebsocketAddr string `toml:"websocket_addr"`
	PprofAddr     string `toml:"pprof_addr"`

	TLSCert string `toml:"tls_cert"`
	TLSKey  string `toml:"tls_key"`

	RetryCount int `toml:"retry_count"`
	MaxRoomNum int `toml:"max_room_num"`

	DefaultMaxPlayers uint32 `toml:"default_max_players"`
	DefaultDeadline   uint32 `toml:"default_deadline"`
	DefaultLoglevel   uint32 `toml:"default_loglevel"`
}

type LobbyConf struct {
	Hostname  string
	Net       string
	Addr      string
	PprofAddr string `toml:"pprof_addr"`
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
		},
		Lobby: LobbyConf{},
	}

	_, err := toml.DecodeFile(conffile, c)
	if err != nil {
		return nil, err
	}

	err = c.Db.loadAuthfile(conffile)
	if err != nil {
		return nil, err
	}

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
	return fmt.Sprintf("%s@tcp(%s:%d)/%s", user, db.Host, db.Port, db.DBName)
}
