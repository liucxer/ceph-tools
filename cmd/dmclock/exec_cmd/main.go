package main

import (
	"ceph-tools/pkg/cluster_client"
	"github.com/sirupsen/logrus"
	"os"
	"time"
)

func main() {
	// fio /home/lxq/qos/perf_script/fio/fio_4k_randwrite_1client.rbd.cfg
	var (
		err error
		c   *cluster_client.Cluster
	)

	localLogDir := "/Users/liucx/Desktop/ceph/"
	err = os.RemoveAll(localLogDir)
	if err != nil {
		logrus.Errorf("os.RemoveAll err. [err:%v, localLogDir:%s]", err, localLogDir)
		return
	}

	err = os.Mkdir(localLogDir, os.ModePerm)
	if err != nil {
		logrus.Errorf("os.Mkdir err. [err:%v, localLogDir:%s]", err, localLogDir)
		return
	}

	logrus.SetLevel(logrus.DebugLevel)
	c, err = cluster_client.NewCluster([]string{"10.0.20.27"})
	if err != nil {
		return
	}
	defer func() { _ = c.Close() }()

	timeNow := time.Now()

	err = c.ClearCephLog()
	if err != nil {
		return
	}

	_, err = c.Clients[0].Host.ExecCmd("fio /home/liucx/fio_4M_write_1client_16.rbd.cfg")
	if err != nil {
		return
	}

	err = c.CollectCephLog(localLogDir)
	if err != nil {
		return
	}

	err = c.ExecCmd("ls -lh /var/log/ceph")
	if err != nil {
		return
	}

	logrus.Infof("cost:%f s", time.Now().Sub(timeNow).Seconds())
}
