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
	"sync"
	"time"
)

type ExecConfig struct {
	DiskType string  `json:"diskType"`
	IpAddr   string  `json:"ipAddr"`
	OsdNum   []int64 `json:"osdNum"`
	Iops     float64 `json:"iops"`
	MaxLimit float64 `json:"maxLimit"`
	MinLimit float64 `json:"minLimit"`
	Zoom     float64 `json:"zoom"`
	cluster  *cluster_client.Cluster
}

func (execConfig *ExecConfig) ReadConfig(configFilePath string) error {
	bts, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		logrus.Errorf("ioutil.ReadFile err:%v", err)
		return err
	}

	err = json.Unmarshal(bts, execConfig)
	if err != nil {
		logrus.Errorf("json.Unmarshal err:%v", err)
		return err
	}
	return err
}

func (execConfig *ExecConfig) RunOneOsd(osdNum int64) error {
	jobCostList, err := ceph.GetJobCostList(execConfig.cluster.Master, osdNum)
	if err != nil {
		return err
	}
	if len(jobCostList) == 0 {
		cmdStr := "ceph daemon osd." + strconv.Itoa(int(osdNum)) + " config set osd_op_queue_mclock_recov_lim 99999"
		_, err = execConfig.cluster.Master.ExecCmd(cmdStr)
		return err
	}

	avgExpectCost := jobCostList.ExpectCost()
	avgActualCost := jobCostList.ActualCost()

	coefficient := avgActualCost / (avgExpectCost * 1000 / float64(execConfig.Iops))
	stdCoefficient := 6.7829 * math.Pow(avgExpectCost, -0.509)

	minCoefficient := stdCoefficient
	maxCoefficient := stdCoefficient * execConfig.Zoom

	k := float64(1)
	if coefficient < minCoefficient {
		k = 0
	} else if coefficient > maxCoefficient {
		k = 1
	} else {
		k = math.Abs(coefficient-minCoefficient) / (maxCoefficient - minCoefficient)
	}

	limit := execConfig.MaxLimit - (execConfig.MaxLimit-execConfig.MinLimit)*k

	logrus.Infof("execConfig:%+v, "+
		"avgExpectCost:%0.2f,"+
		"avgActualCost:%0.2f,"+
		"coefficient:%0.2f,"+
		"stdCoefficient:%0.2f,"+
		"k:%0.2f,"+
		"limit:%0.2f",
		execConfig, avgExpectCost, avgActualCost, coefficient, stdCoefficient, k, limit)

	cmdStr := "ceph daemon osd." + strconv.Itoa(int(osdNum)) + " config set osd_op_queue_mclock_recov_lim " + strconv.FormatFloat(limit, 'E', -1, 32)
	_, err = execConfig.cluster.Master.ExecCmd(cmdStr)
	return err
}

func (execConfig *ExecConfig) Run() error {
	var wg sync.WaitGroup

	for _, osdNum := range execConfig.OsdNum {
		wg.Add(1)
		go func(osdNum int64) {
			for {
				_ = execConfig.RunOneOsd(osdNum)
				time.Sleep(1 * time.Second)
			}
			wg.Done()
		}(osdNum)
	}
	wg.Wait()
	return nil
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
	err := execConfig.ReadConfig(os.Args[1])
	if err != nil {
		return
	}

	cluster, err := cluster_client.NewCluster([]string{execConfig.IpAddr})
	if err != nil {
		return
	}
	defer func() { _ = cluster.Close() }()

	execConfig.cluster = cluster
	_ = execConfig.Run()
}
