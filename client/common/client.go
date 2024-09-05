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
const LEN_SERVER_MSG_SIZE = 2

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
		return err
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

// sendWaitingForLotteryMessage Sends a message to the server indicating
// that the client is waiting for the lottery to start
func (c *Client) sendWaitingForLotteryMessage() error {
	encodedBytesLen := make([]byte, SIZE_UINT16)
	binary.LittleEndian.PutUint16(encodedBytesLen, uint16(1))

	if _, err := SafeWriteBytes(c.conn, encodedBytesLen); err != nil {
		return err
	}

	if _, err := SafeWriteBytes(c.conn, []byte{c.betReader.agency}); err != nil {
		return err
	}

	return nil
}

// obtainBetsFilePath Returns the path where the bets file is stored
// based on the client ID (which is really the agency also)
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

mainCLientLoop:
	for {
		func() {
			// Create the connection the server in every loop iteration.
			if err := c.createClientSocket(); err != nil {
				log.Errorf("action: create_socket | result: fail | client_id: %v | error: %v",
					c.config.ID,
					err,
				)
				return
			}
		}()

		// This closes the socket on every iteration so we dont have to worry about
		// closing it manually on every case that we want to cut the connection.
		// That is why we enclosed the socket creation in a function, so this defer can
		// close the socket in every case
		defer c.conn.Close()

		// Obtain the bet from the BetReader
		bets := c.betReader.ReadBets()

		if len(bets) != 0 {
			// Having to send bets means that the client has not finished yet
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

			// Read the answer from the server
			res, err := SafeReadVariableBytes(bufio.NewReader(c.conn))

			if res == nil {
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

			// Process the server response and log the result
			ServerResponse := ServerResponseFromBytes(res)

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

			// Wait a time between sending one message and the next one, and also
			// handle SIGTERN Signaling
			select {
			case <-c.chnl:
				log.Infof("action: SIGTERM received | result: finishing early | client_id: %v", c.config.ID)
				break mainCLientLoop

			case <-time.After(c.config.LoopPeriod):
				continue
			}
		} else {
			// Lottery time!
			// Send the message to the server indicating that we are waiting for the lottery
			if err := c.sendWaitingForLotteryMessage(); err != nil {
				log.Errorf("action: send_waiting_for_lottery | result: fail | client_id: %v | error: %v",
					c.config.ID,
					err,
				)
				return
			}

			log.Infof("action: waiting_for_lottery | result: success | client_id: %v", c.config.ID)

			lotteryChannel := make(chan []byte)

			// Fire the go routine that waits for the winners, this way we are able
			// tu cut execution if we receive a SIGTERM
			go func() {
				res, err := SafeReadVariableBytes(bufio.NewReader(c.conn))
				if len(res) == 0 {
					log.Infof("action: receive_message | result: server disconnected | client_id: %v",
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

				lotteryChannel <- res
			}()

			var serverResponse *ServerResponse
			// Keep blocked until a message is received or a SIGTERM is received
			select {
			case <-c.chnl:
				log.Infof("action: SIGTERM received | result: finishing early | client_id: %v", c.config.ID)
				c.conn.Close()
				break mainCLientLoop

			case res := <-lotteryChannel:
				serverResponse = ServerResponseFromBytes(res)
			}

			log.Infof("action: consulta_ganadores | result: success | cant_ganadores: %d",
				len(serverResponse.Winners))

			break mainCLientLoop
		}
	}

	log.Infof("action: client_finished | result: success | client_id: %v", c.config.ID)
}
