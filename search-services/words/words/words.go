package words

import (
	"log/slog"
	"maps"
	"slices"
	"strings"
	"unicode"

	"github.com/kljensen/snowball"
	snowballEnglish "github.com/kljensen/snowball/english"
)

const averageRequestWords = 4

func Norm(phrase string) []string {

	words := strings.FieldsFunc(phrase, func(r rune) bool {
		return !unicode.IsDigit(r) && !unicode.IsLetter(r)
	})

	seen := make(map[string]struct{}, averageRequestWords)

	for _, word := range words {

		word = strings.ToLower(word)

		if snowballEnglish.IsStopWord(word) {
			continue
		}

		stemmed, err := snowball.Stem(word, "english", true)
		if err != nil {
			slog.Error("Error stemming word:", "word", word, "error", err)
			continue
		}

		if _, ok := seen[stemmed]; ok {
			continue
		}

		seen[stemmed] = struct{}{}
	}

	return slices.Collect(maps.Keys(seen))
}
