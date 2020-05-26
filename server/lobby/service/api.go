package service

import (
	"bytes"
	"context"
	"io/ioutil"
	"net"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/vmihailenco/msgpack/v4"
	"golang.org/x/xerrors"

	"wsnet2/log"
	"wsnet2/pb"
)

func (sv *LobbyService) serveAPI(ctx context.Context) <-chan error {
	errCh := make(chan error)

	go func() {
		network := sv.conf.Net
		laddr := sv.conf.Addr
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
}

func parseRequest(r *http.Request) (map[string]interface{}, error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	r.Body.Close()
	params := make(map[string]interface{})
	msgpack.Unmarshal(body, &params)
	return params, nil
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
	appID := r.Header.Get("X-App-Id")
	userID := r.Header.Get("X-User-Id")

	log.Infof("handleCreateRoom: appID=%s, userID=%s", appID, userID)

	var param CreateParam
	err := msgpack.NewDecoder(r.Body).UseJSONTag(true).Decode(&param)
	if err != nil {
		log.Errorf("Failed to read request body: %v", err)
		http.Error(w, "Failed to request body", http.StatusInternalServerError)
		return
	}

	// TODO: 必要に応じて一部のパラメータを上書き？

	room, err := sv.roomService.Create(appID, &param.RoomOption, &param.ClientInfo)
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
