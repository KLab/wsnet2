package cmd

import (
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/cobra"

	"wsnet2"
	"wsnet2/config"
)

var (
	confFile string
	conf     *config.Config
	db       *sqlx.DB
	verbose  bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "wsnet2-tool",
	Short: "WSNet2 tool",
	Long:  "CLI tool for WSNet2 " + wsnet2.Version,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if confFile == "" {
			return fmt.Errorf("need --config option\n")
		}

		var err error
		conf, err = config.Load(confFile)
		if err != nil {
			return err
		}
		db, err = sqlx.Open("mysql", conf.Db.DSN())
		if err != nil {
			return err
		}
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v", err)
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVarP(&confFile, "config", "f", "", "Config toml file")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	_ = rootCmd.MarkPersistentFlagRequired("config")
}
