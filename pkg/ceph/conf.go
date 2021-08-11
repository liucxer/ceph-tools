package ceph

import (
	"github.com/liucxer/ceph-tools/pkg/host_client"
	"github.com/liucxer/ceph-tools/pkg/interfacer"
	"github.com/sirupsen/logrus"
)

type CephConf struct {
	Ips []string `json:"ips"`
	HostClient []host_client.HostClient `json:"hostClient"`
	OsdNumMap map[int64]*host_client.HostClient `json:"osdNumMap"`

}

func NewCephConf(worker interfacer.Worker,osdNums []int64) (*CephConf, error) {
	var (
		conf CephConf
		err error
	)

	conf.OsdNumMap = map[int64]*host_client.HostClient{}
	for _, osdNum := range osdNums {
		ip, err := GetOSDIp(worker, osdNum)
		if err != nil {
			return &conf, err
		}
		hostClient, err := host_client.NewHostClient(ip)
		if err != nil {
			return &conf, err
		}
		conf.OsdNumMap[osdNum] = hostClient
	}

	conf.Ips, err = GetAllIps(worker)
	if err != nil {
		return &conf, err
	}

	for _, ip := range conf.Ips {
		hostClient, err := host_client.NewHostClient(ip)
		if err != nil {
			return &conf, err
		}
		conf.HostClient = append(conf.HostClient, *hostClient)
	}

	logrus.Debugf("NewCephConf. CephConf:%+v", conf)
	return &conf, err
}