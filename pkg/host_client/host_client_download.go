package host_client

import (
	"github.com/pkg/sftp"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"time"
)

func (client *HostClient) Download(dstPath string, srcPath string) error {
	startTime := time.Now()
	logrus.Debugf("Download start. client:%v, srcPath:%s", client, srcPath)
	defer func() {
		cost := time.Now().Sub(startTime).Seconds()
		logrus.Debugf("Download end.   client:%v, srcPath:%s, cost:%f", client, srcPath, cost)
	}()

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

	srcFile, err := sftpClient.Open(srcPath)
	if err != nil {
		logrus.Errorf("sftpClient.Open err. [err:%v,client:%v,srcPath:%s]", err, client, srcPath)
		return err
	}
	defer func() { _ = srcFile.Close() }()

	dstFile, err := os.Create(dstPath)
	if err != nil {
		logrus.Errorf("os.Open err. [err:%v,dstPath:%s]", err, dstPath)
		return err
	}
	defer func() { _ = dstFile.Close() }()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		logrus.Errorf("io.Copy err. [err:%v,dstPath:%s, srcFile:%s]", err, dstPath, srcPath)
		return err
	}

	return nil
}
