package host_client

import (
	"ceph-tools/pkg/ceph_conf"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"time"
)

type HostClient struct {
	IpAddr     string
	Port       string
	User       string
	Password   string
	cephConfig *ceph_conf.CephConf
}

func NewHostClient(IpAddr string) *HostClient {
	return &HostClient{
		IpAddr:   IpAddr,
		Port:     "22",
		User:     "root",
		Password: "daemon",
	}
}

func NewHostClientWithPassword(ipAddr, port, user, password string) *HostClient {
	return &HostClient{
		IpAddr:   ipAddr,
		Port:     port,
		User:     user,
		Password: password,
	}
}

func (client *HostClient) SSHConfig() *ssh.ClientConfig {
	config := &ssh.ClientConfig{
		Timeout:         time.Second,
		User:            client.User,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	config.Auth = []ssh.AuthMethod{ssh.Password(client.Password)}
	return config
}

func (client *HostClient) OpenSSHClient() (*ssh.Client, error) {
	addr := client.IpAddr + ":" + client.Port
	sshClient, err := ssh.Dial("tcp", addr, client.SSHConfig())
	if err != nil {
		logrus.Errorf("ssh.Dial err. [err:%v,client:%v]", err, client)
		return nil, err
	}
	return sshClient, nil
}
