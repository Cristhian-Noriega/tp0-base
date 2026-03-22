package common

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// LoadBetsFromCSV reads agency bets from the given CSV file.
// Expected column order: nombre, apellido, documento, nacimiento, numero
func LoadBetsFromCSV(path string, agency int) ([]Bet, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	reader := csv.NewReader(f)
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	bets := make([]Bet, 0, len(rows))
	for _, row := range rows {
		if len(row) < 5 {
			continue
		}
		number, err := strconv.Atoi(strings.TrimSpace(row[4]))
		if err != nil {
			return nil, fmt.Errorf("invalid number in csv: %w", err)
		}
		bets = append(bets, Bet{
			Agency:    agency,
			FirstName: strings.TrimSpace(row[0]),
			LastName:  strings.TrimSpace(row[1]),
			Document:  strings.TrimSpace(row[2]),
			Birthdate: strings.TrimSpace(row[3]),
			Number:    number,
		})
	}
	return bets, nil
}
