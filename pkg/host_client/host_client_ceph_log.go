package host_client

import (
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"strings"
)

func (client *HostClient) BackupAndClearCephLog() error {
	var (
		err error
	)

	osdLogFiles, err := client.OsdLogFiles()
	if err != nil {
		return err
	}

	for _, logfile := range osdLogFiles {
		dstLogFile := logfile + ".backup"
		backupCmdString := " cp -f " + logfile + " " + dstLogFile + " &&" + " echo '' >  /var/log/ceph/" + logfile
		_, err = client.ExecCmd(backupCmdString)
		if err != nil {
			return err
		}
	}

	return nil
}

type BackupCephLog struct {
	FileName string
	LogData  *[]byte
}

func (client *HostClient) ReadBackupCephLog() ([]BackupCephLog, error) {
	var (
		err error
		res []BackupCephLog
	)

	osdLogFiles, err := client.OsdBackupLogFiles()
	if err != nil {
		return res, err
	}

	for _, logfile := range osdLogFiles {
		srcFile, err := client.sftpClient.Open(logfile)
		if err != nil {
			logrus.Errorf("sftpClient.Open err. [err:%v,client:%v,srcLogFile:%s]", err, client, logfile)
			return res, err
		}
		defer func() { _ = srcFile.Close() }()
		logData, err := ioutil.ReadAll(srcFile)
		if err != nil {
			logrus.Errorf("ioutil.ReadAll err. [err:%v,client:%v,srcLogFile:%s]", err, client, logfile)
			return res, err
		}

		item := BackupCephLog{
			FileName: logfile,
			LogData:  &logData,
		}
		res = append(res, item)
	}

	return res, err
}

func (client *HostClient) OsdBackupLogFiles() ([]string, error) {
	var (
		err            error
		osdLogFiles    []string
		tmpLogFiles    []string
		lsCmdStringRes []byte
	)
	lsCmdString := "ls /var/log/ceph"
	lsCmdStringRes, err = client.ExecCmd(lsCmdString)
	if err != nil {
		return osdLogFiles, err
	}

	tmpLogFiles = strings.Split(string(lsCmdStringRes), "\n")

	for _, logfile := range tmpLogFiles {
		if logfile == "" {
			continue
		}
		if !strings.Contains(logfile, "ceph-osd") {
			continue
		}
		if !strings.Contains(logfile, ".backup") {
			continue
		}
		osdLogFiles = append(osdLogFiles, "/var/log/ceph/"+logfile)
	}
	return osdLogFiles, err
}

func (client *HostClient) OsdLogFiles() ([]string, error) {
	var (
		err            error
		osdLogFiles    []string
		tmpLogFiles    []string
		lsCmdStringRes []byte
	)
	lsCmdString := "ls /var/log/ceph"
	lsCmdStringRes, err = client.ExecCmd(lsCmdString)
	if err != nil {
		return osdLogFiles, err
	}

	tmpLogFiles = strings.Split(string(lsCmdStringRes), "\n")

	for _, logfile := range tmpLogFiles {
		if logfile == "" {
			continue
		}
		if !strings.Contains(logfile, "ceph-osd") {
			continue
		}
		if strings.Contains(logfile, ".backup") {
			continue
		}
		osdLogFiles = append(osdLogFiles, "/var/log/ceph/"+logfile)
	}
	return osdLogFiles, err
}

func (client *HostClient) ClearCephLog() error {
	var (
		err error
	)

	osdLogFiles, err := client.OsdLogFiles()
	if err != nil {
		return err
	}

	for _, logfile := range osdLogFiles {
		clearCmdString := "echo '' >  " + logfile
		_, err = client.ExecCmd(clearCmdString)
		if err != nil {
			return err
		}
	}

	return nil
}

func (client *HostClient) CollectCephLog(dstDir string) error {
	var (
		err error
	)

	osdLogFiles, err := client.OsdLogFiles()
	if err != nil {
		return err
	}

	for _, logfile := range osdLogFiles {
		fileNameList := strings.Split(logfile, "/")
		fileName := fileNameList[len(fileNameList)-1]
		dstPath := dstDir + "/" + fileName
		err = client.Download(dstPath, logfile)
		if err != nil {
			return err
		}
	}
	return nil
}
