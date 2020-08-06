package service

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/vmihailenco/msgpack/v4"
	"golang.org/x/xerrors"

	"wsnet2/auth"
	"wsnet2/lobby"
	"wsnet2/log"
	"wsnet2/pb"
)

const expirationTime = 30

func (sv *LobbyService) serveAPI(ctx context.Context) <-chan error {
	errCh := make(chan error)

	go func() {
		network := sv.conf.Net
		laddr := fmt.Sprintf(":%d", sv.conf.Port)
		log.Infof("lobby api: %#v %#v", network, laddr)

		listener, err := net.Listen(network, laddr)
		if err != nil {
			errCh <- xerrors.Errorf("listen error: %w", err)
			return
		}

		if network == "unix" {
			fi, err := os.Stat(laddr)
			if err != nil {
				errCh <- xerrors.Errorf("stat error: %w", err)
				return
			}
			err = os.Chmod(laddr, fi.Mode()|0777)
			if err != nil {
				errCh <- xerrors.Errorf("chmod error: %w", err)
				return
			}
		}

		r := mux.NewRouter()
		sv.registerRoutes(r)

		errCh <- http.Serve(listener, r)
	}()

	return errCh
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("wsnet2 works\n"))
}

func (sv *LobbyService) registerRoutes(r *mux.Router) {
	r.HandleFunc("/health", handleHealth).Methods("GET")
	r.HandleFunc("/health/", handleHealth).Methods("GET")

	if sv.conf.Hostname != "" {
		r = r.Host(sv.conf.Hostname).Subrouter()
	}

	r.HandleFunc("/rooms", sv.handleCreateRoom).Methods("POST")
	r.HandleFunc("/rooms/join/id/{roomId}", sv.handleJoinRoom).Methods("POST")
	r.HandleFunc("/rooms/join/number/{roomNumber:[0-9]+}", sv.handleJoinRoomByNumber).Methods("POST")
	r.HandleFunc("/rooms/join/random/{searchGroup:[0-9]+}", sv.handleJoinRoomAtRandom).Methods("POST")
	r.HandleFunc("/rooms/search", sv.handleSearchRoom).Methods("POST")
	r.HandleFunc("/rooms/watch/id/{roomId}", sv.handleWatchRoom).Methods("POST")
}

type header struct {
	appId     string
	userId    string
	timestamp string
	nonce     string
	hash      string
}

func parseSpecificHeader(r *http.Request) *header {
	return &header{
		appId:     r.Header.Get("X-Wsnet-App"),
		userId:    r.Header.Get("X-Wsnet-User"),
		timestamp: r.Header.Get("X-Wsnet-Timestamp"),
		nonce:     r.Header.Get("X-Wsnet-Nonce"),
		hash:      r.Header.Get("X-Wsnet-Hash"),
	}
}

func renderResponse(w http.ResponseWriter, res interface{}) error {
	var body bytes.Buffer
	err := msgpack.NewEncoder(&body).UseJSONTag(true).Encode(res)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/x-msgpack")
	w.Write(body.Bytes())
	return nil
}

func (sv *LobbyService) authUser(h *header) error {
	appKey, found := sv.roomService.GetAppKey(h.appId)
	if !found {
		return xerrors.Errorf("Invalid appId: %v", h.appId)
	}
	ts, err := strconv.ParseInt(h.timestamp, 10, 64)
	if err != nil {
		return xerrors.Errorf("Invalid timestamp: %w", err)
	}
	now := time.Now().Unix()
	if now < ts {
		return xerrors.Errorf("Invalid timestamp: now=%v, ts=%v", now, ts)
	}
	// TODO: expirationTimeはコンフィグに定義？
	if now-ts > expirationTime {
		return xerrors.Errorf("Expired timestamp: now=%v, ts=%v, expirationTime=%v", now, ts, expirationTime)
	}
	if !auth.ValidHexHMAC(h.hash, []byte(appKey), h.userId, h.timestamp, h.nonce) {
		return xerrors.Errorf("Invalid HMAC: appId=%v, userId=%v, timestamp=%v, nonce=%v, hash=%v", h.appId, h.userId, h.timestamp, h.nonce, h.hash)
	}
	return nil
}

type CreateParam struct {
	RoomOption pb.RoomOption
	ClientInfo pb.ClientInfo
}

// 部屋を作成する
// Method: POST
// Path: /rooms
// POST Params: {"max_player": 0, "with_room_number": true}
// Response: 200 OK
func (sv *LobbyService) handleCreateRoom(w http.ResponseWriter, r *http.Request) {
	h := parseSpecificHeader(r)

	log.Infof("handleCreateRoom: appID=%s, userID=%s", h.appId, h.userId)

	if err := sv.authUser(h); err != nil {
		log.Errorf("Failed to user auth: %v", err)
		http.Error(w, "Failed to user auth", http.StatusUnauthorized)
		return
	}

	var param CreateParam
	err := msgpack.NewDecoder(r.Body).UseJSONTag(true).Decode(&param)
	if err != nil {
		log.Errorf("Failed to read request body: %v", err)
		http.Error(w, "Failed to request body", http.StatusInternalServerError)
		return
	}

	// TODO: 必要に応じて一部のパラメータを上書き？

	room, err := sv.roomService.Create(h.appId, &param.RoomOption, &param.ClientInfo)
	if err != nil {
		log.Errorf("Failed to create room: %v", err)
		http.Error(w, "Failed to create room", http.StatusInternalServerError)
		return
	}
	log.Debugf("%#v", room)

	err = renderResponse(w, room)
	if err != nil {
		log.Errorf("Failed to marshal room: %v", err)
		http.Error(w, "Failed to marshal room", http.StatusInternalServerError)
		return
	}
}

type JoinParam struct {
	Queries    []lobby.PropQueries
	ClientInfo pb.ClientInfo
}

type JoinVars map[string]string

func (vars JoinVars) roomId() (string, bool) {
	id, found := vars["roomId"]
	return id, found
}

func (vars JoinVars) roomNumber() (number int32, found bool) {
	if v, found := vars["roomNumber"]; found {
		n, _ := strconv.ParseInt(v, 10, 32)
		number = int32(n)
	}
	return number, found
}

func (vars JoinVars) searchGroup() (sg uint32, found bool) {
	if v, found := vars["searchGroup"]; found {
		n, _ := strconv.ParseUint(v, 10, 32)
		sg = uint32(n)
	}
	return sg, found
}

func (sv *LobbyService) handleJoinRoom(w http.ResponseWriter, r *http.Request) {
	h := parseSpecificHeader(r)

	log.Infof("handleJoinRoom: appID=%s, userID=%s", h.appId, h.userId)

	if err := sv.authUser(h); err != nil {
		log.Errorf("Failed to user auth: %v", err)
		http.Error(w, "Failed to user auth", http.StatusUnauthorized)
		return
	}

	var param JoinParam
	err := msgpack.NewDecoder(r.Body).UseJSONTag(true).Decode(&param)
	if err != nil {
		log.Errorf("Failed to read request body: %v", err)
		http.Error(w, "Failed to request body", http.StatusInternalServerError)
		return
	}

	vars := JoinVars(mux.Vars(r))
	roomId, _ := vars.roomId()

	room, err := sv.roomService.JoinById(h.appId, roomId, param.Queries, &param.ClientInfo)
	if err != nil {
		log.Errorf("Failed to join room: %v", err)
		http.Error(w, "Failed to join room", http.StatusInternalServerError)
		return
	}
	log.Debugf("%#v", room)

	err = renderResponse(w, room)
	if err != nil {
		log.Errorf("Failed to marshal room: %v", err)
		http.Error(w, "Failed to marshal room", http.StatusInternalServerError)
		return
	}
}

func (sv *LobbyService) handleJoinRoomByNumber(w http.ResponseWriter, r *http.Request) {
	h := parseSpecificHeader(r)

	log.Infof("handleJoinRoomByNumber: appID=%s, userID=%s", h.appId, h.userId)

	if err := sv.authUser(h); err != nil {
		log.Errorf("Failed to user auth: %v", err)
		http.Error(w, "Failed to user auth", http.StatusUnauthorized)
		return
	}

	var param JoinParam
	err := msgpack.NewDecoder(r.Body).UseJSONTag(true).Decode(&param)
	if err != nil {
		log.Errorf("Failed to read request body: %v", err)
		http.Error(w, "Failed to request body", http.StatusInternalServerError)
		return
	}

	vars := JoinVars(mux.Vars(r))
	roomNumber, _ := vars.roomNumber()

	room, err := sv.roomService.JoinByNumber(h.appId, roomNumber, param.Queries, &param.ClientInfo)
	if err != nil {
		log.Errorf("Failed to join room: %v", err)
		http.Error(w, "Failed to join room", http.StatusInternalServerError)
		return
	}
	log.Debugf("%#v", room)

	err = renderResponse(w, room)
	if err != nil {
		log.Errorf("Failed to marshal room: %v", err)
		http.Error(w, "Failed to marshal room", http.StatusInternalServerError)
		return
	}
}

func (sv *LobbyService) handleJoinRoomAtRandom(w http.ResponseWriter, r *http.Request) {
	h := parseSpecificHeader(r)

	log.Infof("handleJoinRoomAtRandom: appID=%s, userID=%s", h.appId, h.userId)

	if err := sv.authUser(h); err != nil {
		log.Errorf("Failed to user auth: %v", err)
		http.Error(w, "Failed to user auth", http.StatusUnauthorized)
		return
	}

	var param JoinParam
	err := msgpack.NewDecoder(r.Body).UseJSONTag(true).Decode(&param)
	if err != nil {
		log.Errorf("Failed to read request body: %v", err)
		http.Error(w, "Failed to request body", http.StatusInternalServerError)
		return
	}

	vars := JoinVars(mux.Vars(r))
	searchGroup, _ := vars.searchGroup()

	room, err := sv.roomService.JoinAtRandom(h.appId, searchGroup, param.Queries, &param.ClientInfo)
	if err != nil {
		log.Errorf("Failed to join room: %v", err)
		http.Error(w, "Failed to join room", http.StatusInternalServerError)
		return
	}
	log.Debugf("%#v", room)

	err = renderResponse(w, room)
	if err != nil {
		log.Errorf("Failed to marshal room: %v", err)
		http.Error(w, "Failed to marshal room", http.StatusInternalServerError)
		return
	}
}

type SearchParam struct {
	SearchGroup uint32
	Queries     []lobby.PropQueries
	Limit       uint32
}

func (sv *LobbyService) handleSearchRoom(w http.ResponseWriter, r *http.Request) {
	h := parseSpecificHeader(r)

	log.Infof("handleSearchRoom: appID=%s, userID=%s", h.appId, h.userId)

	if err := sv.authUser(h); err != nil {
		log.Errorf("Failed to user auth: %v", err)
		http.Error(w, "Failed to user auth", http.StatusUnauthorized)
		return
	}

	var param SearchParam
	err := msgpack.NewDecoder(r.Body).UseJSONTag(true).Decode(&param)
	if err != nil {
		log.Errorf("Failed to read request body: %v", err)
		http.Error(w, "Failed to request body", http.StatusInternalServerError)
		return
	}

	log.Debugf("%#v", param)

	rooms, err := sv.roomService.Search(h.appId, param.SearchGroup, param.Queries, int(param.Limit))
	if err != nil {
		log.Errorf("Failed to search room: %v", err)
		http.Error(w, "Failed to search room", http.StatusInternalServerError)
		return
	}
	log.Debugf("%#v", rooms)

	err = renderResponse(w, rooms)
	if err != nil {
		log.Errorf("Failed to marshal room: %v", err)
		http.Error(w, "Failed to marshal room", http.StatusInternalServerError)
		return
	}
}

func (sv *LobbyService) handleWatchRoom(w http.ResponseWriter, r *http.Request) {
	h := parseSpecificHeader(r)

	log.Infof("handleJoinRoom: appID=%s, userID=%s", h.appId, h.userId)

	if err := sv.authUser(h); err != nil {
		log.Errorf("Failed to user auth: %v", err)
		http.Error(w, "Failed to user auth", http.StatusUnauthorized)
		return
	}

	roomId := mux.Vars(r)["roomId"]
	log.Infof("watch roomid=%v", roomId)

	var param JoinParam
	err := msgpack.NewDecoder(r.Body).UseJSONTag(true).Decode(&param)
	if err != nil {
		log.Errorf("Failed to read request body: %v", err)
		http.Error(w, "Failed to request body", http.StatusInternalServerError)
		return
	}

	room, err := sv.roomService.WatchById(h.appId, roomId, &param.ClientInfo)
	if err != nil {
		log.Errorf("Failed to join room: %v", err)
		http.Error(w, "Failed to join room", http.StatusInternalServerError)
		return
	}
	log.Debugf("%#v", room)

	err = renderResponse(w, room)
	if err != nil {
		log.Errorf("Failed to marshal room: %v", err)
		http.Error(w, "Failed to marshal room", http.StatusInternalServerError)
		return
	}
}