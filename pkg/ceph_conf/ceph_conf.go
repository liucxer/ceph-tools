package ceph_conf

import (
	"github.com/BurntSushi/toml"
	"github.com/sirupsen/logrus"
)

type GlobalConf struct {
	Fsid string `json:"fsid"`
}

type OSDConf struct {
}

type CephConf struct {
	GlobalConf OSDConf `toml:"global"`
	OSDConf    OSDConf `toml:"osd"`
}

func (conf *CephConf) UnmarshalJSON(bts []byte) error {
	metaData, err := toml.Decode(string(bts), conf)
	if err != nil {
		logrus.Errorf("toml.Decode, err:%v", err)
		return err
	}
	_ = metaData
	return nil
}

func (conf *CephConf) MarshalJSON() ([]byte, error) {
	return nil, nil
}
