docker-build:
	docker build --no-cache -t personalisation-poc .

start:
	docker compose up --force-recreate --no-deps -d

stop:
	docker compose down
	docker rmi -f personalisation-poc:latest 

build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o main .

restart: stop start

test:
	go test -v ./...
