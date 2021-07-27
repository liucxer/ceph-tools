package ceph_client

type AuthReq struct {
	UserName string `json:"username"`
	Password string `json:"password"`
}

type AuthResp struct {
	Token string `json:"token"`
}

type HealthFullReq struct {
}

type HealthFullResp struct {
}

type HealthMinimalReq struct {
}

/*
	"client_perf": {
		"read_bytes_sec": 5033,
		"read_op_per_sec": 2,
		"recovering_bytes_per_sec": 0,
		"write_bytes_sec": 938103,
		"write_op_per_sec": 229
	},
*/

type HealthMinimalResp struct {
	ClientPerf struct {
		ReadBytesSec          int64 `json:"read_bytes_sec"`
		ReadOpPerSec          int64 `json:"read_op_per_sec"`
		RecoveringBytesPerSec int64 `json:"recovering_bytes_per_sec"`
		WriteBytesSec         int64 `json:"write_bytes_sec"`
		WriteOpPerSec         int64 `json:"write_op_per_sec"`
	} `json:"client_perf"`
}
