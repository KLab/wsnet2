package service

import (
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

func (sv *LobbyService) handleCreateRoom(w http.ResponseWriter, r *http.Request) {
	appID := r.Header.Get("X-App-Id")
	userID := r.Header.Get("X-User-Id")

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Errorf("Failed to read request body: %w", err)
		http.Error(w, "Failed to request body", http.StatusInternalServerError)
		return
	}
	r.Body.Close()

	params := make(map[string]interface{})
	msgpack.Unmarshal(body, &params)

	log.Debugf("%v", params)

	roomOption := &pb.RoomOption{
		Visible:   true,
		Watchable: true,
		LogLevel:  4,
	}
	clientInfo := &pb.ClientInfo{
		Id: userID,
	}
	room, err := sv.roomService.Create(appID, roomOption, clientInfo)
	if err != nil {
		log.Errorf("Failed to create room: %w", err)
		http.Error(w, "Failed to create room", http.StatusInternalServerError)
		return
	}
	log.Debugf("%#v", room)
	res, err := msgpack.Marshal(room)
	if err != nil {
		log.Errorf("Failed to marshal room: %w", err)
		http.Error(w, "Failed to marshal room", http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/x-msgpack; charset=utf-8")
	w.Write(res)
}
