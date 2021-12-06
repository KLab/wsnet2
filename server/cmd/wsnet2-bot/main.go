package main

import (
	crand "crypto/rand"
	"flag"
	"fmt"
	"math"
	"math/big"
	"math/rand"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	appID  = "testapp"
	appKey = "testapppkey"
)

var (
	WSNet2Version string = "LOCAL"
	WSNet2Commit  string = "LOCAL"

	logger *zap.SugaredLogger
)

type subcmd interface {
	Name() string
	Execute([]string)
}

var cmds = []subcmd{
	NewNormalBot(),
	NewStressBot(),
	NewStaticBot(),
}

var lobbyPrefix string = "http://192.168.0.1:3000"

func main() {
	verbose := flag.Bool("v", false, "verbose")
	flag.StringVar(&lobbyPrefix, "lobby", "http://localhost:8000", "lobby schema://host:port")
	flag.Parse()
	seed, _ := crand.Int(crand.Reader, big.NewInt(math.MaxInt64))
	rand.Seed(seed.Int64())

	cfg := zap.NewDevelopmentConfig()
	if !*verbose {
		cfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}
	cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("15:04:05.000")
	lg, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	defer lg.Sync()

	logger = lg.Sugar()

	fmt.Println("WSNet2-Bot")
	fmt.Println("WSNet2Version:", WSNet2Version)
	fmt.Println("WSNet2Commit:", WSNet2Commit)

	subcmd := "normal"
	args := flag.Args()
	if len(args) > 0 {
		subcmd = args[0]
		args = args[1:]
	}
	for _, cmd := range cmds {
		if cmd.Name() == subcmd {
			cmd.Execute(args)
			return
		}
	}
	logger.Errorf("command not found: %v", subcmd)
}
