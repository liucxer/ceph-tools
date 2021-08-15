all: auto_change_limit_mac auto_change_limit_linux auto_change_limit_windows

auto_change_limit_mac: main.go
	mkdir -p build
	go build -o ./build/auto_change_limit_mac main.go

auto_change_limit_linux: main.go
	mkdir -p build && cd build
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./build/auto_change_limit_linux main.go

auto_change_limit_windows: main.go
	mkdir -p build && cd build
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o ./build/auto_change_limit_windows main.go

install: all
	scp ./build/auto_change_limit_linux root@10.0.20.27:/home/liucx/auto_change_limit/