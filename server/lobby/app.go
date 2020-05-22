package lobby

import (
	"github.com/jmoiron/sqlx"

	"wsnet2/log"
	"wsnet2/pb"
)

const appQueryString = "SELECT id, `key` FROM app"

func scanApp(rows *sqlx.Rows) (interface{}, error) {
	defer rows.Close()
	apps := make([]pb.App, 0)
	for rows.Next() {
		app := pb.App{}
		err := rows.StructScan(&app)
		if err != nil {
			log.Errorf("Failed to scan app: %v", err)
			continue
		}
		apps = append(apps, app)
	}
	return apps, nil
}
