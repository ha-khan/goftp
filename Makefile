container_work_dir = /go/src/app

run:
	docker-compose up

build-dev:
	WORK_DIR=$(container_work_dir) docker build -f Dockerfile.dev . --tag dev

build:
	WORK_DIR=$(container_work_dir) docker build . --tag goftp/prod

build-binary:
	CGO_ENABLED=0 go build -o bin/main ./cmd/goftp/main.go

clean:
	rm -f ./bin/*
