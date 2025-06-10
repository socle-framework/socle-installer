## build: builds the command line tool dist directory
build:
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./socle-installer ./cmd/cli 