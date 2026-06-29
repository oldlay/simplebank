package util

import (
	"github.com/shopspring/decimal"
	google_decimal "google.golang.org/genproto/googleapis/type/decimal"
)

func ShopDecimalToProto(shopDec decimal.Decimal) *google_decimal.Decimal {
	return &google_decimal.Decimal{
		Value: shopDec.String(),
	}
}

func ProtoToShopDecimal(gDecimal *google_decimal.Decimal) (decimal.Decimal, error) {
	if gDecimal == nil {
		return decimal.Zero, nil
	}
	val := gDecimal.GetValue()
	if val == "" {
		return decimal.Zero, nil
	}
	return decimal.NewFromString(val)
}
