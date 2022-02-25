/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"github.com/spf13/cobra"
)

type app struct {
	Id   string `db:"id"`
	Name string `db:"name"`
	Key  string `db:"key"`
}

// appsCmd represents the apps command
var appsCmd = &cobra.Command{
	Use:   "apps",
	Short: "Show applications",
	Long:  "Show applications registered on the DB",
	Run: func(cmd *cobra.Command, args []string) {
		const sql = "SELECT `id`, `key`, `name` FROM `app`"

		var apps []*app
		err := db.SelectContext(cmd.Context(), &apps, sql)
		if err != nil {
			panic(err)
		}

		if verbose {
			cmd.Println("id\tkey\tname")
		}

		for _, app := range apps {
			cmd.Printf("%s\t%s\t%q\n", app.Id, app.Key, app.Name)
		}
	},
}

func init() {
	rootCmd.AddCommand(appsCmd)
}
