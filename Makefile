.PHONY: run dev clean build

clean:
	- rm -rf bin

build:
	go tool templ generate
	go build -o bin/trady

run: build
	./bin/trady -db-path=bin/trady.db

dev:
	go tool templ generate --watch \
		--cmd="go run . -db-path=bin/trady.db" \
		--proxy="http://localhost:55000" \
		--proxybind="localhost" --proxyport="8080" \
		--open-browser=false
