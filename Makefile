.PHONY: tidy run crawl api

tidy:
	go mod tidy

crawl:
	go run ./cmd/crawler

api:
	go run ./cmd/api

run: crawl api
