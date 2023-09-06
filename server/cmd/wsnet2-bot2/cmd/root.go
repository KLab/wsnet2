package cmd

import (
	"crypto/tls"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var (
	lobbyURL string
	appId    string
	appKey   string

	proxyURL      string
	skipTLSVerify bool
	timeout       time.Duration

	msgBody = make([]byte, 5000)
	cli     *http.Client
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "wsnet2-bot",
	Short: "wsnet2 testing bot",
	Long:  `wsnet2 testing bot`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
	PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) {
		cli, err = newClient()
		return err
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.wsnet2-bot2.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.PersistentFlags().StringVarP(&lobbyURL, "lobby", "l", "http://localhost:8080", "Lobby URL")
	rootCmd.PersistentFlags().StringVarP(&appId, "app-id", "a", "testapp", "App ID")
	rootCmd.PersistentFlags().StringVarP(&appKey, "app-key", "k", "testapppkey", "App key")
	rootCmd.PersistentFlags().StringVarP(&proxyURL, "proxy", "p", "", "Proxy URL")
	rootCmd.PersistentFlags().BoolVarP(&skipTLSVerify, "skip-tls-verify", "s", false, "Skip TLS verify")
	rootCmd.PersistentFlags().DurationVarP(&timeout, "timeout", "t", 10*time.Second, "Timeout")
}

// newClient returns http.Client with proxy and TLS settings
func newClient() (*http.Client, error) {
	if !skipTLSVerify && proxyURL == "" {
		return &http.Client{
			Timeout: timeout,
		}, nil
	}

	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
	}
	if skipTLSVerify {
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	if proxyURL != "" {
		purl, err := url.Parse(proxyURL)
		if err != nil {
			return nil, err
		}
		tr.Proxy = http.ProxyURL(purl)
	}
	return &http.Client{
		Transport: tr,
		Timeout:   timeout,
	}, nil
}
