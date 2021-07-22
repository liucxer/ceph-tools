package cluster_client

import (
	"ceph-tools/pkg/host_client"
	"fmt"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"strings"
	"time"
)

type Client struct {
	Host       *host_client.HostClient
	sshClient  *ssh.Client
	sshSession *ssh.Session
}

type Cluster struct {
	Clients []Client
}

func NewCluster(ipAddrs []string) (*Cluster, error) {
	var cluster Cluster
	defaultUser := "root"
	defaultPassword := "daemon"
	defaultPort := "22"
	for _, ipAddr := range ipAddrs {
		//创建sshp登陆配置
		config := &ssh.ClientConfig{
			Timeout:         time.Second, //ssh 连接time out 时间一秒钟, 如果ssh验证错误 会在一秒内返回
			User:            defaultUser,
			HostKeyCallback: ssh.InsecureIgnoreHostKey(), //这个可以， 但是不够安全
			//HostKeyCallback: hostKeyCallBackFunc(h.Host),
		}
		config.Auth = []ssh.AuthMethod{ssh.Password(defaultPassword)}

		//dial 获取ssh client
		addr := fmt.Sprintf("%s:%s", ipAddr, defaultPort)
		sshClient, err := ssh.Dial("tcp", addr, config)
		if err != nil {
			logrus.Errorf("ssh.Dial error [ipAddr:%s, err:%v]", ipAddr, err)
			return nil, err
		}

		//创建ssh-session
		session, err := sshClient.NewSession()
		if err != nil {
			logrus.Errorf("sshClient.NewSession error [ipAddr:%s, err:%v]", ipAddr, err)
			return nil, err
		}
		cluster.Clients = append(cluster.Clients, Client{
			Host:       host_client.NewHostClient(ipAddr),
			sshClient:  sshClient,
			sshSession: session,
		})
	}
	return &cluster, nil
}

func (cluster *Cluster) Close() error {
	var (
		err error
	)
	for _, client := range cluster.Clients {
		err = client.sshClient.Close()
		if err != nil {
			return err
		}
		err = client.sshSession.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func (cluster *Cluster) ExecCmd(cmd string) error {
	for _, client := range cluster.Clients {
		_, err := host_client.NewHostClient(client.Host.IpAddr).ExecCmd(cmd)
		if err != nil {
			logrus.Errorf("host :%s, ExecCmd: err:%v", client.Host.IpAddr, err)
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
		cmdStr := "ls /var/log/ceph"
		session, err := client.sshClient.NewSession()
		if err != nil {
			logrus.Errorf("sshClient.NewSession error [host:%s, err:%v]", client.Host, err)
			return err
		}
		res, err := session.CombinedOutput(cmdStr)
		if err != nil {
			logrus.Errorf("sshSession.CombinedOutput error. [host:%s, err:%v, cmdStr:%s]", client.Host, err, cmdStr)
			return err
		}
		session.Close()

		logfiles := strings.Split(string(res), "\n")

		for _, logfile := range logfiles {
			if logfile == "" {
				continue
			}
			cmdStr := "echo '' >  /var/log/ceph/" + logfile
			client.sshSession.Stdout = nil
			client.sshSession.Stderr = nil
			session, err := client.sshClient.NewSession()
			if err != nil {
				logrus.Errorf("sshClient.NewSession error [host:%s, err:%v]", client.Host, err)
				return err
			}
			res, err = session.CombinedOutput(cmdStr)
			if err != nil {
				logrus.Errorf("host :%s, sshSession.CombinedOutput: err:%v, cmdStr:%s", client.Host, err, cmdStr)
				return err
			}
			session.Close()
		}
	}
	return nil
}

func (cluster *Cluster) CollectCephLog(dstDir string) error {
	for _, client := range cluster.Clients {
		cmdStr := "ls /var/log/ceph"
		session, err := client.sshClient.NewSession()
		if err != nil {
			logrus.Errorf("sshClient.NewSession error [host:%s, err:%v]", client.Host, err)
			return err
		}
		res, err := session.CombinedOutput(cmdStr)
		if err != nil {
			logrus.Errorf("sshSession.CombinedOutput error. [host:%s, err:%v, cmdStr:%s]", client.Host, err, cmdStr)
			return err
		}
		defer func() { _ = session.Close() }()

		logfiles := strings.Split(string(res), "\n")

		for _, logfile := range logfiles {
			if logfile == "" {
				continue
			}
			if !strings.Contains(logfile, "ceph-osd") {
				continue
			}
			srcPath := "/var/log/ceph/" + logfile
			dstPath := dstDir + "/" + logfile
			err = client.Host.Download(dstPath, srcPath)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
