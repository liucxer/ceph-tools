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
	"strings"
	"time"
)

type ExecConfig struct {
	MapOsd4KRandWriteIops map[int64]float64 `json:"mapOsd4KRandWriteIops"`
	AvgOsd4KRandWriteIops float64           `json:"avgOsd4KRandWriteIops"`
	DiskType              string            `json:"diskType"`
	IpAddr                string            `json:"ipAddr"`
	OsdNum                []int64           `json:"osdNum"`
	MaxLimit              float64           `json:"maxLimit"`
	MinLimit              float64           `json:"minLimit"`
	Zoom                  float64           `json:"zoom"`
	WriteStdCoefficient   string            `json:"writeStdCoefficient"`
	ReadStdCoefficient    string            `json:"readStdCoefficient"`
	cluster               *cluster_client.Cluster
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
			aStr := strings.Split(execConfig.WriteStdCoefficient, "x")[0]
			aFloat, _ := strconv.ParseFloat(aStr, 32)

			bStr := strings.Split(execConfig.WriteStdCoefficient, "x")[1]
			bFloat, _ := strconv.ParseFloat(bStr, 32)
			expectCost := aFloat*jobCost.ExpectCost + bFloat
			coefficient := jobCost.ActualCost / expectCost
			coefficients = append(coefficients, coefficient)
		} else {
			aStr := strings.Split(execConfig.WriteStdCoefficient, "x")[0]
			aFloat, _ := strconv.ParseFloat(aStr, 32)

			bStr := strings.Split(execConfig.WriteStdCoefficient, "x")[1]
			bFloat, _ := strconv.ParseFloat(bStr, 32)
			expectCost := aFloat*jobCost.ExpectCost + bFloat
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

func (execConfig *ExecConfig) Init4KRandWriteIops() error {
	execConfig.MapOsd4KRandWriteIops = map[int64]float64{}
	sumIops := float64(0)
	for _, osdNum := range execConfig.OsdNum {
		iops, err := ceph.Get4KRandWriteIops(execConfig.cluster.Master, osdNum)
		if err != nil {
			return err
		}
		execConfig.MapOsd4KRandWriteIops[osdNum] = iops
		sumIops += iops
	}

	execConfig.AvgOsd4KRandWriteIops = sumIops / float64(len(execConfig.OsdNum))
	return nil
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

	for {
		_ = execConfig.Run()
		time.Sleep(1 * time.Second)
	}

}
