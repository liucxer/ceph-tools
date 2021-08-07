package host_client

import "encoding/json"

type OSDPerf struct {
	OSD struct {
		Numpg         int64 `json:"numpg"`
		NumpgPrimary  int64 `json:"numpg_primary"`
		NumpgReplica  int64 `json:"numpg_replica"`
		NumpgStray    int64 `json:"numpg_stray"`
		NumpgRemoving int64 `json:"numpg_removing"`
	}
}

func (client *HostClient) OSDPerf(osd string) (*OSDPerf, error) {
	var res OSDPerf
	resp, err := client.ExecCmd("ceph daemon " + osd + " perf dump")
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil

}
