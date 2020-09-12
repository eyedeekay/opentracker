package samcore

import (
	"net/http"
	//"fmt"
	"log"

	"github.com/gorilla/websocket"
	"github.com/justinas/alice"
	"github.com/vvampirius/retracker/core/common"
	Receiver "github.com/vvampirius/retracker/core/receiver"
	Storage "github.com/vvampirius/retracker/core/storage"
)

import (
	"github.com/eyedeekay/sam3"
	"github.com/eyedeekay/sam3/i2pkeys"
	"os"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:    4096,
	WriteBufferSize:   4096,
	EnableCompression: true,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func Sammy() (*sam3.StreamListener, error) {
	if sam, err := sam3.NewSAM("127.0.0.1:7656"); err != nil {
		return nil, err
	} else {
		if file, err := os.Open("./keys.i2pkeys"); err == nil {
			if keys, err := i2pkeys.LoadKeysIncompat(file); err != nil {
				return nil, err
			} else {
				if stream, err := sam.NewStreamSession("serverTun", keys, sam3.Options_Fat); err != nil {
					return nil, err
				} else {
					return stream.Listen()
				}
			}
		} else {
			if keys, err := sam.NewKeys(); err != nil {
				return nil, err
			} else {
				if file, err := os.Create("./keys.i2pkeys"); err != nil {
					return nil, err
				} else {
					if err := i2pkeys.StoreKeysIncompat(keys, file); err != nil {
						return nil, err
					}
					if stream, err := sam.NewStreamSession("serverTun", keys, sam3.Options_Fat); err != nil {
						return nil, err
					} else {
						return stream.Listen()
					}
				}
			}
		}
	}

}

type Core struct {
	Config   *common.Config
	Storage  *Storage.Storage
	Receiver *Receiver.Receiver
}

func wsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		for {
			messageType, p, err := conn.ReadMessage()
			if err != nil {
				log.Println(err)
				return
			}
			if err := conn.WriteMessage(messageType, p); err != nil {
				log.Println(err)
				return
			}
		}
	})
}

func New(config *common.Config) (*Core, error) {
	storage := Storage.New(config)
	core := Core{
		Config:   config,
		Storage:  storage,
		Receiver: Receiver.New(config, storage),
	}
	ln, err := Sammy()
	if err != nil {
		return nil, err
	}
	ws := alice.New(wsMiddleware)
	http.Handle("/announce", ws.Then(http.HandlerFunc(core.Receiver.Announce.HttpHandler)))
	http.Handle("/a", ws.Then(http.HandlerFunc(core.Receiver.Announce.HttpHandler)))
	http.Handle("/ws/announce", ws.Then(http.HandlerFunc(core.Receiver.Announce.HttpHandler)))
	http.Handle("/ws/a", ws.Then(http.HandlerFunc(core.Receiver.Announce.HttpHandler)))
	http.Handle("/announce/ws", ws.Then(http.HandlerFunc(core.Receiver.Announce.HttpHandler)))
	http.Handle("/a/ws", ws.Then(http.HandlerFunc(core.Receiver.Announce.HttpHandler)))
	if err := http.Serve(ln, nil); err != nil { // set listen port
		return nil, err
	}
	//TODO: do it with context
	return &core, nil
}
