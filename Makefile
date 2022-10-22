all: build

dev:
	go run . --addr localhost:9999 --dev --db ./bland.db

prod:
	go run . --addr localhost:9999

build:
	go build .