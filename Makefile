.PHONY: build run reindex clean

build:
	go build -o curius-search ./cmd/curius-search

run: build
	./curius-search

reindex: build
	./curius-search --reindex

index-only: build
	./curius-search --index-only

clean:
	rm -f curius-search
	rm -f data/index.json
