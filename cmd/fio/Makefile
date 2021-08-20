all: ceph_tools_fio_mac ceph_tools_fio_linux ceph_tools_fio_windows

ceph_tools_fio_mac: main.go
	mkdir -p build
	go build -o ./build/ceph_tools_fio_mac main.go

ceph_tools_fio_linux: main.go
	mkdir -p build && cd build
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./build/ceph_tools_fio_linux main.go

ceph_tools_fio_windows: main.go
	mkdir -p build && cd build
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o ./build/ceph_tools_fio_windows main.go

install: all
	scp ./build/ceph_tools_fio_linux root@10.0.20.27:/workspace/liuchangxi/fio/