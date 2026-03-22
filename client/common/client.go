package common

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID             string
	ServerAddress  string
	LoopAmount     int
	LoopPeriod     time.Duration
	BatchMaxAmount int
}

// Client Entity that encapsulates how
type Client struct {
	config ClientConfig
	conn   net.Conn
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	client := &Client{
		config: config,
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

func (c *Client) StartClient(bets []Bet, quit chan os.Signal) {
	select {
	case <-quit:
		log.Infof("action: loop_interrupted | result: success | client_id: %v", c.config.ID)
		return
	default:
	}

	err := c.createClientSocket()
	if err != nil {
		log.Errorf("action: create_socket | result: fail | client_id: %v | error: %v", c.config.ID, err)
		return
	}

	log.Infof("action: connect | result: success | client_id: %v | server_address: %v", c.config.ID, c.config.ServerAddress)
	defer func() {
		c.conn.Close()
		log.Infof("action: close_resource | result: success | client_id: %v | resource: connection", c.config.ID)
	}()

	for i := 0; i < len(bets); i += c.config.BatchMaxAmount {
		select {
		case <-quit:
			log.Infof("action: loop_interrupted | result: success | client_id: %v", c.config.ID)
			return
		default:
		}

		end := i + c.config.BatchMaxAmount
		if end > len(bets) {
			end = len(bets)
		}
		chunk := bets[i:end]
		if err := c.sendChunk(chunk); err != nil {
			log.Errorf("action: send_chunk | result: fail | client_id: %v | error: %v", c.config.ID, err)
			return
		}
	}
}

func (c *Client) sendChunk(chunk []Bet) error {
	if err := SendBatch(chunk, c.conn); err != nil {
		log.Errorf("action: send_batch | result: fail | client_id: %v | error: %v", c.config.ID, err)
		return err
	}

	success, err := RecvAck(c.conn)
	if err != nil {
		log.Errorf("action: apuesta_enviada | result: fail | client_id: %v | error: %v", c.config.ID, err)
		return err
	}
	if !success {
		log.Errorf("action: apuesta_enviada | result: fail | client_id: %v | error: invalid_ack", c.config.ID)
		return fmt.Errorf("invalid ack from server")
	}

	for _, bet := range chunk {
		log.Infof("action: apuesta_enviada | result: success | dni: %v | numero: %v", bet.Document, bet.Number)
	}
	return nil
}
