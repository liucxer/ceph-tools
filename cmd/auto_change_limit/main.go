package main

import (
	"encoding/json"
	"fmt"
	"github.com/liucxer/ceph-tools/pkg/ceph"
	"github.com/liucxer/ceph-tools/pkg/cluster_client"
	"github.com/liucxer/confmiddleware/conflogger"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"math"
	"os"
	"strconv"
	"time"
)

type ExecConfig struct {
	DiskType string  `json:"diskType"`
	IpAddr   string  `json:"ipAddr"`
	OsdNum   []int64 `json:"osdNum"`
	MaxLimit float64 `json:"maxLimit"`
	MinLimit float64 `json:"minLimit"`
	Zoom     float64 `json:"zoom"`
	*ceph.CephConf
	cluster *cluster_client.Cluster
}

func (execConfig *ExecConfig) RefreshExecConfig(configFilePath string) error {
	go func() {
		for {
			bts, err := ioutil.ReadFile(configFilePath)
			if err != nil {
				logrus.Errorf("ioutil.ReadFile err:%v", err)
				return
			}

			err = json.Unmarshal(bts, execConfig)
			if err != nil {
				logrus.Errorf("json.Unmarshal err:%v", err)
				return
			}
			time.Sleep(time.Second)
		}
	}()
	time.Sleep(time.Second)
	return nil
}

func (execConfig *ExecConfig) Run() error {
	var (
		err error
	)

	var jobCostList ceph.JobCostList
	for _, osdNum := range execConfig.OsdNum {
		item, err := ceph.GetJobCostList(execConfig.cluster.Master, osdNum)
		if err != nil {
			return err
		}
		jobCostList = append(jobCostList, item...)
	}

	if len(jobCostList) == 0 {
		for _, osdNum := range execConfig.OsdNum {
			cmdStr := "ceph daemon osd." + strconv.Itoa(int(osdNum)) + " config set osd_op_queue_mclock_recov_lim 99999"
			_, err = execConfig.cluster.Master.ExecCmd(cmdStr)
		}
		return err
	}

	var coefficients []float64
	for _, jobCost := range jobCostList {
		if jobCost.Type == "write" {
			expectCost := execConfig.WriteLineMetaData.A*jobCost.ExpectCost + execConfig.WriteLineMetaData.B
			coefficient := jobCost.ActualCost / expectCost
			coefficients = append(coefficients, coefficient)
		} else {
			expectCost := execConfig.ReadLineMetaData.A*jobCost.ExpectCost + execConfig.ReadLineMetaData.B
			coefficient := jobCost.ActualCost / expectCost
			coefficients = append(coefficients, coefficient)
		}
	}

	sumCoefficient := float64(0)
	for _, coefficient := range coefficients {
		sumCoefficient += coefficient
	}
	avgCoefficient := sumCoefficient / float64(len(coefficients))

	minCoefficient := float64(1)
	maxCoefficient := execConfig.Zoom

	k := float64(1)
	if avgCoefficient < minCoefficient {
		k = 0
	} else if avgCoefficient > maxCoefficient {
		k = 1
	} else {
		k = math.Abs(avgCoefficient-minCoefficient) / (maxCoefficient - minCoefficient)
	}

	limit := execConfig.MaxLimit - (execConfig.MaxLimit-execConfig.MinLimit)*k

	logrus.Infof("execConfig:%+v, "+
		"avgCoefficient:%0.2f,"+
		"k:%0.2f,"+
		"limit:%0.2f",
		execConfig, avgCoefficient, k, limit)

	limitStr := fmt.Sprintf("%0.2f", limit)
	for _, osdNum := range execConfig.OsdNum {
		cmdStr := "ceph daemon osd." + strconv.Itoa(int(osdNum)) + " config set osd_op_queue_mclock_recov_lim " + limitStr
		_, err = execConfig.cluster.Master.ExecCmd(cmdStr)
	}
	return err
}

func init() {
	var logger = conflogger.Log{
		Name:  "fio",
		Level: "Debug",
	}
	logger.SetDefaults()
	logger.Init()
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage:\n     ./cmd config.json")
		return
	}

	execConfig := ExecConfig{}

	err := execConfig.RefreshExecConfig(os.Args[1])
	if err != nil {
		return
	}

	cluster, err := cluster_client.NewCluster([]string{execConfig.IpAddr})
	if err != nil {
		return
	}
	defer func() { _ = cluster.Close() }()

	execConfig.cluster = cluster

	execConfig.CephConf, err = ceph.NewCephConf(execConfig.cluster.Master, execConfig.OsdNum)
	if err != nil {
		return
	}

	for {
		_ = execConfig.Run()
		time.Sleep(1 * time.Second)
	}

}
