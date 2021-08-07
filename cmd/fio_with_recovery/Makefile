all: ceph-tools-fio-with-recovery-mac ceph-tools-fio-with-recovery-linux ceph-tools-fio-with-recovery-windows

ceph-tools-fio-with-recovery-mac: main.go
	mkdir -p build
	go build -o ./build/ceph-tools-fio-with-recovery-mac main.go

ceph-tools-fio-with-recovery-linux: main.go
	mkdir -p build && cd build
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./build/ceph-tools-fio-with-recovery-linux main.go

ceph-tools-fio-with-recovery-windows: main.go
	mkdir -p build && cd build
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o ./build/ceph-tools-fio-with-recovery-windows main.go