package common

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
)

const (
	headerSize = 2
	ackSize    = 1
	byteShift  = 8
	byteMask   = 0xFF
	ackSuccess = 0x01
)

var ErrNotReady = fmt.Errorf("sorteo not ready")

func sendAll(conn net.Conn, data []byte) error {
	sent := 0
	for sent < len(data) {
		n, err := conn.Write(data[sent:])
		if err != nil {
			return err
		}
		sent += n
	}
	return nil
}

func SendBet(bet Bet, conn net.Conn) error {
	payload := strings.Join([]string{
		strconv.Itoa(bet.Agency),
		bet.FirstName,
		bet.LastName,
		bet.Document,
		bet.Birthdate,
		strconv.Itoa(bet.Number),
	}, "\n") + "\n"

	length := len(payload)
	header := make([]byte, headerSize)
	header[0] = byte(length >> byteShift)
	header[1] = byte(length & byteMask)

	return sendAll(conn, append(header, []byte(payload)...))
}

func RecvAck(conn net.Conn) (bool, error) {
	buffer := make([]byte, ackSize)
	_, err := io.ReadFull(conn, buffer)
	if err != nil {
		return false, err
	}
	return buffer[0] == ackSuccess, nil
}

func SendBatch(bets []Bet, conn net.Conn) error {
	n := len(bets)
	countHeader := make([]byte, headerSize)
	countHeader[0] = byte(n >> byteShift)
	countHeader[1] = byte(n & byteMask)
	if err := sendAll(conn, countHeader); err != nil {
		return err
	}
	for _, bet := range bets {
		if err := SendBet(bet, conn); err != nil {
			return err
		}
	}
	return nil
}

func SendFin(conn net.Conn) error {
	header := make([]byte, headerSize)
	return sendAll(conn, header)
}


func RecvWinners(conn net.Conn) ([]string, error) {
	ready := make([]byte, 1)
	if _, err := io.ReadFull(conn, ready); err != nil {
		return nil, err
	}

	if ready[0] == 0x00 { // NOT_READY
		return nil, ErrNotReady
	}

	// Read Count Header (number of winners)
	header := make([]byte, headerSize)
	if _, err := io.ReadFull(conn, header); err != nil {
		return nil, err
	}
	expectedCount := int(header[0])<<byteShift | int(header[1])

	// Read Size Header (total bytes of payload)
	sizeHeader := make([]byte, headerSize)
	if _, err := io.ReadFull(conn, sizeHeader); err != nil {
		return nil, err
	}
	payloadSize := int(sizeHeader[0])<<byteShift | int(sizeHeader[1])

	if payloadSize == 0 {
		return []string{}, nil
	}

	payload := make([]byte, payloadSize)
	if _, err := io.ReadFull(conn, payload); err != nil {
		return nil, err
	}

	content := strings.TrimSpace(string(payload))
	if content == "" {
		return []string{}, nil
	}
	dnis := strings.Split(content, "\n")

	// Verify count matches expected amount of winners
	if len(dnis) != expectedCount {
		log.Warningf("action: recv_winners | result: count_mismatch | expected: %v | got: %v", expectedCount, len(dnis))
	}

	return dnis, nil
}

