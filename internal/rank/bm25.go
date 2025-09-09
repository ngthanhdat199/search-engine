package rank

import "search-engine/internal/index"

type Score struct {
	DocID int
	Score float64
}

func BM25(ix *index.Engine, qTerms []string, k1, b float64, topK int) []Score {
	if len(qTerms) == 0 {
		return nil
	}
	scores := map[int]float64{}

	for _, t := range qTerms {
		pl := ix.Postings[t]
		if pl == nil {
			continue
		}
		idf := ix.IDF[t]
		for _, p := range pl {
			dl := float64(ix.DocLen[p.DocID])
			den := float64(p.TF) + k1*(1-b+b*dl/(ix.AvgDL+1e-9))
			scores[p.DocID] += idf * (float64(p.TF) * (k1 + 1)) / den
		}
	}
	// select topK (simple partial sort)
	arr := make([]Score, 0, len(scores))
	for id, s := range scores {
		arr = append(arr, Score{DocID: id, Score: s})
	}
	// naive sort (ok for small corpora)
	for i := 0; i < len(arr); i++ {
		for j := i + 1; j < len(arr); j++ {
			if arr[j].Score > arr[i].Score {
				arr[i], arr[j] = arr[j], arr[i]
			}
		}
	}
	if topK > 0 && topK < len(arr) {
		arr = arr[:topK]
	}
	return arr
}
