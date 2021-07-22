package main

import (
	"ceph-tools/pkg/cluster_client"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage:\n     ./cmd ipaddrs")
		return
	}
	logrus.SetLevel(logrus.DebugLevel)
	c, err := cluster_client.NewCluster(os.Args[1:])
	if err != nil {
		return
	}
	defer func() { _ = c.Close() }()

	for _, client := range c.Clients {
		_, err = client.Host.ExecCmd("ls -lh /var/log/ceph/")
		if err != nil {
			return
		}
	}

	err = c.ClearCephLog()
	if err != nil {
		return
	}

	for _, client := range c.Clients {
		_, err = client.Host.ExecCmd("ls -lh /var/log/ceph/")
		if err != nil {
			return
		}
	}
}
