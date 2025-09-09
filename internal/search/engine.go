package search

import (
	"fmt"
	"strings"

	"search-engine/internal/index"
	"search-engine/internal/rank"
	"search-engine/internal/text"

	util "search-engine/pkg/util"
)

type Item struct {
	DocID   int
	URL     string
	Title   string
	Score   float64
	Snippet string
}

func Query(ix *index.Engine, q string, topK int) (terms []string, results []Item) {
	fmt.Println("Query:", q)
	fmt.Println("TopK:", topK)
	qTerms := text.Tokenize(strings.TrimSpace(q))
	fmt.Println("Query terms:", qTerms)
	scores := rank.BM25(ix, qTerms, 1.5, 0.75, topK)
	fmt.Println("Scores:", scores)
	out := make([]Item, 0, len(scores))
	for _, s := range scores {
		out = append(out, Item{
			DocID: s.DocID,
			URL:   ix.URL[s.DocID],
			Title: ix.Title[s.DocID],
			Score: s.Score,
		})
	}

	fmt.Println("Results:", util.ToJSON(out))

	return qTerms, out
}
