all: build
build:
	GOOS=windows GOARCH=amd64 go build  -o /Volumes/HomeNas/cli.exe ./cli/main.go

