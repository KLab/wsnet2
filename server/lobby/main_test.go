package lobby

import (
	"fmt"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var (
	lobbyDB *sqlx.DB
)

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
