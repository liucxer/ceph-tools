package host_client

import (
	"github.com/sirupsen/logrus"
	"io"
	"os"
)

func (client *HostClient) Download(dstPath string, srcPath string) error {
	var (
		err error
	)

	srcFile, err := client.sftpClient.Open(srcPath)
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
