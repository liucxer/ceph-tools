package host_client

import (
	"github.com/pkg/sftp"
	"github.com/sirupsen/logrus"
	"io/ioutil"
)

func (client *HostClient) ReadCephConf() error {
	cephConfPath := "/etc/ceph/ceph.conf"
	sshClient, err := client.OpenSSHClient()
	if err != nil {
		return err
	}
	defer func() { _ = sshClient.Close() }()

	sftpClient, err := sftp.NewClient(sshClient, sftp.MaxPacket(1<<15))
	if err != nil {
		logrus.Errorf("sftp.NewClient err. [err:%v,client:%v]", err, client)
		return err
	}
	defer func() { _ = sftpClient.Close() }()

	srcFile, err := sftpClient.Open(cephConfPath)
	if err != nil {
		logrus.Errorf("sftpClient.Open err. [err:%v,client:%v,srcPath:%s]", err, client, cephConfPath)
		return err
	}
	defer func() { _ = srcFile.Close() }()

	bts, err := ioutil.ReadAll(srcFile)
	if err != nil {
		logrus.Errorf("ioutil.ReadAll err. [err:%v,client:%v,srcPath:%s]", err, client, cephConfPath)
		return err
	}

	_ = bts
	return nil
}
