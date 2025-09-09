package snippet

import (
	"context"
	"database/sql"
	"regexp"
)

func Make(ctx context.Context, db *sql.DB, docID int, qTerms []string, maxLen int) string {
	var content string
	_ = db.QueryRowContext(ctx, `SELECT content FROM documents WHERE id=?`, docID).Scan(&content)
	if content == "" { return "" }
	if maxLen <= 0 { maxLen = 200 }

	pat := ""
	for i, t := range qTerms {
		if i > 0 { pat += "|" }
		pat += regexp.QuoteMeta(t)
	}
	if pat == "" { if len(content) > maxLen { return content[:maxLen] + "…" }; return content }

	re := regexp.MustCompile("(?i)(" + pat + ")")
	loc := re.FindStringIndex(content)
	if loc == nil {
		if len(content) > maxLen { return content[:maxLen] + "…" }
		return content
	}
	start := loc[0] - maxLen/3
	if start < 0 { start = 0 }
	end := start + maxLen
	if end > len(content) { end = len(content) }
	s := content[start:end]
	return s
}
