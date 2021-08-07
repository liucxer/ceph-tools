all: ceph-tools-fio-mac ceph-tools-fio-linux ceph-tools-fio-windows

ceph-tools-fio-mac: main.go
	mkdir -p build
	go build -o ./build/ceph-tools-fio-mac main.go

ceph-tools-fio-linux: main.go
	mkdir -p build && cd build
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./build/ceph-tools-fio-linux main.go

ceph-tools-fio-windows: main.go
	mkdir -p build && cd build
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o ./build/ceph-tools-fio-windows main.go