package main

import (
	"ceph-tools/cmd/dmclock/log_analyze"
	"ceph-tools/pkg/cluster_client"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"strconv"
	"sync"
	"time"
)

type FioConfig struct {
	RunTime int
	BS      string
	IODepth int
}

func (conf FioConfig) Config() string {
	cfgData := `[global]
ioengine=rbd
clientname=admin
invalidate=0
time_based
direct=1
group_reporting

[write]
rw=write
runtime=` + strconv.Itoa(int(conf.RunTime)) + `
bs=` + conf.BS + `
iodepth=` + strconv.Itoa(int(conf.IODepth)) + `
pool=bd_pool
rbdname=image1`
	return cfgData
}

type FioResult struct {
	FioConfig      *FioConfig
	CephStatusList *cluster_client.CephStatusList
	DMClockJobList *log_analyze.DMClockJobList
}

type FioResultList []FioResult

func (list FioResultList) ToCsv() string {
	var res = ""
	header := "runTime, BS, IODepth," +
		"ExpectCost,ActualCost," +
		"WriteOpPerSec"
	res = res + header + "\n"
	for _, item := range list {
		itemStr := fmt.Sprintf("%d, %s, %d, %f, %f, %f",
			item.FioConfig.RunTime,
			item.FioConfig.BS,
			item.FioConfig.IODepth,
			item.DMClockJobList.ExpectCost(),
			item.DMClockJobList.ActualCost(),
			item.CephStatusList.WriteBytesSec(),
		)
		res = res + itemStr + "\n"
	}

	return res
}

func  ExecFio (c *cluster_client.Cluster, fioConfig *FioConfig) (*FioResult, error) {
	var (
		err error
		res FioResult
	)
	// 清空日志
	err = c.ClearCephLog()
	if err != nil {
		return nil, err
	}
	localLogDir := os.TempDir() + "ceph"
	err = os.RemoveAll(localLogDir)
	if err != nil {
		logrus.Errorf("os.RemoveAll err. [err:%v, localLogDir:%s]", err, localLogDir)
		return nil, err
	}

	err = os.Mkdir(localLogDir, os.ModePerm)
	if err != nil {
		logrus.Errorf("os.Mkdir err. [err:%v, localLogDir:%s]", err, localLogDir)
		return nil, err
	}

	// 创建配置文件
	bsFilePath := "/home/bs_" + fioConfig.BS + ".cfg"
	_, err = c.Clients[0].Host.ExecCmd("touch " + bsFilePath)
	if err != nil {
		return nil, err
	}
	_, err = c.Clients[0].Host.ExecCmd("echo '" + fioConfig.Config() + "' > " + bsFilePath)
	if err != nil {
		return nil, err
	}

	defer func() {
		// 删除配置文件
		_, err = c.Clients[0].Host.ExecCmd("rm " + bsFilePath)
		if err != nil {
			return
		}
	}()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		// 执行fio命令
		_, _ = c.Clients[0].Host.ExecCmd("fio " + bsFilePath)
		wg.Done()
	}()
	wg.Add(1)
	var cephStatusList *cluster_client.CephStatusList
	go func() {
		time.Sleep(time.Second)
		// 获取集群状态信息
		cephStatusList, err = c.CephStatus(fioConfig.RunTime-2)
		wg.Done()
	}()
	wg.Wait()

	// 收集日志
	err = c.CollectCephLog(localLogDir)
	if err != nil {
		return nil, err
	}

	// 统计分析
	dmClockJobList, err := log_analyze.LogAnalyze(localLogDir)
	if err != nil {
		return nil, err
	}

	res.FioConfig = fioConfig
	res.DMClockJobList = dmClockJobList
	res.CephStatusList = cephStatusList
	return &res, err
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage:\n     ./detect ipaddr")
		return
	}
	ipAddr := os.Args[1]

	totalTime := 120 // 单位秒
	// 连接集群
	logrus.SetLevel(logrus.DebugLevel)
	cluster, err := cluster_client.NewCluster([]string{ipAddr})
	if err != nil {
		return
	}
	defer func() { _ = cluster.Close() }()

	bsList := []string{"1k", "2k", "4k", "8k", "16k", "32k", "64k", "128k", "256k", "512k", "1M", "2M", "4M"}
	iodepthList := []int{1, 2, 4, 8, 16, 32, 64}

	var fioResultList FioResultList
	for _, bs := range bsList {
		for _, iodepth := range iodepthList {
			fioConfig := FioConfig{
				RunTime: totalTime,
				BS:      bs, IODepth: iodepth,
			}
			bsRes, err := ExecFio(cluster, &fioConfig)
			if err != nil {
				return
			}
			fioResultList = append(fioResultList, *bsRes)
		}
	}


	for _ ,fioResult := range fioResultList {
		fmt.Println(fioResult.FioConfig,
			fioResult.DMClockJobList.ExpectCost(),
			fioResult.DMClockJobList.ActualCost(),
			fioResult.CephStatusList.Avg().WriteOpPerSec,
		)
	}
	csv := fioResultList.ToCsv()

	fmt.Println(csv)
}
