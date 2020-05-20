package lobby

import (
	crand "crypto/rand"
	"math"
	"math/big"
	"math/rand"
	"net/http"

	"github.com/gorilla/mux"
)

func init() {
	seed, _ := crand.Int(crand.Reader, big.NewInt(math.MaxInt64))
	rand.Seed(seed.Int64())
}

func (rs *RoomService) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/rooms", rs.HandleCreateRoom).Methods("POST")
}

func (rs *RoomService) HandleCreateRoom(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("Create Room\n"))
}
