.PHONY: run dev clean

clean:
	- rm -rf bin

build:
	go build -o bin/trady

run: build
	./bin/trady -db-path=bin/trady.db
