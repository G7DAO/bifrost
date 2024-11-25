.PHONY: clean build rebuild

build: bin/bifrost

rebuild: clean build

bin/bifrost:
	mkdir -p bin
	go mod tidy
	go build -o bin/bifrost ./cmd/bifrost

clean:
	rm -rf bin/*