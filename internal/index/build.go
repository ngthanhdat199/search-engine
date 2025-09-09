package index

import (
	"database/sql"
	"math"

	"search-engine/internal/text"
)

type Posting struct {
	DocID int
	TF    int
}

type Engine struct {
	Postings map[string][]Posting   // term -> postings
	DocLen   map[int]int            // docID -> token count
	Title    map[int]string
	URL      map[int]string
	IDF      map[string]float64
	N        int
	AvgDL    float64
}

func Build(db *sql.DB) (*Engine, error) {
	rows, err := db.Query(`SELECT id, url, title, content FROM documents`)
	if err != nil { return nil, err }
	defer rows.Close()

	e := &Engine{
		Postings: map[string][]Posting{},
		DocLen:   map[int]int{},
		Title:    map[int]string{},
		URL:      map[int]string{},
		IDF:      map[string]float64{},
	}

	tfMap := map[string]map[int]int{}
	for rows.Next() {
		var id int
		var url, title, content string
		if err := rows.Scan(&id, &url, &title, &content); err != nil { return nil, err }

		toks := text.Tokenize(title + " " + content)
		e.DocLen[id] = len(toks)
		e.Title[id] = title
		e.URL[id] = url
		if len(toks) == 0 { continue }

		if tfMap == nil { tfMap = map[string]map[int]int{} }
		for _, t := range toks {
			m := tfMap[t]
			if m == nil { m = map[int]int{}; tfMap[t] = m }
			m[id]++
		}
		e.N++
	}
	// avgdl
	sum := 0
	for _, l := range e.DocLen { sum += l }
	if e.N > 0 { e.AvgDL = float64(sum)/float64(e.N) }

	// finalize postings + idf
	for term, m := range tfMap {
		df := len(m)
		idf := math.Log(1 + (float64(e.N)-float64(df)+0.5)/(float64(df)+0.5))
		e.IDF[term] = idf

		pl := make([]Posting, 0, df)
		for docID, tf := range m {
			pl = append(pl, Posting{DocID: docID, TF: tf})
		}
		e.Postings[term] = pl
	}
	return e, nil
}
