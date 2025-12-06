.PHONY: build-cpp build-go run-docker test clean

build-cpp:
	cd crypto_cpp && mkdir -p build && cd build && cmake .. && make

build-go:
	go build -o bin/server ./cmd/server
	go build -o bin/worker ./cmd/worker

run-docker:
	docker compose -f infra/docker-compose.yml up --build

test:
	go test -v ./...

clean:
	rm -rf bin
	rm -rf crypto_cpp/build
