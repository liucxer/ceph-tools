all: ceph_tools_fio_mac ceph_tools_fio_linux ceph_tools_fio_windows

ceph_tools_fio_mac: main.go
	mkdir _p build
	go build _o ./build/ceph_tools_fio_mac main.go

ceph_tools_fio_linux: main.go
	mkdir _p build && cd build
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build _o ./build/ceph_tools_fio_linux main.go

ceph_tools_fio_windows: main.go
	mkdir _p build && cd build
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build _o ./build/ceph_tools_fio_windows main.go