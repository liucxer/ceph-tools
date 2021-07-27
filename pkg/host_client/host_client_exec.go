package host_client

import (
	"github.com/sirupsen/logrus"
	"time"
)

func (client *HostClient) ExecCmd(cmd string) ([]byte, error) {
	var (
		err error
	)
	startTime := time.Now()
	result := ""
	logrus.Debugf("ExecCmd start. ipaddr:%s, cmd:\"%s\"", client.IpAddr, cmd)
	defer func() {
		cost := time.Now().Sub(startTime).Seconds()
		logrus.Debugf("ExecCmd end.   ipaddr:%s, cmd:\"%s\", cost:%fs, result:\n%s", client.IpAddr, cmd, cost, result)
	}()

	sshSession, err := client.sshClient.NewSession()
	if err != nil {
		return nil, err
	}
	defer func() { _ = sshSession.Close() }()

	combo, err := sshSession.CombinedOutput(cmd)
	result = string(combo)
	if err != nil {
		logrus.Errorf("session.CombinedOutput error [host:%s, err:%v, cmd:%s]", client.IpAddr, err, cmd)
		return nil, err
	}
	return combo, err
}
