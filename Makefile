BINARY := spotui
MODULE := github.com/dcbto/spotui
LDFLAGS := -ldflags "-X $(MODULE)/internal/clientid.Value=$(SPOTIFY_CLIENT_ID)"

.PHONY: build run clean

build:
	go build $(LDFLAGS) -o $(BINARY) ./main.go

run:
	go run ./main.go

clean:
	rm -f $(BINARY)
