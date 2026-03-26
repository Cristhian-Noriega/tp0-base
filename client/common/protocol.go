package common

import (
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

func serializeBet(bet Bet) string {
	return strings.Join([]string{
		strconv.Itoa(bet.Agency),
		bet.FirstName,
		bet.LastName,
		bet.Document,
		bet.Birthdate,
		strconv.Itoa(bet.Number),
	}, "\n") + "\n"
}

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
	payload := serializeBet(bet)

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

func BetPacketSize(bet Bet) int {
	return headerSize + len(serializeBet(bet))
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
