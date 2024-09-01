package common

import (
	"bufio"
	"net"
	"time"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	LoopAmount    int
	LoopPeriod    time.Duration
}

// Client Entity that encapsulates how
type Client struct {
	config    ClientConfig
	conn      net.Conn
	chnl      chan bool
	betReader *BetReader
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig, chnl chan bool) *Client {
	client := &Client{
		config:    config,
		chnl:      chnl,
		betReader: NewBetReader(),
	}
	return client
}

// CreateClientSocket Initializes client socket. In case of
// failure, error is printed in stdout/stderr and exit 1
// is returned
func (c *Client) createClientSocket() error {
	conn, err := net.Dial("tcp", c.config.ServerAddress)
	if err != nil {
		log.Criticalf(
			"action: connect | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
	}
	c.conn = conn
	return nil
}

func (c *Client) sendBets(bets *Bet) error {
	encodedBet, err := bets.Encode()
	if err != nil {
		log.Errorf("action: encode_bet | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return err
	}

	_, err = bets.safeWriteBytes(c.conn, encodedBet)
	if err != nil {
		log.Errorf("action: send_bet | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return err
	}

	return nil
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop() {
	// There is an autoincremental msgID to identify every message sent
	// Messages if the message amount threshold has not been surpassed
clientLoop:
	for msgID := 1; msgID <= c.config.LoopAmount; msgID++ {
		// Create the connection the server in every loop iteration. Send an
		c.createClientSocket()

		//Obtain the bet from the BetReader
		bets := c.betReader.ReadBets()

		// Send the bet to the server
		c.sendBets(bets)

		// // TODO: Modify the send to avoid short-write
		// fmt.Fprintf(
		// 	c.conn,
		// 	"[CLIENT %v] Message N°%v\n",
		// 	c.config.ID,
		// 	msgID,
		// )
		msg, err := bufio.NewReader(c.conn).ReadString('\n')
		c.conn.Close()

		if err != nil {
			log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			return
		}

		log.Infof("action: receive_message | result: success | client_id: %v | msg: %v",
			c.config.ID,
			msg,
		)

		// Wait a time between sending one message and the next one
		select {
		case <-c.chnl:
			log.Infof("action: SIGTERM received | result: finishing early | client_id: %v", c.config.ID)
			break clientLoop

		case <-time.After(c.config.LoopPeriod):
			continue
		}

	}
	log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
}
