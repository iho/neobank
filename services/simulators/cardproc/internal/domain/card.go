package domain

import "time"

// Card is a virtual card this simulator issued on behalf of the card
// service, identified externally by Ref (this card's ID).
type Card struct {
	CreatedAt      time.Time
	ID             string
	ExternalRef    string
	CardholderName string
	PANToken       string
	LastFour       string
	Status         string
	ExpiryMonth    int
	ExpiryYear     int
}
