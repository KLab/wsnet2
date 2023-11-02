package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
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

		r := chi.NewMux()
		sv.registerRoutes(r)

		errCh <- http.Serve(listener, r)
	}()

	return errCh
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("wsnet2 works\n"))
}

func (sv *LobbyService) registerRoutes(r chi.Router) {
	r.Get("/health", handleHealth)
	r.Get("/health/", handleHealth)

	r.Post("/rooms", sv.handleCreateRoom)
	r.Post("/rooms/join/id/{roomId}", sv.handleJoinRoom)
	r.Post("/rooms/join/number/{roomNumber:[0-9]+}", sv.handleJoinRoomByNumber)
	r.Post("/rooms/join/random/{searchGroup:[0-9]+}", sv.handleJoinRoomAtRandom)
	r.Post("/rooms/search", sv.handleSearchRooms)
	r.Post("/rooms/search/ids", sv.handleSearchByIds)
	r.Post("/rooms/search/numbers", sv.handleSearchByNumbers)
	r.Post("/rooms/search/current", sv.handleSearchCurrentRooms)
	r.Post("/rooms/watch/id/{roomId}", sv.handleWatchRoom)
	r.Post("/rooms/watch/number/{roomNumber:[0-9]+}", sv.handleWatchRoomByNumber)
	r.Post("/_admin/kick", sv.handleAdminKick)
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

func prepareLogger(handler string, hdr header, r *http.Request) log.Logger {
	raddr, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		raddr = r.RemoteAddr
	}
	if f := r.Header.Get("x-forwarded-for"); f != "" {
		if raddr != "127.0.0.1" && raddr != "::1" {
			f += ", " + raddr
		}
		raddr = f
	}
	l := log.GetLoggerWith(
		log.KeyHandler, handler,
		log.KeyRequestedAt, float64(time.Now().UnixMilli())/1000,
		log.KeyApp, hdr.appId,
		log.KeyClient, hdr.userId,
		log.KeyRemoteAddr, raddr)
	if err != nil {
		l.Errorf("SplitHostPort: %v", err)
	}
	return l
}

func renderResponse(w http.ResponseWriter, res *lobby.Response, logger log.Logger) {
	var body bytes.Buffer
	enc := msgpack.NewEncoder(&body)
	enc.SetCustomStructTag("json")
	enc.UseCompactInts(true)
	err := enc.Encode(res)
	if err != nil {
		logger.Errorf("Failed to marshal response: %+v", err)
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}
	logger.Infof("Response(%v): %v", res.Type, res.Msg)
	w.Header().Set("Content-Type", "application/x-msgpack")
	w.Write(body.Bytes())
}

func renderJoinedRoomResponse(w http.ResponseWriter, room *pb.JoinedRoomRes, logger log.Logger) {
	logger = logger.With(log.KeyRoom, room.RoomInfo.Id)
	logger.Debugf("joined room: %v", room)
	renderResponse(w, &lobby.Response{Msg: "OK", Room: room}, logger)
}

func renderFoundRoomsResponse(w http.ResponseWriter, rooms []*pb.RoomInfo, logger log.Logger) {
	logger = logger.With(log.KeyRoomCount, len(rooms))
	logger.Debugf("found rooms: %v", rooms)
	t := lobby.ResponseTypeOK
	if len(rooms) == 0 {
		t = lobby.ResponseTypeNoRoomFound
	}
	renderResponse(w, &lobby.Response{Msg: "OK", Type: t, Rooms: rooms}, logger)
}

func renderErrorResponse(w http.ResponseWriter, msg string, status int, err error, logger log.Logger) {
	logmsg := msg
	if e, ok := err.(lobby.ErrorWithType); ok {
		if m := e.Message(); m != "" {
			msg = m
		}
		switch e.ErrType() {
		case lobby.ErrArgument:
			status = http.StatusBadRequest
		case lobby.ErrRoomLimit:
			logger.Infof("Failed with status OK: %+v", err)
			renderResponse(w, &lobby.Response{Msg: msg, Type: lobby.ResponseTypeRoomLimit}, logger)
			return
		case lobby.ErrAlreadyJoined:
			status = http.StatusConflict
		case lobby.ErrRoomFull:
			logger.Infof("Failed with status OK: %+v", err)
			renderResponse(w, &lobby.Response{Msg: msg, Type: lobby.ResponseTypeRoomFull}, logger)
			return
		case lobby.ErrNoJoinableRoom, lobby.ErrNoWatchableRoom:
			logger.Infof("Failed with status OK: %+v", err)
			renderResponse(w, &lobby.Response{Msg: msg, Type: lobby.ResponseTypeNoRoomFound}, logger)
			return
		}
	}
	logger.Errorf("ErrorResponse: %d %s: %+v", status, logmsg, err)
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
	logger := prepareLogger("lobby:create", h, r)
	logger.Debugf("handleCreateRoom")

	appKey, err := sv.authUser(h)
	if err != nil {
		renderErrorResponse(w, "Failed to user auth", http.StatusUnauthorized, err, logger)
		return
	}

	var param lobby.CreateParam
	if err := msgpackDecode(r.Body, &param); err != nil {
		renderErrorResponse(w, "Failed to read request body", http.StatusBadRequest, err, logger)
		return
	}
	macKey, err := auth.DecryptMACKey(appKey, param.EncMACKey)
	if err != nil {
		renderErrorResponse(w, "Failed to read MAC Key", http.StatusBadRequest, err, logger)
		return
	}

	room, err := sv.roomService.Create(ctx, h.appId, param.RoomOption, param.ClientInfo, macKey)
	if err != nil {
		renderErrorResponse(w, "Failed to create room", http.StatusInternalServerError, err, logger)
		return
	}

	renderJoinedRoomResponse(w, room, logger)
}

type JoinVars struct {
	ctx *chi.Context
}

func NewJoinVars(r *http.Request) *JoinVars {
	return &JoinVars{
		ctx: chi.RouteContext(r.Context()),
	}
}

func (vars JoinVars) roomId() string {
	id := vars.ctx.URLParam("roomId")
	return id
}

func (vars JoinVars) roomNumber() (number int32) {
	v := vars.ctx.URLParam("roomNumber")
	if v != "" {
		n, _ := strconv.ParseInt(v, 10, 32)
		number = int32(n)
	}
	return number
}

func (vars JoinVars) searchGroup() (sg uint32) {
	v := vars.ctx.URLParam("searchGroup")
	if v != "" {
		n, _ := strconv.ParseInt(v, 10, 32)
		sg = uint32(n)
	}
	return sg
}

func (sv *LobbyService) handleJoinRoom(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(sv.conf.ApiTimeout))
	defer cancel()

	h := parseSpecificHeader(r)
	logger := prepareLogger("lobby:join/id", h, r)
	logger.Debugf("handleJoinRoom")

	appKey, err := sv.authUser(h)
	if err != nil {
		renderErrorResponse(w, "Failed to user auth", http.StatusUnauthorized, err, logger)
		return
	}

	var param lobby.JoinParam
	err = msgpackDecode(r.Body, &param)
	if err != nil {
		renderErrorResponse(w, "Failed to read request body", http.StatusBadRequest, err, logger)
		return
	}

	macKey, err := auth.DecryptMACKey(appKey, param.EncMACKey)
	if err != nil {
		renderErrorResponse(w, "Failed to read MAC Key", http.StatusBadRequest, err, logger)
		return
	}

	vars := NewJoinVars(r)
	roomId := vars.roomId()
	if roomId == "" {
		renderErrorResponse(
			w, "Invalid room id", http.StatusBadRequest, xerrors.Errorf("Invalid room id"), logger)
		return
	}
	logger = logger.With(log.KeyRoom, roomId)

	room, err := sv.roomService.JoinById(ctx, h.appId, roomId, param.Queries, param.ClientInfo, macKey, logger)
	if err != nil {
		renderErrorResponse(w, "Failed to join room", http.StatusInternalServerError, err, logger)
		return
	}

	renderJoinedRoomResponse(w, room, logger)
}

func (sv *LobbyService) handleJoinRoomByNumber(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(sv.conf.ApiTimeout))
	defer cancel()

	h := parseSpecificHeader(r)
	logger := prepareLogger("lobby:join/number", h, r)
	logger.Debugf("handleJoinRoomByNumber")

	appKey, err := sv.authUser(h)
	if err != nil {
		renderErrorResponse(w, "Failed to user auth", http.StatusUnauthorized, err, logger)
		return
	}

	var param lobby.JoinParam
	err = msgpackDecode(r.Body, &param)
	if err != nil {
		renderErrorResponse(w, "Failed to read request body", http.StatusBadRequest, err, logger)
		return
	}

	macKey, err := auth.DecryptMACKey(appKey, param.EncMACKey)
	if err != nil {
		renderErrorResponse(w, "Failed to read MAC Key", http.StatusBadRequest, err, logger)
		return
	}

	vars := NewJoinVars(r)
	roomNumber := vars.roomNumber()
	if roomNumber == 0 {
		renderErrorResponse(
			w, "Invalid room number", http.StatusBadRequest, xerrors.Errorf("Invalid room number: 0"), logger)
		return
	}
	logger = logger.With(log.KeyRoomNumber, roomNumber)

	room, err := sv.roomService.JoinByNumber(ctx, h.appId, roomNumber, param.Queries, param.ClientInfo, macKey, logger)
	if err != nil {
		renderErrorResponse(w, "Failed to join room", http.StatusInternalServerError, err, logger)
		return
	}

	renderJoinedRoomResponse(w, room, logger)
}

func (sv *LobbyService) handleJoinRoomAtRandom(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(sv.conf.ApiTimeout))
	defer cancel()

	h := parseSpecificHeader(r)
	logger := prepareLogger("lobby:join/random", h, r)
	logger.Debugf("handleJoinRoomAtRandom")

	appKey, err := sv.authUser(h)
	if err != nil {
		renderErrorResponse(w, "Failed to user auth", http.StatusUnauthorized, err, logger)
		return
	}

	var param lobby.JoinParam
	err = msgpackDecode(r.Body, &param)
	if err != nil {
		renderErrorResponse(w, "Failed to read request body", http.StatusBadRequest, err, logger)
		return
	}

	macKey, err := auth.DecryptMACKey(appKey, param.EncMACKey)
	if err != nil {
		renderErrorResponse(w, "Failed to read MAC Key", http.StatusBadRequest, err, logger)
		return
	}

	vars := NewJoinVars(r)
	searchGroup := vars.searchGroup()
	logger = logger.With(log.KeySearchGroup, searchGroup)

	room, err := sv.roomService.JoinAtRandom(ctx, h.appId, searchGroup, param.Queries, param.ClientInfo, macKey, logger)
	if err != nil {
		renderErrorResponse(w, "Failed to join room", http.StatusInternalServerError, err, logger)
		return
	}

	renderJoinedRoomResponse(w, room, logger)
}

func (sv *LobbyService) handleSearchRooms(w http.ResponseWriter, r *http.Request) {
	h := parseSpecificHeader(r)
	logger := prepareLogger("lobby:search", h, r)
	logger.Debugf("handleSearchRoom")

	if _, err := sv.authUser(h); err != nil {
		renderErrorResponse(w, "Failed to user auth", http.StatusUnauthorized, err, logger)
		return
	}

	var param lobby.SearchParam
	err := msgpackDecode(r.Body, &param)
	if err != nil {
		renderErrorResponse(w, "Failed to read request body", http.StatusBadRequest, err, logger)
		return
	}

	logger.Debugf("search param: %#v", param)
	logger = logger.With(log.KeySearchGroup, param.SearchGroup)

	rooms, err := sv.roomService.Search(r.Context(),
		h.appId, param.SearchGroup, param.Queries, int(param.Limit), param.CheckJoinable, param.CheckWatchable, logger)
	if err != nil {
		renderErrorResponse(w, "Failed to search rooms", http.StatusInternalServerError, err, logger)
		return
	}

	renderFoundRoomsResponse(w, rooms, logger)
}

func (sv *LobbyService) handleSearchByIds(w http.ResponseWriter, r *http.Request) {
	h := parseSpecificHeader(r)
	logger := prepareLogger("lobby:search/ids", h, r)
	logger.Debugf("handleSearchByIds")

	if _, err := sv.authUser(h); err != nil {
		renderErrorResponse(w, "Failed to user auth", http.StatusUnauthorized, err, logger)
		return
	}

	var param lobby.SearchByIdsParam
	err := msgpackDecode(r.Body, &param)
	if err != nil {
		renderErrorResponse(w, "Failed to read request body", http.StatusBadRequest, err, logger)
		return
	}

	logger.Debugf("search param: %#v", param)
	logger = logger.With(log.KeyRoomIds, param.RoomIDs)

	rooms, err := sv.roomService.SearchByIds(r.Context(), h.appId, param.RoomIDs, param.Queries, logger)
	if err != nil {
		renderErrorResponse(w, "Failed to list rooms", http.StatusInternalServerError, err, logger)
		return
	}

	renderFoundRoomsResponse(w, rooms, logger)
}

func (sv *LobbyService) handleSearchByNumbers(w http.ResponseWriter, r *http.Request) {
	h := parseSpecificHeader(r)
	logger := prepareLogger("lobby:search/numbers", h, r)
	logger.Debugf("handleSearchByNumbers")

	if _, err := sv.authUser(h); err != nil {
		renderErrorResponse(w, "Failed to user auth", http.StatusUnauthorized, err, logger)
		return
	}

	var param lobby.SearchByNumbersParam
	err := msgpackDecode(r.Body, &param)
	if err != nil {
		renderErrorResponse(w, "Failed to read request body", http.StatusBadRequest, err, logger)
		return
	}

	logger.Debugf("search param: %#v", param)
	logger = logger.With(log.KeyRoomNumbers, param.RoomNumbers)

	rooms, err := sv.roomService.SearchByNumbers(r.Context(), h.appId, param.RoomNumbers, param.Queries, logger)
	if err != nil {
		renderErrorResponse(w, "Failed to list rooms", http.StatusInternalServerError, err, logger)
		return
	}

	renderFoundRoomsResponse(w, rooms, logger)
}

func (sv *LobbyService) handleSearchCurrentRooms(w http.ResponseWriter, r *http.Request) {
	h := parseSpecificHeader(r)
	logger := prepareLogger("lobby:search/current", h, r)
	logger.Debugf("handleSearchCurrentRooms")

	if _, err := sv.authUser(h); err != nil {
		renderErrorResponse(w, "Failed to user auth", http.StatusUnauthorized, err, logger)
		return
	}

	var param lobby.SearchCurrentRoomsParam
	err := msgpackDecode(r.Body, &param)
	if err != nil {
		renderErrorResponse(w, "Failed to read request body", http.StatusBadRequest, err, logger)
		return
	}

	logger.Debugf("search current param: %#v", param)

	rooms, err := sv.roomService.SearchCurrentRooms(r.Context(), h.appId, h.userId, param.Queries, logger)
	if err != nil {
		renderErrorResponse(w, "Failed to get search rooms", http.StatusInternalServerError, err, logger)
		return
	}

	renderFoundRoomsResponse(w, rooms, logger)
}

func (sv *LobbyService) handleWatchRoom(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(sv.conf.ApiTimeout))
	defer cancel()

	h := parseSpecificHeader(r)
	logger := prepareLogger("lobby:watch/id", h, r)
	logger.Debugf("handleWatchRoom")

	appKey, err := sv.authUser(h)
	if err != nil {
		renderErrorResponse(w, "Failed to user auth", http.StatusUnauthorized, err, logger)
		return
	}

	var param lobby.JoinParam
	err = msgpackDecode(r.Body, &param)
	if err != nil {
		renderErrorResponse(w, "Failed to read request body", http.StatusBadRequest, err, logger)
		return
	}

	macKey, err := auth.DecryptMACKey(appKey, param.EncMACKey)
	if err != nil {
		renderErrorResponse(w, "Failed to read MAC Key", http.StatusBadRequest, err, logger)
		return
	}

	vars := NewJoinVars(r)
	roomId := vars.roomId()
	if roomId == "" {
		renderErrorResponse(
			w, "Invalid room id", http.StatusBadRequest, xerrors.Errorf("Invalid room id"), logger)
		return
	}
	logger = logger.With(log.KeyRoom, roomId)

	room, err := sv.roomService.WatchById(ctx, h.appId, roomId, param.Queries, param.ClientInfo, macKey, logger)
	if err != nil {
		renderErrorResponse(w, "Failed to watch room", http.StatusInternalServerError, err, logger)
		return
	}

	renderJoinedRoomResponse(w, room, logger)
}

func (sv *LobbyService) handleWatchRoomByNumber(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(sv.conf.ApiTimeout))
	defer cancel()

	h := parseSpecificHeader(r)
	logger := prepareLogger("lobby:watch/number", h, r)
	logger.Debugf("handleWatchRoomByNumber")

	appKey, err := sv.authUser(h)
	if err != nil {
		renderErrorResponse(w, "Failed to user auth", http.StatusUnauthorized, err, logger)
		return
	}

	var param lobby.JoinParam
	err = msgpackDecode(r.Body, &param)
	if err != nil {
		renderErrorResponse(w, "Failed to read request body", http.StatusBadRequest, err, logger)
		return
	}

	macKey, err := auth.DecryptMACKey(appKey, param.EncMACKey)
	if err != nil {
		renderErrorResponse(w, "Failed to read MAC Key", http.StatusBadRequest, err, logger)
		return
	}

	vars := NewJoinVars(r)
	roomNumber := vars.roomNumber()
	if roomNumber == 0 {
		renderErrorResponse(
			w, "Invalid room number", http.StatusBadRequest, xerrors.Errorf("Invalid room number: 0"), logger)
		return
	}
	logger = logger.With(log.KeyRoomNumber, roomNumber)

	room, err := sv.roomService.WatchByNumber(ctx, h.appId, roomNumber, param.Queries, param.ClientInfo, macKey, logger)
	if err != nil {
		renderErrorResponse(w, "Failed to watch room", http.StatusInternalServerError, err, logger)
		return
	}

	renderJoinedRoomResponse(w, room, logger)
}

// 対象ユーザーをKickする。ゲームAPIサーバーからリクエストされる。
// php, Python等からアクセスしやすくするために、msgpackではなくてJSONを使う。
func (sv *LobbyService) handleAdminKick(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(sv.conf.ApiTimeout))
	defer cancel()

	h := parseSpecificHeader(r)
	logger := prepareLogger("lobby:admin/kick", h, r)
	if h.appId != h.userId {
		err := xerrors.Errorf("bad userID: appID=%q userID=%q", h.appId, h.userId)
		renderErrorResponse(w, "Failed to auth", http.StatusForbidden, err, logger)
		return
	}

	_, err := sv.authUser(h)
	if err != nil {
		renderErrorResponse(w, "Failed to user auth", http.StatusUnauthorized, err, logger)
		return
	}

	var req lobby.AdminKickParam
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		renderErrorResponse(w, "failed to decode JSON request", http.StatusBadRequest, err, logger)
		return
	}

	err = sv.roomService.AdminKick(ctx, h.appId, req.TargetID, logger)
	if err != nil {
		renderErrorResponse(w, "Internal Server Error", http.StatusInternalServerError, err, logger)
		return
	}
	logger.Infof("Rresponse(OK): kick by admin: %v", req.TargetID)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"msg": "ok"}`))
}
