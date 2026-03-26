package common

import (
	"encoding/csv"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

const maxBatchSizeBytes = 8 * 1024

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

func (c *Client) readCSV(csvPath string, quit chan os.Signal) error {
	f, err := os.Open(csvPath)
	if err != nil {
		return err
	}
	defer f.Close()

	reader := csv.NewReader(f)
	agency, _ := strconv.Atoi(c.config.ID)
	var chunk []Bet
	chunkSizeBytes := headerSize

	for {
		select {
		case <-quit:
			log.Infof("action: loop_interrupted | result: success | client_id: %v", c.config.ID)
			return nil
		default:
		}

		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Warningf("action: read_csv | result: fail | client_id: %v | error: %v", c.config.ID, err)
			continue
		}

		if len(row) < csvMinColumns {
			continue
		}

		bet, err := parseBetRow(row, agency)
		if err != nil {
			continue
		}

		betSizeBytes := BetPacketSize(bet)
		if betSizeBytes + headerSize > maxBatchSizeBytes {
			log.Warningf(
				"action: read_csv | result: fail | client_id: %v | error: bet_exceeds_max_batch_size | dni: %v | numero: %v | size_bytes: %v",
				c.config.ID,
				bet.Document,
				bet.Number,
				betSizeBytes + headerSize,
			)
			continue
		}

		nextChunkSizeBytes := chunkSizeBytes + betSizeBytes
		if len(chunk) > 0 && (len(chunk) >= c.config.BatchMaxAmount || nextChunkSizeBytes > maxBatchSizeBytes) {
			if err := c.sendChunk(chunk); err != nil {
				return err
			}
			chunk = nil
			chunkSizeBytes = headerSize
			nextChunkSizeBytes = chunkSizeBytes + betSizeBytes
		}

		chunk = append(chunk, bet)
		chunkSizeBytes = nextChunkSizeBytes

		if len(chunk) >= c.config.BatchMaxAmount {
			if err := c.sendChunk(chunk); err != nil {
				return err
			}
			chunk = nil
			chunkSizeBytes = headerSize
		}
	}

	if len(chunk) > 0 {
		if err := c.sendChunk(chunk); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) StartClient(csvPath string, quit chan os.Signal) {
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

	if err := c.readCSV(csvPath, quit); err != nil {
		log.Errorf("action: read_csv | result: fail | client_id: %v | error: %v", c.config.ID, err)
		return
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
