package main

import (
	"encoding/json"
	"fmt"
	"github.com/liucxer/ceph-tools/pkg/host_client"
	"os"
)

type CephStatus struct {
	ReadBytesSec          int64 `json:"read_bytes_sec"`
	ReadOpPerSec          int64 `json:"read_op_per_sec"`
	RecoveringBytesPerSec int64 `json:"recovering_bytes_per_sec"`
	WriteBytesSec         int64 `json:"write_bytes_sec"`
	WriteOpPerSec         int64 `json:"write_op_per_sec"`
}

func GetCephStatus(ipAddr string) (CephStatus, error) {
	var (
		res CephStatus
		err error
	)

	type CephStatusResp struct {
		PGMap CephStatus `json:"pgmap"`
	}

	client, err := host_client.NewHostClient(ipAddr)
	if err != nil {
		return res, err
	}

	resp, err := client.ExecCmd("ceph status -f json")
	if err != nil {
		return res, err
	}

	var cephStatusResp CephStatusResp
	err = json.Unmarshal(resp, &cephStatusResp)
	if err != nil {
		return res, err
	}

	res = cephStatusResp.PGMap
	return res, nil
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage:\n     ./cmd ipaddr")
		return
	}

	for {
		cephStatus, err := GetCephStatus(os.Args[1])
		if err != nil {
			return
		}
		fmt.Println(cephStatus)
	}
}
