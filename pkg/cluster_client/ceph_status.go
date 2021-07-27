package cluster_client

import (
	"encoding/json"
)

type CephStatus struct {
	ReadBytesSec          float64 `json:"read_bytes_sec"`
	ReadOpPerSec          float64 `json:"read_op_per_sec"`
	RecoveringBytesPerSec float64 `json:"recovering_bytes_per_sec"`
	WriteBytesSec         float64 `json:"write_bytes_sec"`
	WriteOpPerSec         float64 `json:"write_op_per_sec"`
}

type CephStatusList []CephStatus

func (l CephStatusList) ReadBytesSec() float64 {
	res := float64(0)
	for _, v := range l {
		res = res + v.ReadBytesSec
	}
	return res / float64(len(l))
}

func (l CephStatusList) ReadOpPerSec() float64 {
	res := float64(0)
	for _, v := range l {
		res = res + v.ReadOpPerSec
	}
	return res / float64(len(l))
}

func (l CephStatusList) RecoveringBytesPerSec() float64 {
	res := float64(0)
	for _, v := range l {
		res = res + v.RecoveringBytesPerSec
	}
	return res / float64(len(l))
}

func (l CephStatusList) WriteBytesSec() float64 {
	res := float64(0)
	for _, v := range l {
		res = res + v.WriteBytesSec
	}
	return res / float64(len(l))
}

func (l CephStatusList) WriteOpPerSec() float64 {
	res := float64(0)
	for _, v := range l {
		res = res + v.WriteOpPerSec
	}
	return res / float64(len(l))
}

func (cluster *Cluster) CephStatus(second int) (*CephStatusList, error) {
	type CephStatusResp struct {
		PGMap CephStatus `json:"pgmap"`
	}

	var list CephStatusList
	for i := 0; i < second; i++ {
		var res CephStatus
		resp, err := cluster.Clients[0].ExecCmd("ceph status -f json")
		if err != nil {
			return nil, err
		}

		var cephStatusResp CephStatusResp
		err = json.Unmarshal(resp, &cephStatusResp)
		if err != nil {
			return nil, err
		}

		res = cephStatusResp.PGMap
		list = append(list, res)
	}

	return &list, nil
}
