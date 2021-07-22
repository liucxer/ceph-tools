package main

import (
	"ceph-tools/host"
	"github.com/sirupsen/logrus"
)

func main() {
	var (
		err error
	)

	logrus.SetLevel(logrus.DebugLevel)
	// build
	hostClient := host.NewHost("10.0.7.242")
	_, err = hostClient.ExecCmd("cd /workspace/liuchangxi/inficore-v4/env/infinity-4.2.8/build && make -j16 ceph-osd")
	if err != nil {
		return
	}

	//// 下载 ceph-osd
	//dstPath := os.TempDir() + "ceph.osd"
	//scp.Download(dstPath, hostClient, "/workspace/liuchangxi/inficore-v4/env/infinity-4.2.8/build/bin/ceph-osd")
	//
	//// deploy
	//clusterClient, err := cluster.NewCluster([]string{"10.0.20.28"})
	//if err != nil {
	//	return
	//}
	//
	//defer clusterClient.Close()
	//
	//err = clusterClient.ExecCmd("systemctl stop ceph.target")
	//if err != nil {
	//	return
	//}
	//
	//for _, hostClient := range clusterClient.Clients {
	//	err = scp.Upload(hostClient.Host, "/usr/bin/ceph-osd", dstPath)
	//	if err != nil {
	//		return
	//	}
	//}
	//
	//err = clusterClient.ExecCmd("systemctl start ceph.target")
	//if err != nil {
	//	return
	//}
}
