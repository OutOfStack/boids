build:
	mkdir -p bin
	go build -o bin/boids main.go

run:
	go run .