package host_client

import (
	"github.com/pkg/sftp"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"time"
)

type HostClient struct {
	IpAddr     string
	Port       string
	User       string
	Password   string
	sshClient  *ssh.Client
	sftpClient *sftp.Client
}

func NewHostClient(IpAddr string) (*HostClient, error) {
	var (
		err    error
		client HostClient
	)
	client = HostClient{
		IpAddr:   IpAddr,
		Port:     "22",
		User:     "root",
		Password: "daemon",
	}

	err = client.open()
	return &client, err
}

func NewHostClientWithPassword(ipAddr, port, user, password string) *HostClient {
	return &HostClient{
		IpAddr:   ipAddr,
		Port:     port,
		User:     user,
		Password: password,
	}
}

func (client *HostClient) sshConfig() *ssh.ClientConfig {
	config := &ssh.ClientConfig{
		Timeout:         time.Second,
		User:            client.User,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	config.Auth = []ssh.AuthMethod{ssh.Password(client.Password)}
	return config
}

func (client *HostClient) open() error {
	var (
		err error
	)
	addr := client.IpAddr + ":" + client.Port
	client.sshClient, err = ssh.Dial("tcp", addr, client.sshConfig())
	if err != nil {
		logrus.Errorf("ssh.Dial err. [err:%v,client:%v]", err, client)
		return err
	}

	client.sftpClient, err = sftp.NewClient(client.sshClient, sftp.MaxPacket(1<<15))
	if err != nil {
		logrus.Errorf("sftp.NewClient err. [err:%v,client:%v]", err, client)
		return err
	}

	return nil
}

func (client *HostClient) Close() error {
	var (
		err error
	)

	err = client.sftpClient.Close()
	if err != nil {
		logrus.Errorf("client.sshSession.Close error [client:%v, err:%v]", client, err)
		return err
	}

	err = client.sshClient.Close()
	if err != nil {
		logrus.Errorf("client.sshClient.Close error [client:%v, err:%v]", client, err)
		return err
	}

	return nil
}
