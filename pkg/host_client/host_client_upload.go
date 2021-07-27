package host_client

import (
	"github.com/sirupsen/logrus"
	"io"
	"os"
)

func (client *HostClient) Upload(dstPath string, srcPath string) error {
	fileInfo, err := client.sftpClient.Stat(dstPath)
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
			err = client.sftpClient.Remove(dstPath)
			if err != nil {
				logrus.Errorf("sftp.Remove err. [err:%v,client:%v]", err, client)
				return err
			}
		}
	}

	dstFile, err := client.sftpClient.Create(dstPath)
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
