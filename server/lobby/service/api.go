package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/vmihailenco/msgpack/v5"
	"golang.org/x/xerrors"

	"wsnet2/auth"
	"wsnet2/lobby"
	"wsnet2/log"
	"wsnet2/pb"
)

func msgpackDecode(r io.Reader, out interface{}) error {
	dec := msgpack.NewDecoder(r)
	dec.SetCustomStructTag("json")
	return dec.Decode(out)
}

func (sv *LobbyService) serveAPI(ctx context.Context) <-chan error {
	errCh := make(chan error)

	go func() {
		network := sv.conf.Net

		laddr := fmt.Sprintf(":%d", sv.conf.Port)
		if network == "unix" {
			laddr = sv.conf.UnixPath
		}

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
	r.HandleFunc("/rooms/watch/number/{roomNumber:[0-9]+}", sv.handleWatchRoomByNumber).Methods("POST")
}

type header struct {
	appId    string
	userId   string
	authData string
}

func parseSpecificHeader(r *http.Request) (hdr header) {
	hdr.appId = r.Header.Get("Wsnet2-App")
	hdr.userId = r.Header.Get("Wsnet2-User")

	bearer := r.Header.Get("Authorization")
	if strings.HasPrefix(bearer, "Bearer ") {
		hdr.authData = bearer[len("Bearer "):]
	}

	return hdr
}

func renderResponse(w http.ResponseWriter, res *LobbyResponse) {
	var body bytes.Buffer
	enc := msgpack.NewEncoder(&body)
	enc.SetCustomStructTag("json")
	enc.UseCompactInts(true)
	err := enc.Encode(res)
	if err != nil {
		log.Errorf("Failed to marshal response: %v", err)
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/x-msgpack")
	w.Write(body.Bytes())
}

func renderJoinedRoomResponse(w http.ResponseWriter, room *pb.JoinedRoomRes) {
	log.Debugf("joined room: %#v", room)
	renderResponse(w, &LobbyResponse{Msg: "OK", Room: room})
}

func renderSearchRoomResponse(w http.ResponseWriter, rooms []*pb.RoomInfo) {
	log.Debugf("search found rooms: %#v", rooms)
	renderResponse(w, &LobbyResponse{Msg: "OK", Rooms: rooms})
}

func renderErrorResponse(w http.ResponseWriter, msg string, status int, err error) {
	if ews, ok := err.(lobby.ErrorWithStatus); ok {
		status = ews.Status()
		if m := ews.Message(); m != "" {
			msg = m
		}
	}
	if status == http.StatusOK {
		log.Debugf("Failed with status OK: %+v", err)
		renderResponse(w, &LobbyResponse{Msg: msg})
		return
	}
	log.Errorf("ErrorResponse: %d %s: %+v", status, msg, err)
	http.Error(w, msg, status)
}

func (sv *LobbyService) authUser(h header) (string, error) {
	appKey, found := sv.roomService.GetAppKey(h.appId)
	if !found {
		return "", xerrors.Errorf("Invalid appId: %v", h.appId)
	}
	expired := time.Now().Add(-time.Duration(sv.conf.AuthDataExpire))
	if err := auth.ValidAuthData(h.authData, appKey, h.userId, expired); err != nil {
		return "", xerrors.Errorf("invalid authdata: %w", err)
	}
	return appKey, nil
}

// 部屋を作成する
// Method: POST
// Path: /rooms
// POST Params: {"max_player": 0, "with_room_number": true}
// Response: 200 OK
func (sv *LobbyService) handleCreateRoom(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(sv.conf.ApiTimeout))
	defer cancel()

	h := parseSpecificHeader(r)

	log.Infof("handleCreateRoom: appID=%s, userID=%s", h.appId, h.userId)

	appKey, err := sv.authUser(h)
	if err != nil {
		renderErrorResponse(w, "Failed to user auth", http.StatusUnauthorized, err)
		return
	}

	var param CreateParam
	if err := msgpackDecode(r.Body, &param); err != nil {
		renderErrorResponse(w, "Failed to read request body", http.StatusBadRequest, err)
		return
	}
	macKey, err := auth.DecryptMACKey(param.EncMACKey, appKey)
	if err != nil {
		renderErrorResponse(w, "Failed to read MAC Key", http.StatusBadRequest, err)
		return
	}

	// TODO: 必要に応じて一部のパラメータを上書き？

	room, err := sv.roomService.Create(ctx, h.appId, &param.RoomOption, &param.ClientInfo, macKey)
	if err != nil {
		renderErrorResponse(w, "Failed to create room", http.StatusInternalServerError, err)
		return
	}

	renderJoinedRoomResponse(w, room)
}

type JoinVars map[string]string

func (vars JoinVars) roomId() string {
	id := vars["roomId"]
	return id
}

func (vars JoinVars) roomNumber() (number int32) {
	if v, found := vars["roomNumber"]; found {
		n, _ := strconv.ParseInt(v, 10, 32)
		number = int32(n)
	}
	return number
}

func (vars JoinVars) searchGroup() (sg uint32) {
	if v, found := vars["searchGroup"]; found {
		n, _ := strconv.ParseUint(v, 10, 32)
		sg = uint32(n)
	}
	return sg
}

func (sv *LobbyService) handleJoinRoom(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(sv.conf.ApiTimeout))
	defer cancel()

	h := parseSpecificHeader(r)

	log.Infof("handleJoinRoom: appID=%s, userID=%s", h.appId, h.userId)

	appKey, err := sv.authUser(h)
	if err != nil {
		renderErrorResponse(w, "Failed to user auth", http.StatusUnauthorized, err)
		return
	}

	var param JoinParam
	err = msgpackDecode(r.Body, &param)
	if err != nil {
		renderErrorResponse(w, "Failed to read request body", http.StatusBadRequest, err)
		return
	}

	macKey, err := auth.DecryptMACKey(param.EncMACKey, appKey)
	if err != nil {
		renderErrorResponse(w, "Failed to read MAC Key", http.StatusBadRequest, err)
		return
	}

	vars := JoinVars(mux.Vars(r))
	roomId := vars.roomId()
	if roomId == "" {
		renderErrorResponse(
			w, "Invalid room id", http.StatusBadRequest, xerrors.Errorf("Invalid room id"))
		return
	}

	room, err := sv.roomService.JoinById(ctx, h.appId, roomId, param.Queries, &param.ClientInfo, macKey)
	if err != nil {
		renderErrorResponse(w, "Failed to join room", http.StatusInternalServerError, err)
		return
	}

	renderJoinedRoomResponse(w, room)
}

func (sv *LobbyService) handleJoinRoomByNumber(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(sv.conf.ApiTimeout))
	defer cancel()

	h := parseSpecificHeader(r)

	log.Infof("handleJoinRoomByNumber: appID=%s, userID=%s", h.appId, h.userId)

	appKey, err := sv.authUser(h)
	if err != nil {
		renderErrorResponse(w, "Failed to user auth", http.StatusUnauthorized, err)
		return
	}

	var param JoinParam
	err = msgpackDecode(r.Body, &param)
	if err != nil {
		renderErrorResponse(w, "Failed to read request body", http.StatusBadRequest, err)
		return
	}

	macKey, err := auth.DecryptMACKey(param.EncMACKey, appKey)
	if err != nil {
		renderErrorResponse(w, "Failed to read MAC Key", http.StatusBadRequest, err)
		return
	}

	vars := JoinVars(mux.Vars(r))
	roomNumber := vars.roomNumber()
	if roomNumber == 0 {
		renderErrorResponse(
			w, "Invalid room number", http.StatusBadRequest, xerrors.Errorf("Invalid room number: 0"))
		return
	}

	room, err := sv.roomService.JoinByNumber(ctx, h.appId, roomNumber, param.Queries, &param.ClientInfo, macKey)
	if err != nil {
		renderErrorResponse(w, "Failed to join room", http.StatusInternalServerError, err)
		return
	}

	renderJoinedRoomResponse(w, room)
}

func (sv *LobbyService) handleJoinRoomAtRandom(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(sv.conf.ApiTimeout))
	defer cancel()

	h := parseSpecificHeader(r)

	log.Infof("handleJoinRoomAtRandom: appID=%s, userID=%s", h.appId, h.userId)

	appKey, err := sv.authUser(h)
	if err != nil {
		renderErrorResponse(w, "Failed to user auth", http.StatusUnauthorized, err)
		return
	}

	var param JoinParam
	err = msgpackDecode(r.Body, &param)
	if err != nil {
		renderErrorResponse(w, "Failed to read request body", http.StatusBadRequest, err)
		return
	}

	macKey, err := auth.DecryptMACKey(param.EncMACKey, appKey)
	if err != nil {
		renderErrorResponse(w, "Failed to read MAC Key", http.StatusBadRequest, err)
		return
	}

	vars := JoinVars(mux.Vars(r))
	searchGroup := vars.searchGroup()

	room, err := sv.roomService.JoinAtRandom(ctx, h.appId, searchGroup, param.Queries, &param.ClientInfo, macKey)
	if err != nil {
		renderErrorResponse(w, "Failed to join room", http.StatusInternalServerError, err)
		return
	}

	renderJoinedRoomResponse(w, room)
}

func (sv *LobbyService) handleSearchRoom(w http.ResponseWriter, r *http.Request) {
	h := parseSpecificHeader(r)

	log.Infof("handleSearchRoom: appID=%s, userID=%s", h.appId, h.userId)

	if _, err := sv.authUser(h); err != nil {
		renderErrorResponse(w, "Failed to user auth", http.StatusUnauthorized, err)
		return
	}

	var param SearchParam
	err := msgpackDecode(r.Body, &param)
	if err != nil {
		renderErrorResponse(w, "Failed to read request body", http.StatusBadRequest, err)
		return
	}

	log.Debugf("%#v", param)

	rooms, err := sv.roomService.Search(
		h.appId, param.SearchGroup, param.Queries, int(param.Limit), param.CheckJoinable, param.CheckWatchable)
	if err != nil {
		renderErrorResponse(w, "Failed to search room", http.StatusInternalServerError, err)
		return
	}
	log.Debugf("%#v", rooms)

	renderSearchRoomResponse(w, rooms)
}

func (sv *LobbyService) handleWatchRoom(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(sv.conf.ApiTimeout))
	defer cancel()

	h := parseSpecificHeader(r)

	log.Infof("handleWatchRoom: appID=%s, userID=%s", h.appId, h.userId)

	appKey, err := sv.authUser(h)
	if err != nil {
		renderErrorResponse(w, "Failed to user auth", http.StatusUnauthorized, err)
		return
	}

	var param JoinParam
	err = msgpackDecode(r.Body, &param)
	if err != nil {
		renderErrorResponse(w, "Failed to read request body", http.StatusBadRequest, err)
		return
	}

	macKey, err := auth.DecryptMACKey(param.EncMACKey, appKey)
	if err != nil {
		renderErrorResponse(w, "Failed to read MAC Key", http.StatusBadRequest, err)
		return
	}

	vars := JoinVars(mux.Vars(r))
	roomId := vars.roomId()
	if roomId == "" {
		renderErrorResponse(
			w, "Invalid room id", http.StatusBadRequest, xerrors.Errorf("Invalid room id"))
		return
	}

	room, err := sv.roomService.WatchById(ctx, h.appId, roomId, param.Queries, &param.ClientInfo, macKey)
	if err != nil {
		renderErrorResponse(w, "Failed to watch room", http.StatusInternalServerError, err)
		return
	}
	log.Debugf("%#v", room)

	renderJoinedRoomResponse(w, room)
}

func (sv *LobbyService) handleWatchRoomByNumber(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(sv.conf.ApiTimeout))
	defer cancel()

	h := parseSpecificHeader(r)

	log.Infof("handleWatchRoomByNumber: appID=%s, userID=%s", h.appId, h.userId)

	appKey, err := sv.authUser(h)
	if err != nil {
		renderErrorResponse(w, "Failed to user auth", http.StatusUnauthorized, err)
		return
	}

	var param JoinParam
	err = msgpackDecode(r.Body, &param)
	if err != nil {
		renderErrorResponse(w, "Failed to read request body", http.StatusBadRequest, err)
		return
	}

	macKey, err := auth.DecryptMACKey(param.EncMACKey, appKey)
	if err != nil {
		renderErrorResponse(w, "Failed to read MAC Key", http.StatusBadRequest, err)
		return
	}

	vars := JoinVars(mux.Vars(r))
	roomNumber := vars.roomNumber()
	if roomNumber == 0 {
		renderErrorResponse(
			w, "Invalid room number", http.StatusBadRequest, xerrors.Errorf("Invalid room number: 0"))
		return
	}

	room, err := sv.roomService.WatchByNumber(ctx, h.appId, roomNumber, param.Queries, &param.ClientInfo, macKey)
	if err != nil {
		renderErrorResponse(w, "Failed to watch room", http.StatusInternalServerError, err)
		return
	}
	log.Debugf("%#v", room)

	renderJoinedRoomResponse(w, room)
}
