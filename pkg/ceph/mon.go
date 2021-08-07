package ceph

import (
	"encoding/json"
	"github.com/liucxer/ceph-tools/pkg/interfacer"
	"strings"
)

func GetAllIps(worker interfacer.Worker) ([]string, error) {
	var (
		ips []string
		err error
	)

	type Resp struct {
		MonMap struct {
			Mons []struct {
				PublicAddr string `json:"public_addr"`
			} `json:"mons"`
		} `json:"monmap"`
	}

	bts, err := worker.ExecCmd("ceph mon_status")
	if err != nil {
		return nil, err
	}

	var res Resp
	err = json.Unmarshal(bts, &res)
	if err != nil {
		return nil, err
	}

	for _, mon := range res.MonMap.Mons {
		ip := strings.Split(mon.PublicAddr, ":")[0]
		ips = append(ips, ip)
	}
	return ips, err
}
