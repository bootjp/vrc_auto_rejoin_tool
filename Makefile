all: build
build:
	GOOS=windows GOARCH=amd64 go build -o ./dist/vrc_auto_rejoin_tool_x64.exe ./cli/

