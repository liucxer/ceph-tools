package cluster_client

import (
	"encoding/json"
	"strconv"
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

type OSDPerf struct {
	OSD struct {
		Numpg         int64 `json:"numpg"`
		NumpgPrimary  int64 `json:"numpg_primary"`
		NumpgReplica  int64 `json:"numpg_replica"`
		NumpgStray    int64 `json:"numpg_stray"`
		NumpgRemoving int64 `json:"numpg_removing"`
	}
}

func (cluster *Cluster) CurrentOSDPerf(osd string) (*OSDPerf, error) {
	var res OSDPerf
	resp, err := cluster.Clients[0].ExecCmd("ceph daemon " + osd + " perf dump")
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil

}

type OSDStatus struct {
	ActiveClean bool `json:"activeClean"`
}

func (cluster *Cluster) OsdStatus(osdNum int64) (*OSDStatus, error) {
	type OSDPGResp struct {
		PGStats []struct {
			State string `json:"state"`
		} `json:"pg_stats"`
	}

	resp, err := cluster.Clients[0].ExecCmd("ceph pg ls-by-osd " + strconv.Itoa(int(osdNum)) + " -f json")
	if err != nil {
		return nil, err
	}

	var osdPGResp OSDPGResp
	err = json.Unmarshal(resp, &osdPGResp)
	if err != nil {
		return nil, err
	}

	var res OSDStatus
	res.ActiveClean = true
	for _, pgStats := range osdPGResp.PGStats {
		if pgStats.State != "active+clean" {
			res.ActiveClean = false
			return &res, nil
		}
	}

	type OSDStatusResp struct {
		Nodes []struct {
			Name   string `json:"name"`
			Status string `json:"status"`
		} `json:"nodes"`
	}
	resp, err = cluster.Clients[0].ExecCmd("ceph osd tree -f json")
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
