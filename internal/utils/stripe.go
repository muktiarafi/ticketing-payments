package utils

import (
	"os"

	"github.com/stripe/stripe-go/v72"
)

func InitStripe() {
	key := os.Getenv("STRIPE_KEY")
	stripe.Key = key
}
