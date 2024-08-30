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
	Finished      bool
}

func NewGracefulFinisher(mainChannel chan bool, client *Client, clientChannel chan bool) *GracefulFinisher {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGTERM, syscall.SIGINT)

	return &GracefulFinisher{
		signalChannel: signalChannel,
		mainChannel:   mainChannel,
		client:        client,
		clientChannel: clientChannel,
		Finished:      false,
	}
}

func (g *GracefulFinisher) finishGracefully() {
	g.Finished = true
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
