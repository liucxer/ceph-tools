package main

import (
	"fmt"
	"github.com/liucxer/ceph-tools/pkg/cluster_client"
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

	err = c.ClearCephLog()
	if err != nil {
		return
	}
}
