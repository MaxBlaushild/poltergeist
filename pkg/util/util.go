package util

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const (
	flawless = "✨"
	great    = "🔥"
	good     = "👌"
	bad      = "😳"
	gross    = "🤮"
)

func MatchesGivePointPattern(s string) (bool, string) {
	r := regexp.MustCompile(`(?i)give (<@\w+>) a point`)
	matches := r.FindStringSubmatch(s)
	if len(matches) > 1 {
		return true, matches[1]
	}
	return false, ""
}

// GetShareMessage generates a message based on the correctness value.
func GetShareMessage(guess int, howMany int, correctness float64, explanation string) string {
	var emoji string

	switch {
	case correctness == 1:
		emoji = flawless
	case correctness > 0.9:
		emoji = great
	case correctness > 0.7:
		emoji = good
	case correctness > 0.4:
		emoji = bad
	default:
		emoji = gross
	}

	return fmt.Sprintf("You guessed %d. The answer was %d. You were %.2f%% correct! %s\n\n%s", guess, howMany, correctness*100, emoji, explanation)
}

func LongestNumericSubstring(s string) string {
	// Use a regular expression to match the desired pattern
	r := regexp.MustCompile(`[\d$,]+`)
	matches := r.FindAllString(s, -1)

	// Return the longest matching substring
	var longest string
	for _, match := range matches {
		if len(match) > len(longest) {
			longest = match
		}
	}

	return longest
}

func ParseNumber(s string) (int, error) {
	// Remove whitespaces and commas
	cleaned := strings.ReplaceAll(s, " ", "")
	cleaned = strings.ReplaceAll(cleaned, ",", "")
	cleaned = strings.ReplaceAll(cleaned, "$", "")

	// Attempt to convert cleaned string to int
	num, err := strconv.Atoi(cleaned)
	if err != nil {
		return 0, errors.New("unable to convert string to integer")
	}

	return num, nil
}
