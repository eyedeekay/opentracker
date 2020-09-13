package samcore

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"encoding/json"

	"github.com/gorilla/websocket"
	"github.com/justinas/alice"
	//	"github.com/vvampirius/retracker/bittorrent/tracker"
	"github.com/vvampirius/retracker/core/common"
	Receiver "github.com/vvampirius/retracker/core/receiver"
	Storage "github.com/vvampirius/retracker/core/storage"
)

import (
	"github.com/eyedeekay/sam3"
	"github.com/eyedeekay/sam3/i2pkeys"
	"os"
)

var (
	ln net.Listener
	//  storage *Announce
	core Core
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
			var dat map[string]interface{}
      if err := json.Unmarshal(p, &dat); err != nil {
				log.Println(err)
        return
      }

			rr := core.Receiver.Announce.ProcessAnnounce(
				dat[`X-I2p-Dest-Base64`].(string),
				dat[`info_hash`].(string),
				dat[`peer_id`].(string),
				dat[`port`].(string),
				dat[`uploaded`].(string),
				dat[`downloaded`].(string),
				dat[`left`].(string),
				dat[`ip`].(string),
				dat[`numwant`].(string),
				dat[`event`].(string),
				dat[`compact`].(string),
			)
			if d, err := rr.Bencode(); err == nil {
				fmt.Fprint(w, d)
				if err := conn.WriteMessage(messageType, []byte(d)); err != nil {
					log.Println(err)
					return
				}
			} else {
				//self.Logger.Println(err.Error())
			}

		}
	})
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	fmt.Fprintf(w, "<!DOCTYPE html>")
	fmt.Fprintf(w, "<html>")
	fmt.Fprintf(w, "  <body>")
	fmt.Fprintf(w, "    <h1>Rhizome Open Tracker:</h1>")
	fmt.Fprintf(w, "    <ul>")
	fmt.Fprintf(w, "      <li>")
	fmt.Fprintf(w, "        <span>Base32 URL: </span>")
	fmt.Fprintf(w, "        <a href=\"http://%s\"/>%s</a>", ln.Addr().(i2pkeys.I2PAddr).Base32(), ln.Addr().(i2pkeys.I2PAddr).Base32())
	fmt.Fprintf(w, "      </li>")
	fmt.Fprintf(w, "      <li>")
	fmt.Fprintf(w, "        <span>Address Helper: </span>")
	fmt.Fprintf(w, "        <a href=\"http://%s.%s.i2p/?i2paddresshelper=%s\"/>%s.%s</a>", ln.Addr().(i2pkeys.I2PAddr).Base32()[0:5], "rhz-ot", ln.Addr(), ln.Addr().(i2pkeys.I2PAddr).Base32()[0:5], "rhz-ot.i2p")
	fmt.Fprintf(w, "      </li>")
	fmt.Fprintf(w, "    </ul>")
	fmt.Fprintf(w, "</body>")
	fmt.Fprintf(w, "</html>")
}

func New(config *common.Config) (*Core, error) {
	storage := Storage.New(config)
	core = Core{
		Config:   config,
		Storage:  storage,
		Receiver: Receiver.New(config, storage),
	}
	var err error
	ln, err = Sammy()
	if err != nil {
		return nil, err
	}
	defer ln.Close()
	ws := alice.New(wsMiddleware)
	http.HandleFunc("/", homeHandler)
	http.Handle("/announce", ws.Then(http.HandlerFunc(core.Receiver.Announce.HttpHandler)))
	http.Handle("/a", ws.Then(http.HandlerFunc(core.Receiver.Announce.HttpHandler)))
	http.Handle("/ws/announce", ws.Then(http.HandlerFunc(core.Receiver.Announce.HttpHandler)))
	http.Handle("/ws/a", ws.Then(http.HandlerFunc(core.Receiver.Announce.HttpHandler)))
	http.Handle("/announce/ws", ws.Then(http.HandlerFunc(core.Receiver.Announce.HttpHandler)))
	http.Handle("/a/ws", ws.Then(http.HandlerFunc(core.Receiver.Announce.HttpHandler)))
	ioutil.WriteFile("keys.base32.txt", []byte(ln.Addr().(i2pkeys.I2PAddr).Base32()), 0644)
	if err := http.Serve(ln, nil); err != nil { // set listen port
		return nil, err
	}
	//TODO: do it with context
	return &core, nil
}
