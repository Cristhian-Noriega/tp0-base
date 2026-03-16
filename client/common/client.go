package common

import (
	"bufio"
	"fmt"
	"net"
	"time"
	"os"

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

func (c *Client) StartClient(bet Bet, quit chan os.Signal) {
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
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop(quit chan os.Signal) {
	// There is an autoincremental msgID to identify every message sent
	// Messages if the message amount threshold has not been surpassed
	for msgID := 1; msgID <= c.config.LoopAmount; msgID++ {
		// Create the connection the server in every loop iteration. Send an
		err := c.createClientSocket()
		if err != nil {
			log.Errorf("action: create_socket | result: fail | client_id: %v | error: %v", c.config.ID, err)
			return
		}
		

		// TODO: Modify the send to avoid short-write
		fmt.Fprintf(
			c.conn,
			"[CLIENT %v] Message N°%v\n",
			c.config.ID,
			msgID,
		)
		msg, err := bufio.NewReader(c.conn).ReadString('\n')
		c.conn.Close()
		log.Infof("action: close_resource | result: success | client_id: %v | resource: connection", c.config.ID)

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
		case <-time.After(c.config.LoopPeriod):
			// Continue loop
		case <-quit:
			log.Infof("action: loop_interrupted | result: success | client_id: %v", c.config.ID)
			return
		}

	}
	log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
}
