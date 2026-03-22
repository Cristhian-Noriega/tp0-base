package common

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	csvMinColumns   = 5
	csvFirstNameIdx = 0
	csvLastNameIdx  = 1
	csvDocumentIdx  = 2
	csvBirthdateIdx = 3
	csvNumberIdx    = 4
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
		if len(row) < csvMinColumns {
			continue
		}
		number, err := strconv.Atoi(strings.TrimSpace(row[csvNumberIdx]))
		if err != nil {
			return nil, fmt.Errorf("invalid number in csv: %w", err)
		}
		bets = append(bets, Bet{
			Agency:    agency,
			FirstName: strings.TrimSpace(row[csvFirstNameIdx]),
			LastName:  strings.TrimSpace(row[csvLastNameIdx]),
			Document:  strings.TrimSpace(row[csvDocumentIdx]),
			Birthdate: strings.TrimSpace(row[csvBirthdateIdx]),
			Number:    number,
		})
	}
	return bets, nil
}
