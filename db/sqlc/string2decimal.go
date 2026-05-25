package db

import (
	"fmt"
	"log"

	"github.com/shopspring/decimal"
)

func String2Decimal(s string) decimal.Decimal {
	d, err := decimal.NewFromString(s)
	if err != nil {
		fmt.Fprintf(log.Writer(), "Error converting string to decimal: %v\n", err)
		return decimal.Decimal{}
	}
	return d
}
