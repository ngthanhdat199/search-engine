package index

import (
	"database/sql"
	"fmt"
	"math"

	"search-engine/internal/text"
	util "search-engine/pkg/util"
)

type Posting struct {
	DocID int
	TF    int
}

type Engine struct {
	Postings map[string][]Posting // term -> postings
	DocLen   map[int]int          // docID -> token count
	Title    map[int]string
	URL      map[int]string
	IDF      map[string]float64
	N        int
	AvgDL    float64
}

func Build(db *sql.DB) (*Engine, error) {
	rows, err := db.Query(`SELECT id, url, title, content FROM documents`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	e := &Engine{
		Postings: map[string][]Posting{},
		DocLen:   map[int]int{},
		Title:    map[int]string{},
		URL:      map[int]string{},
		IDF:      map[string]float64{},
	}

	tfMap := map[string]map[int]int{}

	fmt.Println("Starting indexing...", util.ToJSON(e))

	for rows.Next() {
		var id int
		var url, title, content string
		if err := rows.Scan(&id, &url, &title, &content); err != nil {
			return nil, err
		}

		// fmt.Println("Indexing title:", title, "content len:", content)
		toks := text.Tokenize(title + " " + content)
		// fmt.Println("Tokens:", toks)
		// fmt.Println("len toks:", len(toks))
		e.DocLen[id] = len(toks)
		e.Title[id] = title
		e.URL[id] = url
		if len(toks) == 0 {
			continue
		}

		for _, t := range toks {
			m := tfMap[t]
			if m == nil {
				m = map[int]int{}
				tfMap[t] = m
			}
			m[id]++
		}
		e.N++
	}

	fmt.Println("Total docs indexed:", util.ToJSON(e))

	// avgdl
	sum := 0
	for _, l := range e.DocLen {
		sum += l
	}
	if e.N > 0 {
		e.AvgDL = float64(sum) / float64(e.N)
	}

	fmt.Println("AvgDL:", util.ToJSON(e))
	// fmt.Println("Term map:", util.ToJSON(tfMap))

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

	fmt.Println("Final index:", util.ToJSON(e))

	return e, nil
}
