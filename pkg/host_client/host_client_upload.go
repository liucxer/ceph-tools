package host_client

import (
	"github.com/pkg/sftp"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"time"
)

func (client *HostClient) Upload(dstPath string, srcPath string) error {
	startTime := time.Now()
	logrus.Debugf("Upload start. ipaddr:%s, srcPath:%s", client.IpAddr, srcPath)
	defer func() {
		cost := time.Now().Sub(startTime).Seconds()
		logrus.Debugf("Upload end.   ipaddr:%s, srcPath:%s, cost:%f", client.IpAddr, srcPath, cost)
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

	fileInfo, err := sftpClient.Stat(dstPath)
	if err != nil && err.Error() != "file does not exist" {
		logrus.Errorf("sftpClient.Stat err. [err:%v,client:%v]", err, client)
		return err
	}

	if err == nil {
		if fileInfo.IsDir() {
			srcPathFileInfo, err := os.Stat(srcPath)
			if err != nil {
				logrus.Errorf("os.Open err. [err:%v,dstPath:%s]", err, dstPath)
				return err
			}
			dstPath += "/" + srcPathFileInfo.Name()
		} else {
			err = sftpClient.Remove(dstPath)
			if err != nil {
				logrus.Errorf("sftp.Remove err. [err:%v,client:%v]", err, client)
				return err
			}
		}
	}

	dstFile, err := sftpClient.Create(dstPath)
	if err != nil {
		logrus.Errorf("sftpClient.Create err. [err:%v,client:%v,srcPath:%s,dstPath:%s]", err, client, srcPath, dstPath)
		return err
	}
	defer func() { _ = dstFile.Close() }()

	srcFile, err := os.Open(srcPath)
	if err != nil {
		logrus.Errorf("os.Open err. [err:%v,dstPath:%s]", err, dstPath)
		return err
	}
	defer func() { _ = srcFile.Close() }()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		logrus.Errorf("io.Copy err. [err:%v,dstPath:%s, srcFile:%s]", err, dstPath, srcPath)
		return err
	}

	return nil
}
