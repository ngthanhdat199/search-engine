package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"search-engine/internal/store"

	"golang.org/x/net/html"
)

func main() {
	logger := log.New(os.Stdout, "[crawler] ", log.LstdFlags|log.Lshortfile)
	db, err := store.OpenSQLite("mini_search.db")
	if err != nil {
		logger.Fatal(err)
	}
	if err := store.RunMigrations(db, "migrations/0001_init.sql"); err != nil {
		logger.Fatal(err)
	}

	// Seed a few safe docs (change these to sites you’re allowed to crawl)
	seeds := []string{
		"https://go.dev/blog/declaration-syntax",
		"https://go.dev/blog/context",
		"https://kubernetes.io/docs/concepts/overview/what-is-kubernetes/",
	}

	client := &http.Client{Timeout: 10 * time.Second}
	ctx := context.Background()

	for _, u := range seeds {
		logger.Printf("fetch %s", u)
		req, _ := http.NewRequestWithContext(ctx, "GET", u, nil)
		req.Header.Set("User-Agent", "search-engine-crawler/0.1 (+learn)")
		resp, err := client.Do(req)
		if err != nil {
			logger.Printf("skip %s: %v", u, err)
			continue
		}
		if resp.Body != nil {
			defer resp.Body.Close()
		}
		if resp.StatusCode != 200 {
			logger.Printf("skip %s: %s", u, resp.Status)
			continue
		}
		b, _ := io.ReadAll(resp.Body)
		title, text := extract(string(b))
		if len(strings.TrimSpace(text)) < 120 {
			logger.Printf("too short: %s", u)
			continue
		}
		if err := store.UpsertDocument(ctx, db, u, title, text); err != nil {
			logger.Printf("upsert %s: %v", u, err)
		} else {
			fmt.Println("OK:", u, "title:", title)
		}
		time.Sleep(500 * time.Millisecond) // simple politeness
	}
}

func extract(htmlStr string) (string, string) {
	doc, err := html.Parse(strings.NewReader(htmlStr))
	if err != nil {
		return "", ""
	}
	var title string
	var b strings.Builder

	var f func(*html.Node, bool)
	f = func(n *html.Node, visible bool) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "script", "style", "noscript", "template":
				visible = false
			case "title":
				if n.FirstChild != nil {
					title = strings.TrimSpace(n.FirstChild.Data)
				}
			}
		}
		if n.Type == html.TextNode && visible {
			txt := strings.TrimSpace(n.Data)
			if txt != "" {
				if b.Len() > 0 {
					b.WriteByte(' ')
				}
				b.WriteString(txt)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c, visible)
		}
	}
	f(doc, true)
	return title, b.String()
}
