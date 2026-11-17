# poisonprobe Makefile

BINARY_NAME=poisonprobe
GO=go

.PHONY: all build build-web test clean scan sanitize studio

all: build-web build

build-web:
	cd web && npm install && npm run build

build:
	$(GO) build -ldflags="-s -w" -o $(BINARY_NAME) ./cmd/poisonprobe

test:
	$(GO) test -v -race ./...

scan:
	./$(BINARY_NAME) scan examples/test_datasets/poisoned_sleeper_backdoor.jsonl

sanitize:
	./$(BINARY_NAME) sanitize examples/test_datasets/poisoned_sleeper_backdoor.jsonl -o clean.parquet -q quarantine.json -a audit.json

studio:
	./$(BINARY_NAME) studio examples/test_datasets/poisoned_sleeper_backdoor.jsonl --port 7474

clean:
	rm -f $(BINARY_NAME) $(BINARY_NAME).exe clean.parquet quarantine.json audit.json
