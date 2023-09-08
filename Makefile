build:
	go build -ldflags "-s -w" -o ./bin/depviz ./cmd/depviz/main.go
	GOOS=darwin  GOARCH=arm64 go build -ldflags "-s -w" -o ./bin/depviz-darwin-arm64 ./cmd/depviz/main.go
	GOOS=darwin  GOARCH=amd64 go build -ldflags "-s -w" -o ./bin/depviz-darwin-amd64 ./cmd/depviz/main.go
	GOOS=linux   GOARCH=amd64 go build -ldflags "-s -w" -o ./bin/depviz-linux-amd64  ./cmd/depviz/main.go
	GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o ./bin/depviz-win-amd64    ./cmd/depviz/main.go

test:
	go test -race -v ./...

coverage:
	go test -race -coverprofile coverage.out ./...
	go tool cover -html coverage.out -o coverage.html

docker-build:
	docker build -t burenotti/depviz:latest .
