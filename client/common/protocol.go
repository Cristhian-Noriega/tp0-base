package common

import (
	"fmt"
	"net"
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
	payload := fmt.Sprintf("%d %s %s %s %s %d\n", bet.Agency, bet.FirstName, bet.LastName, bet.Document, bet.Birthdate, bet.Number)
	
	lenght := len(payload)
	header := []byte{byte(lenght >> 8), byte(lenght & 0xFF)}

	return sendAll(conn, append(header, []byte(payload)...))
}


func RecvAck(conn net.Conn) (bool, error) {
	buffer := make([]byte, 1)
	_, err := conn.Read(buffer)
	if err != nil {
		return false, err
	}
	return buffer[0] == 0x01, nil
}