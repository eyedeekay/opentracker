module github.com/eyedeekay/tracker

go 1.14

require (
	github.com/eyedeekay/sam3 v0.32.3
	github.com/gorilla/websocket v1.4.2
	github.com/justinas/alice v1.2.0
	github.com/vvampirius/retracker v0.0.0-20171226134001-fdbec17ad537
	github.com/zeebo/bencode v1.0.0 // indirect
)

replace github.com/vvampirius/retracker v0.0.0-20171226134001-fdbec17ad537 => github.com/eyedeekay/retracker v0.0.0-20191208024817-1068d9dccb6d
