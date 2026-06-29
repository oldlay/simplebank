package val

import (
	"fmt"
	"net/mail"
	"regexp"

	"github.com/oldlay/simplebank/util"

	decimal "google.golang.org/genproto/googleapis/type/decimal"
)

var (
	isValidUsername = regexp.MustCompile(`^[a-z0-9_]+$`).MatchString
	isValidFullName = regexp.MustCompile(`^[a-zA-Z\s]+$`).MatchString
)

func ValidateString(value string, minLength int, maxLength int) error {
	n := len(value)
	if n < minLength || n > maxLength {
		return fmt.Errorf("must contain from %d - %d characters", minLength, maxLength)
	}
	return nil
}

func ValidateUsername(value string) error {
	if err := ValidateString(value, 3, 20); err != nil {
		return err
	}
	if !isValidUsername(value) {
		return fmt.Errorf("must contain only lowercase letters, digits, or underscore")
	}
	return nil
}

func ValidatePassword(value string) error {
	return ValidateString(value, 6, 20)
}

func ValidateEmail(value string) error {
	if err := ValidateString(value, 3, 200); err != nil {
		return err
	}
	if _, err := mail.ParseAddress(value); err != nil {
		return fmt.Errorf("is not a valid email address")
	}
	return nil
}

func ValidateFullName(value string) error {
	if err := ValidateString(value, 3, 20); err != nil {
		return err
	}
	if !isValidFullName(value) {
		return fmt.Errorf("must contain only letters or spaces")
	}
	return nil
}

func ValidateEmailId(value int64) error {
	if value <= 0 {
		return fmt.Errorf("must be a positive integer")
	}
	return nil
}

func ValidateSecretCode(value string) error {
	return ValidateString(value, 32, 128)
}

func ValidateCurrency(currency string) error {
	if util.IsSupportedCurrency(currency) == false {
		return fmt.Errorf("must be one of USD, EUR and CAD")
	}
	return nil
}

func ValidateId(id int64) error {
	if id <= 0 {
		return fmt.Errorf("must be a positive integer")
	}
	return nil
}

func ValidateAmount(amount *decimal.Decimal) error {
	_, err := util.ProtoToShopDecimal(amount)
	if err != nil {
		return fmt.Errorf("amount must be a valid decimal")
	}
	return nil
}
