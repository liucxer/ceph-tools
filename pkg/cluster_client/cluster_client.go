package cluster_client

import (
	"errors"
	"github.com/liucxer/ceph-tools/pkg/host_client"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"time"
)

type Cluster struct {
	Master  *host_client.HostClient
	Clients []*host_client.HostClient
}

func NewCluster(ipAddresses []string) (*Cluster, error) {
	if len(ipAddresses) == 0 {
		return nil, errors.New("ipAddresses is empty")
	}
	var cluster Cluster
	defaultUser := "root"
	defaultPassword := "daemon"
	//defaultPort := "22"
	for _, ipAddr := range ipAddresses {
		config := &ssh.ClientConfig{
			Timeout:         3 * time.Second,
			User:            defaultUser,
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			//HostKeyCallback: hostKeyCallBackFunc(h.Host),
		}
		config.Auth = []ssh.AuthMethod{ssh.Password(defaultPassword)}

		client, err := host_client.NewHostClient(ipAddr)
		if err != nil {
			return nil, err
		}

		cluster.Clients = append(cluster.Clients, client)
	}
	cluster.Master = cluster.Clients[0]
	return &cluster, nil
}

func (cluster *Cluster) Close() error {
	var (
		err error
	)
	for _, client := range cluster.Clients {
		err = client.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func (cluster *Cluster) ExecCmd(cmd string) error {
	for _, client := range cluster.Clients {
		_, err := client.ExecCmd(cmd)
		if err != nil {
			logrus.Errorf("host :%s, ExecCmd: err:%v", client.IpAddr, err)
			return err
		}
	}
	return nil
}

func (cluster *Cluster) ClearOsdLog(osdNums []int64) error {
	startTime := time.Now()
	logrus.Debugf("ClearOsdLog start. ")
	defer func() {
		cost := time.Now().Sub(startTime).Seconds()
		logrus.Debugf("ClearOsdLog end. cost:%f", cost)
	}()
	for _, client := range cluster.Clients {
		err := client.ClearOsdLog(osdNums)
		if err != nil {
			logrus.Errorf("host :%s, ExecCmd: err:%v", client.IpAddr, err)
			return err
		}
	}

	return nil
}

func (cluster *Cluster) ClearCephLog() error {
	startTime := time.Now()
	logrus.Debugf("ClearCephLog start. ")
	defer func() {
		cost := time.Now().Sub(startTime).Seconds()
		logrus.Debugf("ClearCephLog end. cost:%f", cost)
	}()
	for _, client := range cluster.Clients {
		err := client.ClearCephLog()
		if err != nil {
			logrus.Errorf("host :%s, ExecCmd: err:%v", client.IpAddr, err)
			return err
		}
	}

	return nil
}

func (cluster *Cluster) CollectOsdLog(dstDir string,osdNums []int64) error {
	var (
		err error
	)

	startTime := time.Now()
	logrus.Debugf("CollectOsdLog start. ")
	defer func() {
		cost := time.Now().Sub(startTime).Seconds()
		logrus.Debugf("CollectOsdLog end. cost:%f", cost)
	}()

	for _, client := range cluster.Clients {
		err = client.CollectOsdLog(dstDir, osdNums)
		if err != nil {
			return err
		}
	}
	return nil
}

func (cluster *Cluster) CollectCephLog(dstDir string) error {
	var (
		err error
	)

	startTime := time.Now()
	logrus.Debugf("CollectCephLog start. ")
	defer func() {
		cost := time.Now().Sub(startTime).Seconds()
		logrus.Debugf("CollectCephLog end. cost:%f", cost)
	}()

	for _, client := range cluster.Clients {
		err = client.CollectCephLog(dstDir)
		if err != nil {
			return err
		}
	}
	return nil
}

func (cluster *Cluster) BackupAndClearCephLog() error {
	startTime := time.Now()
	logrus.Debugf("BackupAndClearCephLog start. ")
	defer func() {
		cost := time.Now().Sub(startTime).Seconds()
		logrus.Debugf("BackupAndClearCephLog end. cost:%f", cost)
	}()

	for _, client := range cluster.Clients {
		err := client.BackupAndClearCephLog()
		if err != nil {
			return err
		}
	}
	return nil
}
