package datatypes

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// Datatypes implements a number of support functions to assist in
// validation and conversion of various specific types of data used
// in the geography database

// ISOCountryCode converts a string into a valid ISO Country Code
func ISOCountryCode(s string, partial bool, empty bool) (string, error) {
	var result strings.Builder

	// Clean the string and validate content
	for _, c := range s {
		if !unicode.IsSpace(c) {
			if !unicode.IsLetter(c) {
				return "", fmt.Errorf("Invalid ISO Country Code")
			}
			result.WriteRune(unicode.ToUpper(c))
		}
	}
	// Empty
	if result.Len() == 0 && !empty {
		return "", fmt.Errorf("Invalid ISO Country Code")
	}
	// Too short
	if result.Len() > 0 && result.Len() < 2 && !partial {
		return "", fmt.Errorf("Invalid ISO Country Code")
	}
	// Too long
	if result.Len() > 2 {
		return "", fmt.Errorf("Invalid ISO Country Code")
	}
	return result.String(), nil
}

// ISORegionCode converts a string into a valid ISO Region Code
func ISORegionCode(s string, partial bool, empty bool) (string, error) {
	var result strings.Builder

	// Clean the string
	for _, c := range s {
		if !unicode.IsSpace(c) {
			if !unicode.In(c, unicode.Letter, unicode.Digit, unicode.Pc, unicode.Pd) {
				return "", fmt.Errorf("Invalid ISO Region Code")
			}
			result.WriteRune(unicode.ToUpper(c))
		}
	}
	// Empty
	if result.Len() == 0 && !empty {
		return "", fmt.Errorf("Invalid ISO Region Code")
	}
	// I don't know the length ranges of an ISO Region Code
	// so it is untested for now (just empty is no good)
	return result.String(), nil
}

// ICAOAirportCode converts a string into a valid ICAO Airport Code
func ICAOAirportCode(s string, partial bool, empty bool) (string, error) {
	var result strings.Builder
	// Clean the string
	for _, c := range s {
		if !unicode.IsSpace(c) {
			if !unicode.IsDigit(c) && !unicode.IsLetter(c) {
				return "", fmt.Errorf("Invalid ICAO Airport Code")
			}
			result.WriteRune(unicode.ToUpper(c))
		}
	}
	// Empty
	if result.Len() == 0 && !empty {
		return "", fmt.Errorf("Invalid ICAO Airport Code")
	}
	// Short
	if result.Len() > 0 && result.Len() < 2 && !partial {
		return "", fmt.Errorf("Invalid ICAO Airport Code")
	}
	// Long
	if result.Len() > 4 {
		return "", fmt.Errorf("Invalid ICAO Airport Code")
	}
	return result.String(), nil
}

// IATAAirportCode converts a string into a valid IATA Airport Code
func IATAAirportCode(s string, partial bool, empty bool) (string, error) {
	var result strings.Builder
	// Clean the string
	for _, c := range s {
		if !unicode.IsSpace(c) {
			if !unicode.IsLetter(c) {
				return "", fmt.Errorf("Invalid IATA Airport Code")
			}
			result.WriteRune(unicode.ToUpper(c))
		}
	}
	// Empty
	if result.Len() == 0 && !empty {
		return "", fmt.Errorf("Invalid IATA Airport Code")
	}
	// Short
	if result.Len() > 0 && result.Len() < 3 && !partial {
		return "", fmt.Errorf("Invalid IATA Airport Code")
	}
	// Long
	if result.Len() > 3 {
		return "", fmt.Errorf("Invalid IATA Airport Code")
	}
	return result.String(), nil
}

// RunwayCode converts a string to a valid RunwayCode
func RunwayCode(s string, partial bool, empty bool) (string, error) {
	var result strings.Builder
	// Clean the string
	for _, c := range s {
		if !unicode.IsSpace(c) {
			if !unicode.In(c, unicode.Letter, unicode.Digit, unicode.Pc, unicode.Pd) {
				return "", fmt.Errorf("Invalid Runway Code")
			}
			result.WriteRune(unicode.ToUpper(c))
		}
	}
	// Empty
	if len(s) == 0 && !empty {
		return "", fmt.Errorf("Invalid Runway Code")
	}
	// No maximum length known, so we also don't know how to define 'partial'
	return result.String(), nil
}

// RunwayLength converts a string to a valid Runway Length in feet
func RunwayLength(s string, empty bool) (int, error) {
	text := strings.TrimSpace(s)
	if len(text) == 0 {
		if !empty {
			return 0, fmt.Errorf("Invalid Runway Length")
		}
		return 0, nil
	}

	value, err := strconv.ParseFloat(text, 32)
	if err != nil {
		return 0, fmt.Errorf("Invalid Runway Length")
	}

	// Check between 1ft and 30000ft (roughly 10KM)
	length := int(value)
	if length <= 0 || (length == 0 && !empty) || length > 30000 {
		return 0, fmt.Errorf("Invalid Runway Length")
	}
	return length, nil
}

// RunwayWidth converts a string to a valid Runway Width in feet
func RunwayWidth(s string, empty bool) (int, error) {
	// Clean up text
	text := strings.TrimSpace(s)
	if len(text) == 0 {
		if !empty {
			return 0, fmt.Errorf("Invalid Runway Width")
		}
		return 0, nil
	}

	// Extract number
	value, err := strconv.ParseFloat(text, 32)
	if err != nil {
		return 0, fmt.Errorf("Invalid Runway Width")
	}

	// Check between 0 and 30000ft (roughly 10KM)
	width := int(value)
	if width < 0 || (width == 0 && !empty) || width > 30000 {
		return 0, fmt.Errorf("Invalid Runway Width")
	}
	return width, nil
}

// RunwayLighted converts a string to a valid Lighed flag
func RunwayLighted(s string, empty bool) (bool, error) {
	// Clean up string
	text := strings.TrimSpace(s)
	if len(text) == 0 {
		if !empty {
			return false, fmt.Errorf("Invalid Runway Lighted")
		}
		return false, nil
	}

	// Extract literal value
	switch strings.TrimSpace(s) {
	case "1":
		return true, nil
	case "0":
		return false, nil
	}

	return false, fmt.Errorf("Invalid Runway Lighted")
}

// RunwayClosed converts a string to a valid Lighed flag
func RunwayClosed(s string, empty bool) (bool, error) {
	// Clean up string
	text := strings.TrimSpace(s)
	if len(text) == 0 {
		if !empty {
			return false, fmt.Errorf("Invalid Runway Closed")
		}
		return false, nil
	}

	// Extract literal value
	switch strings.TrimSpace(s) {
	case "1":
		return true, nil
	case "0":
		return false, nil
	}

	return false, fmt.Errorf("Invalid Runway Closed")
}

// RunwayHeading converts a string to a valid Runway Heading in degrees
func RunwayHeading(s string, empty bool) (int, error) {
	// Clean up string
	text := strings.TrimSpace(s)
	if len(text) == 0 {
		if !empty {
			return 0, fmt.Errorf("Invalid Runway Heading")
		}
		return 0, nil
	}

	// Extract number
	value, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return 0, fmt.Errorf("Invalid Runway Heading")
	}

	// Heading must be between 0 and 360 inclusive
	heading := int(value)
	if heading < 0 || heading > 360 {
		return 0, fmt.Errorf("Invalid Runway Heading")
	}

	return heading, nil
}

// RunwayThreshold converts a string to a valid Runway Threshold in feet
func RunwayThreshold(s string, empty bool) (int, error) {
	// Clean up string
	text := strings.TrimSpace(s)
	if len(text) == 0 {
		if !empty {
			return 0, fmt.Errorf("Invalid Runway Heading")
		}
		return 0, nil
	}

	// Extract number
	value, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return 0, fmt.Errorf("Invalid Runway Threshold")
	}

	// Must be between 0 and 30000ft (10KM)
	threshold := int(value)
	if threshold < 0 || threshold > 30000 {
		return 0, fmt.Errorf("Invalid Runway Threshold")
	}

	return threshold, nil
}

// Latitude converts a string to a valid latitude
func Latitude(s string, empty bool) (float64, error) {
	// Clean up string
	text := strings.TrimSpace(s)
	if len(text) == 0 {
		if !empty {
			return 0.0, fmt.Errorf("Invalid Latitude")
		}
		return 0.0, nil
	}

	// Extract number
	latitude, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return 0.0, fmt.Errorf("Invalid Latitude")
	}

	// Must be between -90deg and +90deg
	if latitude < -90.0 || latitude > 90.0 {
		return 0.0, fmt.Errorf("Invalid Latitude")
	}

	return latitude, nil
}

// Longitude converts a string to a valid longitude
func Longitude(s string, empty bool) (float64, error) {
	// Clean up string
	text := strings.TrimSpace(s)
	if len(text) == 0 {
		if !empty {
			return 0.0, fmt.Errorf("Invalid Longitude")
		}
		return 0.0, nil
	}

	// Extract number
	longitude, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return 0.0, fmt.Errorf("Invalid Longitude")
	}

	// Must be between -180deg and +180deg
	if longitude < -180.0 || longitude > 180.0 {
		return 0.0, fmt.Errorf("Invalid Longitude")
	}

	return longitude, nil
}

// Elevation converts a string to a valid elevation
func Elevation(s string, empty bool) (int, error) {
	// Clean up string
	text := strings.TrimSpace(s)
	if len(text) == 0 {
		if !empty {
			return 0, fmt.Errorf("Invalid Elevation")
		}
		return 0, nil
	}

	// Extract number
	value, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return 0, fmt.Errorf("Invalid Elevation")
	}

	// Must be between -45000 (15KM deep)  and 30000ft (10KM high)
	elevation := int(value)
	if elevation < -45000 || elevation > 30000 {
		return 0, fmt.Errorf("Invalid Elevatino")
	}

	return elevation, nil
}

// Frequency translates a string to a valid civil aircraft frequency
func Frequency(s string, empty bool) (float64, error) {
	// Clean up string
	text := strings.TrimSpace(s)
	if len(text) == 0 {
		if !empty {
			return 0.0, fmt.Errorf("Invalid Frequency")
		}
		return 0.0, nil
	}

	// Extract number
	frequency, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return 0.0, fmt.Errorf("Invalid Frequency")
	}

	// Must be between 118.0 and 137.0MHz (Civil aircraft communications band)
	if frequency < 118.0 || frequency > 137.0 {
		return 0.0, fmt.Errorf("Invalid Frequency")
	}

	return frequency, nil

}
