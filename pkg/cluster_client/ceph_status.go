package cluster_client

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

type CephStatus struct {
	ReadBytesSec          float64 `json:"read_bytes_sec"`
	ReadOpPerSec          float64 `json:"read_op_per_sec"`
	RecoveringBytesPerSec float64 `json:"recovering_bytes_per_sec"`
	WriteBytesSec         float64 `json:"write_bytes_sec"`
	WriteOpPerSec         float64 `json:"write_op_per_sec"`
}

type CephStatusList []CephStatus

func (l CephStatusList) AvgCephStatus() CephStatus {
	var res CephStatus
	res.ReadBytesSec, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", l.AvgReadBytesSec()), 64)
	res.ReadOpPerSec, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", l.AvgReadOpPerSec()), 64)
	res.RecoveringBytesPerSec, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", l.AvgRecoveringBytesPerSec()), 64)
	res.WriteBytesSec, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", l.AvgWriteBytesSec()), 64)
	res.WriteOpPerSec, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", l.AvgWriteOpPerSec()), 64)
	return res
}

func (l CephStatusList) AvgReadBytesSec() float64 {
	res := float64(0)
	for _, v := range l {
		res = res + v.ReadBytesSec
	}
	return res / float64(len(l))
}

func (l CephStatusList) AvgReadOpPerSec() float64 {
	res := float64(0)
	for _, v := range l {
		res = res + v.ReadOpPerSec
	}
	return res / float64(len(l))
}

func (l CephStatusList) AvgRecoveringBytesPerSec() float64 {
	res := float64(0)
	for _, v := range l {
		res = res + v.RecoveringBytesPerSec
	}
	return res / float64(len(l))
}

func (l CephStatusList) AvgWriteBytesSec() float64 {
	res := float64(0)
	for _, v := range l {
		res = res + v.WriteBytesSec
	}
	return res / float64(len(l))
}

func (l CephStatusList) AvgWriteOpPerSec() float64 {
	res := float64(0)
	for _, v := range l {
		res = res + v.WriteOpPerSec
	}
	return res / float64(len(l))
}

type OSDStatus struct {
	ActiveClean bool `json:"activeClean"`
}

func (cluster *Cluster) OsdStatus(osdNum int64) (*OSDStatus, error) {
	var res OSDStatus
	res.ActiveClean = true

	type OSDStatusResp struct {
		Nodes []struct {
			Name   string `json:"name"`
			Status string `json:"status"`
		} `json:"nodes"`
	}
	resp, err := cluster.Master.ExecCmd("ceph osd tree -f json")
	if err != nil {
		return nil, err
	}

	var osdStatusResp OSDStatusResp
	err = json.Unmarshal(resp, &osdStatusResp)
	if err != nil {
		return nil, err
	}

	noExist := true
	for _, node := range osdStatusResp.Nodes {
		if node.Name != "osd."+strconv.Itoa(int(osdNum)) {
			continue
		}
		noExist = false
		if node.Status != "up" {
			res.ActiveClean = false
			return &res, nil
		}
	}

	if noExist {
		res.ActiveClean = false
		return &res, nil
	}

	type OSDPGResp struct {
		PGStats []struct {
			State string `json:"state"`
		} `json:"pg_stats"`
	}

	resp, err = cluster.Master.ExecCmd("ceph pg ls-by-osd " + strconv.Itoa(int(osdNum)) + " -f json")
	if err != nil {
		return nil, err
	}

	var osdPGResp OSDPGResp
	err = json.Unmarshal(resp, &osdPGResp)
	if err != nil {
		return nil, err
	}

	for _, pgStats := range osdPGResp.PGStats {
		if pgStats.State != "active+clean" {
			res.ActiveClean = false
			return &res, nil
		}
	}

	return &res, nil
}

func (cluster *Cluster) CurrentCephStatus() (*CephStatus, error) {
	type CephStatusResp struct {
		PGMap CephStatus `json:"pgmap"`
	}

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

	return &res, nil
}

func (cluster *Cluster) CephStatus(second int) (*CephStatusList, error) {
	type CephStatusResp struct {
		PGMap CephStatus `json:"pgmap"`
	}

	var list CephStatusList
	for i := 0; i < second; i++ {
		var res CephStatus
		resp, err := cluster.Master.ExecCmd("ceph status -f json")
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
		time.Sleep(time.Second)
	}

	return &list, nil
}
