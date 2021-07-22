package main

import (
	"ceph-tools/cluster"
	"github.com/sirupsen/logrus"
	"os"
	"time"
)

// 清空日志
// 执行rbd压力测试 120秒
// 收集日志
// 分析日志

func main() {
	var (
		err error
		c   *cluster.Cluster
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
	c, err = cluster.NewCluster([]string{"10.0.20.28"})
	if err != nil {
		return
	}
	defer func() { _ = c.Close() }()

	timeNow := time.Now()

	err = c.ClearCephLog()
	if err != nil {
		return
	}

	//go func() {
	_, err = c.Clients[0].Host.ExecCmd("rbd -p pool_data bench rbd_image --io-type write --io-size 8M --io-threads 16 --io-total 50G --io-pattern seq")
	if err != nil {
		return
	}
	//}()

	time.Sleep(10 * time.Second)
	_, err = c.Clients[0].Host.ExecCmd("killall rbd")
	if err != nil {
		return
	}

	//time.Sleep(30 * time.Second)
	_, err = c.Clients[0].Host.ExecCmd("ps -aux |grep rbd")
	err = c.CollectCephLog(localLogDir)
	if err != nil {
		return
	}

	err = c.ExecCmd("ls -lh /var/log/ceph")
	if err != nil {
		return
	}

	//logrus.Infof("cost:%f s", time.Now().Sub(timeNow).Seconds())
}
