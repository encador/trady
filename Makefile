.PHONY: run dev clean

clean:
	- rm trady

build:
	go build -o trady

run:
	go run .
