package util

import (
	"math/rand"
	"strings"
	"time"
)

const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const number = "0123456789"

func init() {
	rand.Seed(time.Now().UnixNano())
}

// RandomInt generates random int between min and max
func RandomInt(min, max int64) int64 {
	return min + rand.Int63n(max-min+1)
}

// RandomFloat generates random String
func RandomString(n int) string {
	var sb strings.Builder
	k := len(alphabet)

	for i := 0; i < n; i++ {
		c := alphabet[rand.Intn(k)]
		sb.WriteByte(c)
	}
	return sb.String()
}

// RandomOwner generates random owner name
func RandomOwner() string {
	return RandomString(6)
}

// RandomMoney generates random money amount
func RandomMoney(n int) string {
	var s []byte

	k := len(number)

	for i := 0; i < n; i++ {
		c := number[rand.Intn(k)]
		s = append(s, c)
	}
	index := RandomInt(1, int64(n-1))

	s = append(s[:index], append([]byte{'.'}, s[index:]...)...)

	if s[0] == '.' || s[0] == '0' {
		s[0] = number[RandomInt(1, 9)]
	}

	return string(s)
}

// RandomCurrency generates random currency
func RandomCurrency() string {
	currencies := []string{"USD", "EUR", "CAD"}
	n := len(currencies)
	return currencies[rand.Intn(n)]
}
