all: auto_change_limit-mac auto_change_limit-linux auto_change_limit-windows

auto_change_limit-mac: main.go
	mkdir -p build
	go build -o ./build/auto_change_limit-mac main.go

auto_change_limit-linux: main.go
	mkdir -p build && cd build
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./build/auto_change_limit-linux main.go

auto_change_limit-windows: main.go
	mkdir -p build && cd build
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o ./build/auto_change_limit-windows main.go