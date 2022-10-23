.PHONY: all
all:
	rm -rf ./build
	mkdir ./build
	go build -o ./build/bland
	cp -r ./templates ./build
	cp -r ./sql	./build
	cp -r ./static ./build
