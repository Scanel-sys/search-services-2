package words

import (
	"log/slog"
	"regexp"
	"strings"

	"github.com/kljensen/snowball"
)

const averageRequestWords = 4

var wordRegexp = regexp.MustCompile(`[[:alnum:]]+`)

var stopWords = map[string]struct{}{
	"a": {}, "about": {}, "above": {}, "after": {}, "again": {}, "against": {}, "all": {},
	"am": {}, "an": {}, "and": {}, "any": {}, "are": {}, "as": {}, "at": {},

	"be": {}, "because": {}, "been": {}, "before": {}, "being": {}, "below": {}, "between": {},
	"both": {}, "but": {}, "by": {},

	"can": {}, "could": {}, "couldn": {},

	"did": {}, "didn": {}, "do": {}, "does": {}, "doesn": {}, "doing": {}, "don": {}, "down": {},
	"during": {},

	"each": {},

	"few": {}, "for": {}, "from": {}, "further": {},

	"had": {}, "hadn": {}, "has": {}, "hasn": {}, "have": {}, "haven": {}, "having": {},
	"he": {}, "her": {}, "here": {}, "hers": {}, "herself": {},
	"him": {}, "himself": {}, "his": {}, "how": {},

	"i": {}, "if": {}, "in": {}, "into": {}, "is": {}, "isn": {}, "it": {}, "its": {}, "itself": {},

	"just": {},

	"ma": {}, "me": {}, "more": {}, "most": {}, "mustn": {}, "my": {}, "myself": {},

	"no": {}, "nor": {}, "not": {}, "now": {},

	"of": {}, "off": {}, "on": {}, "once": {}, "only": {}, "or": {}, "other": {},
	"our": {}, "ours": {}, "ourselves": {}, "out": {}, "over": {}, "own": {},

	"s": {}, "same": {}, "shan": {}, "she": {}, "should": {}, "shouldn": {}, "so": {}, "some": {}, "such": {},

	"t": {}, "than": {}, "that": {}, "the": {}, "their": {}, "theirs": {}, "them": {},
	"themselves": {}, "then": {}, "there": {}, "these": {}, "they": {}, "this": {}, "those": {},
	"through": {}, "to": {}, "too": {},

	"under": {}, "until": {}, "up": {},

	"very": {},

	"was": {}, "wasn": {}, "we": {}, "were": {}, "weren": {}, "what": {}, "when": {},
	"where": {}, "which": {}, "while": {}, "who": {}, "whom": {}, "why": {}, "will": {},
	"with": {}, "won": {}, "would": {}, "wouldn": {},

	"you": {}, "your": {}, "yours": {}, "yourself": {}, "yourselves": {},
}

func Norm(phrase string) []string {

	words := wordRegexp.FindAllString(phrase, -1)

	result := make([]string, 0, averageRequestWords)
	seen := make(map[string]struct{}, averageRequestWords)

	for _, word := range words {

		word = strings.ToLower(word)

		if _, ok := stopWords[word]; ok {
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
		result = append(result, stemmed)
	}

	return result
}
