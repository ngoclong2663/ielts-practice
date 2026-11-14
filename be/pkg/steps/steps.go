// Package steps parses Anki-style (re)learning steps such as "1m 10m" or "10m 1h 1d".
package steps

import (
	"errors"
	"strconv"
	"strings"
)

var ErrInvalidSteps = errors.New("invalid steps")

const MaxSteps = 6

// ParseMinutes converts "1m 10m 1h" to minutes. Units: m (minutes), h (hours), d (days).
// An empty string means no steps.
func ParseMinutes(s string) ([]float64, error) {
	fields := strings.Fields(s)
	if len(fields) > MaxSteps {
		return nil, ErrInvalidSteps
	}
	out := make([]float64, 0, len(fields))
	for _, f := range fields {
		if len(f) < 2 {
			return nil, ErrInvalidSteps
		}
		n, err := strconv.ParseFloat(f[:len(f)-1], 64)
		if err != nil || n <= 0 || n > 10000 {
			return nil, ErrInvalidSteps
		}
		switch f[len(f)-1] {
		case 'm':
		case 'h':
			n *= 60
		case 'd':
			n *= 1440
		default:
			return nil, ErrInvalidSteps
		}
		out = append(out, n)
	}
	return out, nil
}
