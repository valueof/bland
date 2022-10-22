all: build

dev:
	go run . --addr localhost:9999 --dev --db ./bland.db

build:
	go build .