package common

import (
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

func parseBetRow(row []string, agency int) (Bet, error) {
	num, err := strconv.Atoi(strings.TrimSpace(row[csvNumberIdx]))
	if err != nil {
		return Bet{}, err
	}
	return Bet{
		Agency:    agency,
		FirstName: strings.TrimSpace(row[csvFirstNameIdx]),
		LastName:  strings.TrimSpace(row[csvLastNameIdx]),
		Document:  strings.TrimSpace(row[csvDocumentIdx]),
		Birthdate: strings.TrimSpace(row[csvBirthdateIdx]),
		Number:    num,
	}, nil
}
