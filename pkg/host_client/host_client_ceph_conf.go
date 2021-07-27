package host_client

import (
	"github.com/sirupsen/logrus"
	"io/ioutil"
)

func (client *HostClient) ReadCephConf() error {
	cephConfPath := "/etc/ceph/ceph.conf"

	srcFile, err := client.sftpClient.Open(cephConfPath)
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
