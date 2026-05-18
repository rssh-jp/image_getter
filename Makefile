BINARY  := bin/image_getter
CMD     := ./cmd/image_getter

URL     ?=
STORAGE ?= ./download
DEPTH   ?= 0

.PHONY: build test clean run

build:
	go build -o $(BINARY) $(CMD)

test:
	go test ./...

clean:
	rm -f $(BINARY)

run: build
	$(BINARY) -u $(URL) -s $(STORAGE) -d $(DEPTH)
