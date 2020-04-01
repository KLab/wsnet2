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
}
