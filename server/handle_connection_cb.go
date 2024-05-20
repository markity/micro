package server

import goreactor "github.com/markity/go-reactor"

func handleConn(conn goreactor.TCPConnection) {
	conn.SetDisConnectedCallback(handleClose)
}
