package common

import (
	"os"
	"os/signal"
	"syscall"
)

type GracefulFinisher struct {
	mainChannel   chan bool
	signalChannel chan os.Signal
	clientChannel chan bool
	client        *Client
	finished      bool
}

func NewGracefulFinisher(mainChannel chan bool, client *Client, clientChannel chan bool) *GracefulFinisher {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGTERM)

	return &GracefulFinisher{
		signalChannel: signalChannel,
		mainChannel:   mainChannel,
		client:        client,
		clientChannel: clientChannel,
		finished:      false,
	}
}

func (g *GracefulFinisher) finishGracefully() {
	g.finished = true
	g.clientChannel <- true
}

func (g *GracefulFinisher) StartGracefulFinisher() {
	select {
	case <-g.signalChannel:
		g.finishGracefully()

	case <-g.mainChannel:
		return
	}
}
