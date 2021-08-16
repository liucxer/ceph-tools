package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/liucxer/ceph-tools/pkg/ceph"
	"github.com/liucxer/ceph-tools/pkg/cluster_client"
	"github.com/liucxer/confmiddleware/conflogger"
	"github.com/sirupsen/logrus"
)

type ExecConfig struct {
	DiskType            string  `json:"diskType"`
	IpAddr              string  `json:"ipAddr"`
	OsdNum              []int64 `json:"osdNum"`
	MaxLimit            float64 `json:"maxLimit"`
	MinLimit            float64 `json:"minLimit"`
	Zoom                float64 `json:"zoom"`
	LastLimit           float64 `json:"lastLimit"`
	WriteStdCoefficient string  `json:"writeStdCoefficient"`
	ReadStdCoefficient  string  `json:"readStdCoefficient"`
	*ceph.CephConf
	*cluster_client.Cluster
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

func (execConfig *ExecConfig) Run(osdNum int64) error {
	var (
		err error
	)

	jobCostList, err := ceph.GetJobCostList(execConfig.Master, osdNum)
	if err != nil {
		return err
	}

	if len(jobCostList) == 0 {
		cmdStr := "ceph daemon osd." + strconv.Itoa(int(osdNum)) + " config set osd_op_queue_mclock_recov_lim " + strconv.Itoa(int(execConfig.MaxLimit))
		_, err = execConfig.Master.ExecCmd(cmdStr)
		return err
	}

	var coefficients []float64
	for _, jobCost := range jobCostList {
		if jobCost.ActualCost < 1 {
			// 忽略命中缓存的数据
			continue
		}
		if jobCost.Type == "write" {
			aStr := strings.Split(execConfig.WriteStdCoefficient, "x")[0]
			aFloat, _ := strconv.ParseFloat(aStr, 64)
			bStr := strings.Split(execConfig.WriteStdCoefficient, "x")[1]
			bFloat, _ := strconv.ParseFloat(bStr, 64)
			expectCost := aFloat*jobCost.ExpectCost + bFloat
			coefficient := jobCost.ActualCost / expectCost
			coefficients = append(coefficients, coefficient)
		} else {
			aStr := strings.Split(execConfig.ReadStdCoefficient, "x")[0]
			aFloat, _ := strconv.ParseFloat(aStr, 64)
			bStr := strings.Split(execConfig.ReadStdCoefficient, "x")[1]
			bFloat, _ := strconv.ParseFloat(bStr, 64)
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

	if avgCoefficient > execConfig.Zoom {
		// 降低limit
		execConfig.LastLimit = math.Max(execConfig.LastLimit-float64(50), execConfig.MinLimit)
	} else {
		// 增大limit
		execConfig.LastLimit = math.Min(execConfig.LastLimit+float64(5), execConfig.MaxLimit)
	}

	logrus.Infof("execConfig:%+v, "+
		"avgCoefficient:%0.2f,"+
		"limit:%0.2f",
		execConfig, avgCoefficient, execConfig.LastLimit)

	cmdStr := "ceph daemon osd." + strconv.Itoa(int(osdNum)) + " config set osd_op_queue_mclock_recov_lim " + strconv.Itoa(int(execConfig.LastLimit))
	_, err = execConfig.Master.ExecCmd(cmdStr)
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

	execConfig.Cluster = cluster
	execConfig.CephConf, err = ceph.NewCephConf(execConfig.Master, execConfig.OsdNum)
	if err != nil {
		return
	}

	var wg sync.WaitGroup
	for _, osdNum := range execConfig.OsdNum {
		itemOsdNum := osdNum
		wg.Add(1)
		go func() {
			for {
				_ = execConfig.Run(itemOsdNum)
				time.Sleep(time.Second)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}
