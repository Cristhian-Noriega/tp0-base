package common

import (
	"io"
	"net"
	"strconv"
	"strings"
)


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
	header := []byte{byte(length >> 8), byte(length & 0xFF)}

	return sendAll(conn, append(header, []byte(payload)...))
}


func RecvAck(conn net.Conn) (bool, error) {
	buffer := make([]byte, 1)
	_, err := io.ReadFull(conn, buffer)
	if err != nil {
		return false, err
	}
	return buffer[0] == 0x01, nil
}


func SendBatch(bets []Bet, conn net.Conn) error {
	n := len(bets)
	countHeader := []byte{byte(n >> 8), byte(n & 0xFF)}
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