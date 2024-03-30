package util

import (
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"
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

// GenerateRandomName generates a random two-word name similar to the naming convention used by Heroku for their servers.
// It combines a random adjective and a random noun with a hyphen in between.
func GenerateRandomName() string {
	adjectives := []string{"autumn", "hidden", "bitter", "misty", "silent", "empty", "dry", "dark", "summer", "icy", "delicate", "quiet", "white", "cool", "spring", "winter", "patient", "twilight", "dawn", "crimson", "wispy", "weathered", "blue", "billowing", "broken", "cold", "damp", "falling", "frosty", "green", "long", "late", "lingering", "bold", "little", "morning", "muddy", "old", "red", "rough", "still", "small", "sparkling", "throbbing", "shy", "wandering", "withered", "wild", "black", "young", "holy", "solitary", "fragrant", "aged", "snowy", "proud", "floral", "restless", "divine", "polished", "ancient", "purple", "lively", "nameless"}
	nouns := []string{"waterfall", "river", "breeze", "moon", "rain", "wind", "sea", "morning", "snow", "lake", "sunset", "pine", "shadow", "leaf", "dawn", "glitter", "forest", "hill", "cloud", "meadow", "sun", "glade", "bird", "brook", "butterfly", "bush", "dew", "dust", "field", "fire", "flower", "firefly", "feather", "grass", "haze", "mountain", "night", "pond", "darkness", "snowflake", "silence", "sound", "sky", "shape", "surf", "thunder", "violet", "water", "wildflower", "wave", "water", "resonance", "sun", "wood", "dream", "cherry", "tree", "fog", "frost", "voice", "paper", "frog", "smoke", "star"}

	rand.Seed(time.Now().UnixNano())
	return adjectives[rand.Intn(len(adjectives))] + "-" + nouns[rand.Intn(len(nouns))]
}
