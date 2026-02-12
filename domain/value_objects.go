package domain

import (
	"fmt"
	"regexp"
	"strings"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// Email is an immutable value object representing a validated email address.
type Email struct {
	value string
}

// NewEmail creates a validated Email value object.
func NewEmail(email string) (Email, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" {
		return Email{}, fmt.Errorf("email cannot be empty")
	}
	if !emailRegex.MatchString(email) {
		return Email{}, fmt.Errorf("invalid email format: %s", email)
	}
	return Email{value: email}, nil
}

// String returns the email string value.
func (e Email) String() string { return e.value }

// Equals checks equality with another Email.
func (e Email) Equals(other Email) bool { return e.value == other.value }

// Phone is an immutable value object representing a phone number.
type Phone struct {
	countryCode string
	number      string
}

// NewPhone creates a validated Phone value object.
func NewPhone(countryCode, number string) (Phone, error) {
	countryCode = strings.TrimSpace(countryCode)
	number = strings.TrimSpace(number)
	if countryCode == "" || number == "" {
		return Phone{}, fmt.Errorf("country code and number are required")
	}
	return Phone{countryCode: countryCode, number: number}, nil
}

// String returns the full phone number.
func (p Phone) String() string { return p.countryCode + p.number }

// CountryCode returns the country code.
func (p Phone) CountryCode() string { return p.countryCode }

// Number returns the phone number without country code.
func (p Phone) Number() string { return p.number }

// Equals checks equality with another Phone.
func (p Phone) Equals(other Phone) bool {
	return p.countryCode == other.countryCode && p.number == other.number
}

// Address is an immutable value object representing a physical address.
type Address struct {
	Line1      string  `json:"line1"`
	Line2      string  `json:"line2,omitempty"`
	City       string  `json:"city"`
	State      string  `json:"state"`
	PostalCode string  `json:"postal_code"`
	Country    string  `json:"country"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
}

// NewAddress creates a validated Address value object.
func NewAddress(line1, line2, city, state, postalCode, country string, lat, lng float64) (Address, error) {
	if line1 == "" {
		return Address{}, fmt.Errorf("address line 1 is required")
	}
	if city == "" {
		return Address{}, fmt.Errorf("city is required")
	}
	if state == "" {
		return Address{}, fmt.Errorf("state is required")
	}
	if country == "" {
		return Address{}, fmt.Errorf("country is required")
	}
	if lat < -90 || lat > 90 {
		return Address{}, fmt.Errorf("latitude must be between -90 and 90")
	}
	if lng < -180 || lng > 180 {
		return Address{}, fmt.Errorf("longitude must be between -180 and 180")
	}
	return Address{
		Line1:      line1,
		Line2:      line2,
		City:       city,
		State:      state,
		PostalCode: postalCode,
		Country:    country,
		Latitude:   lat,
		Longitude:  lng,
	}, nil
}

// FullAddress returns a formatted string of the complete address.
func (a Address) FullAddress() string {
	parts := []string{a.Line1}
	if a.Line2 != "" {
		parts = append(parts, a.Line2)
	}
	parts = append(parts, a.City, a.State, a.PostalCode, a.Country)
	return strings.Join(parts, ", ")
}

// Coordinate is an immutable value object representing a geographic point.
type Coordinate struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// NewCoordinate creates a validated Coordinate value object.
func NewCoordinate(lat, lng float64) (Coordinate, error) {
	if lat < -90 || lat > 90 {
		return Coordinate{}, fmt.Errorf("latitude must be between -90 and 90, got %f", lat)
	}
	if lng < -180 || lng > 180 {
		return Coordinate{}, fmt.Errorf("longitude must be between -180 and 180, got %f", lng)
	}
	return Coordinate{Latitude: lat, Longitude: lng}, nil
}

// ToWKT returns the coordinate as Well-Known Text for PostGIS.
func (c Coordinate) ToWKT() string {
	return fmt.Sprintf("POINT(%f %f)", c.Longitude, c.Latitude)
}

// Equals checks equality with another Coordinate.
func (c Coordinate) Equals(other Coordinate) bool {
	return c.Latitude == other.Latitude && c.Longitude == other.Longitude
}

// GeoPoint wraps Coordinate for PostGIS compatibility.
type GeoPoint struct {
	Coordinate
}

// NewGeoPoint creates a new GeoPoint from latitude and longitude.
func NewGeoPoint(lat, lng float64) (GeoPoint, error) {
	coord, err := NewCoordinate(lat, lng)
	if err != nil {
		return GeoPoint{}, err
	}
	return GeoPoint{Coordinate: coord}, nil
}

// ToPostGISInsert returns the SQL expression for inserting into a PostGIS geometry column.
func (g GeoPoint) ToPostGISInsert() string {
	return fmt.Sprintf("ST_SetSRID(ST_MakePoint(%f, %f), 4326)", g.Longitude, g.Latitude)
}
