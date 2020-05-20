package lobby

import (
	crand "crypto/rand"
	"io/ioutil"
	"math"
	"math/big"
	"math/rand"
	"net/http"

	"github.com/golang/protobuf/proto"
	"github.com/gorilla/mux"

	"wsnet2/log"
	"wsnet2/pb"
)

func init() {
	seed, _ := crand.Int(crand.Reader, big.NewInt(math.MaxInt64))
	rand.Seed(seed.Int64())
}

func (rs *RoomService) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/rooms", rs.HandleCreateRoom).Methods("POST")
}

func (rs *RoomService) HandleCreateRoom(w http.ResponseWriter, r *http.Request) {
	appID := r.Header.Get("X-App-Id")
	userID := r.Header.Get("X-User-Id")

	_, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Errorf("Failed to read request body: %w", err)
		http.Error(w, "Failed to request body", http.StatusInternalServerError)
		return
	}
	r.Body.Close()

	// TODO: Unmarshal body and set RoomOption

	roomOption := &pb.RoomOption{
		Visible:   true,
		Watchable: true,
		LogLevel:  4,
	}
	clientInfo := &pb.ClientInfo{
		Id: userID,
	}
	room, err := rs.Create(appID, roomOption, clientInfo)
	if err != nil {
		log.Errorf("Failed to create room: %w", err)
		http.Error(w, "Failed to create room", http.StatusInternalServerError)
		return
	}
	log.Debugf("%v", room)
	res, err := proto.Marshal(room)
	if err != nil {
		log.Errorf("Failed to marshal room: %w", err)
		http.Error(w, "Failed to marshal room", http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/protobuf; charset=utf-8")
	w.Write(res)
}
