BINARY := spotui

.PHONY: build run clean

build:
	go build -o $(BINARY) ./main.go

run:
	go run ./main.go

clean:
	rm -f $(BINARY)
