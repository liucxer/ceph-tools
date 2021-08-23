package ceph

import (
	"github.com/liucxer/ceph-tools/pkg/host_client"
	"github.com/liucxer/ceph-tools/pkg/interfacer"
	"github.com/liucxer/ceph-tools/pkg/line"
	"github.com/sirupsen/logrus"
)

type CephConf struct {
	interfacer.Worker
	Ips                     []string                          `json:"ips"`
	HostClient              []host_client.HostClient          `json:"hostClient"`
	OsdNum                  []int64                           `json:"osdNum"`
	OsdNumMap               map[int64]*host_client.HostClient `json:"osdNumMap"`
	OsdNumReadLineMetaData  map[int64]line.LineMetaData       `json:"osdNumReadLineMetaData"`
	OsdNumWriteLineMetaData map[int64]line.LineMetaData       `json:"osdNumWriteLineMetaData"`
	ReadLineMetaData        line.LineMetaData                 `json:"readLineMetaData"`
	WriteLineMetaData       line.LineMetaData                 `json:"writeLineMetaData"`
}

func NewCephConf(worker interfacer.Worker, osdNums []int64) (*CephConf, error) {
	var (
		conf CephConf
		err  error
	)

	conf.Worker = worker
	conf.OsdNum = osdNums
	conf.OsdNumMap = map[int64]*host_client.HostClient{}

	ipMap, err := GetOSDIpMap(worker, osdNums)
	if err != nil {
		return &conf, err
	}

	for _, osdNum := range osdNums {
		hostClient, err := host_client.NewHostClient(ipMap[osdNum])
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
