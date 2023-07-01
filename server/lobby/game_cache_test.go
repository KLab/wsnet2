package lobby

import (
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var (
	lobbyDB *sqlx.DB
)

func TestHubCache(t *testing.T) {
	if lobbyDB == nil {
		t.Skip("require database")
	}

	lobbyDB.MustExec("DROP TABLE IF EXISTS `hub_server`")
	// TODO: 10-schema.sql から指定したテーブルの定義を読み込んで実行するような仕組みが欲しい
	lobbyDB.MustExec(
		"CREATE TABLE `hub_server` (\n" +
			"  `id`          INTEGER UNSIGNED NOT NULL PRIMARY KEY AUTO_INCREMENT,\n" +
			"  `hostname`    VARCHAR(191) NOT NULL,\n" +
			"  `public_name` VARCHAR(191) NOT NULL,\n" +
			"  `grpc_port`   INTEGER NOT NULL,\n" +
			"  `ws_port`     INTEGER NOT NULL,\n" +
			"  `status`      TINYINT NOT NULL,\n" +
			"  `heartbeat`   BIGINT,\n" +
			"  UNIQUE KEY `idx_hostname` (`hostname`)\n" +
			") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4")

	now := time.Now()
	nowUnix := now.Unix()
	lobbyDB.MustExec(
		`INSERT INTO hub_server (id, hostname, public_name, grpc_port, ws_port, status, heartbeat) VALUES
		(1, "host1", "global1", 1001, 1002, 0, ?),
		(2, "host2", "global2", 2001, 2002, 1, ?),
		(3, "host3", "global3", 3001, 3002, 2, ?),
		(4, "host4", "global4", 4001, 4002, 1, ?)`,
		nowUnix, nowUnix, nowUnix, nowUnix-100)
	// host1 - not ready
	// host2 - ready
	// host3 - shutting down
	// host4 - expired
	// host2のみが選択される

	hc := newHubCache(lobbyDB, time.Second, time.Second*10)
	err := hc.update()
	if err != nil {
		t.Fatal(err)
	}
	if hc.lastUpdated.Before(now) {
		t.Errorf("lastUpdated is not updated: now=%v lastUpdated=%v", now, hc.lastUpdated)
	}
	if len(hc.servers) != 1 {
		t.Errorf("len(servers) is not 1: %v", hc.servers)
	}
	if len(hc.order) != 1 {
		t.Errorf("len(order) is not 1: %v", hc.order)
	}
	host, err := hc.Rand()
	if err != nil {
		t.Fatalf("hc.Rand(): %v", err)
	}
	if host == nil {
		t.Fatalf("host is nil")
	}
	if host.Id != 2 {
		t.Errorf("host.Id is not 2: %v", host.Id)
	}
	host2, err := hc.Get(host.Id)
	if err != nil {
		t.Fatalf("hc.Get(%v): %v", host.Id, err)
	}
	if host != host2 {
		t.Errorf("host != host2: %+v != %+v", host, host2)
	}
}

func TestMain(m *testing.M) {
	var err error
	// ローカルで実行するときは次のようにしてDBを実行する
	// docker run -e MYSQL_ALLOW_EMPTY_PASSWORD=yes -p 3306:3306 --rm --name mysql mysql:8.0
	// 他のパッケージのテストが並列してDBを使ってもいいように、 lobby パッケージ専用のDBを作って実行している。
	//
	// NOTE:
	// - テスト実行用のcompose.yamlも用意した方がいいか？
	//   - その場合はGitHub Actionでも docker compose を使ってDB起動して共通化できる
	// - dockertest か testcontainers を使うか?
	//   - 別パッケージを並列にテスト実行することを考えたら、コンテナは論理DB作に比べてオーバーヘッドが大きい
	lobbyDB, err = sqlx.Connect("mysql", "root@tcp(127.0.0.1:3306)/")
	if err == nil {
		lobbyDB.MustExec("DROP DATABASE IF EXISTS wsnet2_test_lobby")
		lobbyDB.MustExec("CREATE DATABASE wsnet2_test_lobby")
		lobbyDB.Close()
		lobbyDB = sqlx.MustConnect("mysql", "root@tcp(127.0.0.1:3306)/wsnet2_test_lobby")
	} else {
		fmt.Printf("### failed to connect mysql: %v\n", err)
		// CI環境等でDBを使ったテストがスキップされても気づかない状況を避けるために、
		// 環境変数を使ってスキップではなく失敗扱いにする。
		if os.Getenv("WSNET2_FORCE_DB_TEST") == "" {
			fmt.Println("  runs only tests that do not require mysql")
		} else {
			fmt.Println("  failed because WSNET2_FORCE_DB_TEST is set")
			os.Exit(1)
		}
	}

	// 上の情報とテスト実行の出力の間に空行を挟んで見やすくする
	fmt.Println()
	os.Exit(m.Run())
}
