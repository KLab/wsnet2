package common

import (
	crand "crypto/rand"
	"math"
	"math/big"
	"math/rand"
)

// プロジェクトで一度だけ実行する.
func init() {
	seed, _ := crand.Int(crand.Reader, big.NewInt(math.MaxInt64))
	rand.Seed(seed.Int64())
}
