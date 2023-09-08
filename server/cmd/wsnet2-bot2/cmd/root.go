package cmd

import (
	"context"
	"crypto/tls"
	"errors"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"wsnet2/binary"
	"wsnet2/client"
	"wsnet2/pb"
)

var (
	lobbyURL string
	appId    string
	appKey   string

	proxyURL      string
	skipTLSVerify bool
	timeout       time.Duration

	verbose bool

	msgBody = make([]byte, 5000)
	logger  *zap.SugaredLogger
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
		return errors.Join(
			setupLogger(),
			setupClient(),
		)
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		logger.Sync()
	},
	SilenceUsage: true,
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
	rootCmd.PersistentFlags().DurationVarP(&timeout, "timeout", "t", 5*time.Second, "Lobby request timeout")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose log output")
}

func setupLogger() error {
	cfg := zap.NewDevelopmentConfig()
	if verbose {
		cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	} else {
		cfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}
	lg, err := cfg.Build()
	if err != nil {
		return err
	}
	logger = lg.Sugar()
	return nil
}

func setupClient() error {
	client.LobbyTimeout = timeout

	if skipTLSVerify || proxyURL == "" {
		tr := &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		}
		if skipTLSVerify {
			tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		}
		if proxyURL != "" {
			purl, err := url.Parse(proxyURL)
			if err != nil {
				return err
			}
			tr.Proxy = http.ProxyURL(purl)
		}

		client.LobbyTransport = tr
	}

	return nil
}

// createRoom creates room
func createRoom(ctx context.Context, owner string, group uint32, pubprops binary.Dict) (*client.Room, *client.Connection, error) {
	if pubprops == nil {
		pubprops = make(binary.Dict)
	}

	accinfo, err := client.GenAccessInfo(lobbyURL, appId, appKey, owner)
	if err != nil {
		return nil, nil, err
	}

	roomopt := &pb.RoomOption{
		Visible:     true,
		Joinable:    true,
		Watchable:   true,
		SearchGroup: group,
		PublicProps: binary.MarshalDict(pubprops),
	}

	cinfo := &pb.ClientInfo{Id: owner}

	return client.Create(ctx, accinfo, roomopt, cinfo, nil)
}

// joinRandom joins the player to a room randomly
func joinRandom(ctx context.Context, player string, group uint32, query *client.Query) (*client.Room, *client.Connection, error) {
	accinfo, err := client.GenAccessInfo(lobbyURL, appId, appKey, player)
	if err != nil {
		return nil, nil, err
	}

	cinfo := &pb.ClientInfo{Id: player}

	return client.RandomJoin(ctx, accinfo, group, query, cinfo, nil)
}

// watchRoom joins the watcher to the room
func watchRoom(ctx context.Context, watcher string, roomId string) (*client.Room, *client.Connection, error) {
	accinfo, err := client.GenAccessInfo(lobbyURL, appId, appKey, watcher)
	if err != nil {
		return nil, nil, err
	}

	return client.Watch(ctx, accinfo, roomId, nil, nil)
}
