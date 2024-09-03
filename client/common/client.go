package common

import (
	"bufio"
	"encoding/binary"
	"net"
	"strconv"
	"time"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

const SIZE_UINT16 = 2
const SERVER_MSG_SIZE = 4

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	LoopAmount    int
	LoopPeriod    time.Duration
	MaxBatchSize  int
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
	agency, _ := strconv.Atoi(config.ID)
	client := &Client{
		config:    config,
		chnl:      chnl,
		betReader: NewBetReader(config.MaxBatchSize, uint8(agency)),
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

// SendBets Sends the bet to the server, ensuring all the bytes
// are sent. if an error occurs, it is loggedand the function
// returns an error
func (c *Client) sendBets(bets []*Bet) error {
	encodedBets, err := EncodeBets(bets)
	if err != nil {
		log.Errorf("action: encode_bet | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return err
	}

	encodedBytesLen := make([]byte, SIZE_UINT16)
	binary.LittleEndian.PutUint16(encodedBytesLen, uint16(len(encodedBets)))
	_, err = SafeWriteBytes(c.conn, encodedBytesLen)
	if err != nil {
		log.Errorf("action: send_bet | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return err
	}

	_, err = SafeWriteBytes(c.conn, encodedBets)
	if err != nil {
		log.Errorf("action: send_bet | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return err
	}

	return nil
}

func (c *Client) sendWaitingForLotteryMessage() error {
	if _, err := SafeWriteBytes(c.conn, []byte{1}); err != nil {
		return err
	}

	if _, err := SafeWriteBytes(c.conn, []byte{c.betReader.agency}); err != nil {
		return err
	}

	return nil
}

func (c *Client) obtainBetsFilePath() string {
	return "/data/agency-" + c.config.ID + ".csv"
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop() {
	// There is an autoincremental msgID to identify every message sent
	// Messages if the message amount threshold has not been surpassed
	if err := c.betReader.OpenFile(c.obtainBetsFilePath()); err != nil {
		log.Infof("action: open_file | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return
	}

	defer c.betReader.CloseFile()

	for !c.betReader.Finished {
		// Create the connection the server in every loop iteration. Send an
		if err := c.createClientSocket(); err != nil {
			log.Errorf("action: create_socket | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			return
		}

		//Obtain the bet from the BetReader
		bets := c.betReader.ReadBets()

		// Send the bet to the server
		err := c.sendBets(bets)
		if err != nil {
			log.Errorf("action: send_bet | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			return
		}

		log.Infof("action: apuesta_enviada | result: success | bets_sent: %d",
			len(bets))

		responseBytes := make([]byte, SERVER_MSG_SIZE)

		n, err := SafeReadBytes(bufio.NewReader(c.conn), responseBytes)

		if n == 0 {
			log.Errorf("action: receive_message | result: server disconnected | client_id: %v",
				c.config.ID,
			)
			return
		}

		if err != nil {
			log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			return
		}

		ServerResponse := ServerResponseFromBytes(responseBytes)

		if ServerResponse.AmountOfBets > 0 {
			log.Infof("action: server_processed_bets | result: success | client_id: %v | bets_processed: %d",
				c.config.ID,
				ServerResponse.AmountOfBets,
			)
		} else {
			log.Infof("action: server_processed_bets | result: failed | client_id: %v ",
				c.config.ID,
			)
		}

		c.conn.Close()

		// Wait a time between sending one message and the next one
		select {
		case <-c.chnl:
			log.Infof("action: SIGTERM received | result: finishing early | client_id: %v", c.config.ID)
			return

		case <-time.After(c.config.LoopPeriod):
			continue
		}

	}
	log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)

	// Creates the last socket to the server to send the waiting for lottery message
	if err := c.createClientSocket(); err != nil {
		log.Errorf("action: create_socket | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return
	}

	if err := c.sendWaitingForLotteryMessage(); err != nil {
		log.Errorf("action: send_waiting_for_lottery | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return
	}

	log.Infof("action: waiting_for_lottery | result: success | client_id: %v", c.config.ID)

	response := make([]byte, SERVER_MSG_SIZE)
	if _, err := SafeReadBytes(bufio.NewReader(c.conn), response); err != nil {
		log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return
	}

	serverResponse := ServerResponseFromBytes(response)

	log.Infof("action: consulta_ganadores | result: success | cant_ganadores: %d",
		len(serverResponse.Winners))

	c.conn.Close()

	log.Infof("action: client_finished | result: success | client_id: %v", c.config.ID)

}
