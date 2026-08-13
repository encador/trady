.PHONY: run dev clean build

clean:
	- rm -rf bin

build:
	go tool templ generate
	go build -o bin/trady

run: build
	./bin/trady -db=bin/trady.db -uploads=bin/uploads

dev:
	go tool templ generate --watch \
		--cmd="go run . -db=bin/trady.db -uploads=bin/uploads -port=55000" \
		--proxy="http://localhost:55000" \
		--proxybind="localhost" --proxyport="8080" \
		--open-browser=false

RUNNER ?= docker
docker:
	$(RUNNER) build -t trady .
	- $(RUNNER) stop trady-app
	- $(RUNNER) rm trady-app
	$(RUNNER) run -d --name trady-app -p 8080:8080 trady
