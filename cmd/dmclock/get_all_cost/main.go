package main

import (
	"fmt"
	"github.com/liucxer/ceph-tools/cmd/dmclock/log_analyze"
	"github.com/liucxer/ceph-tools/pkg/cluster_client"
	"github.com/sirupsen/logrus"
	"os"
	"strconv"
	"sync"
	"time"
)

type FioConfig struct {
	Limit   float64
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
pool=data_pool
rbdname=foo1`
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
	header := "runTime, BS, IODepth, recovery, " +
		"ExpectCost,ActualCost," +
		"WriteOpPerSec"
	res = res + header + "\n"
	for _, item := range list {
		itemStr := fmt.Sprintf("%d, %s, %d, %f, %f, %f, %f",
			item.FioConfig.RunTime,
			item.FioConfig.BS,
			item.FioConfig.IODepth,
			item.FioConfig.Limit,
			item.DMClockJobList.ExpectCost(),
			item.DMClockJobList.ActualCost(),
			item.CephStatusList.WriteOpPerSec(),
		)
		res = res + itemStr + "\n"
	}

	return res
}

func (fioConfig *FioConfig) Exec(c *cluster_client.Cluster) (*FioResult, error) {
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
	_, err = c.Master.ExecCmd("touch " + bsFilePath)
	if err != nil {
		return nil, err
	}
	_, err = c.Master.ExecCmd("echo '" + fioConfig.Config() + "' > " + bsFilePath)
	if err != nil {
		return nil, err
	}

	defer func() {
		// 删除配置文件
		_, err = c.Master.ExecCmd("rm " + bsFilePath)
		if err != nil {
			return
		}
	}()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		// 执行fio命令
		_, _ = c.Master.ExecCmd("fio " + bsFilePath)
		wg.Done()
	}()
	wg.Add(1)
	var cephStatusList *cluster_client.CephStatusList
	go func() {
		time.Sleep(time.Second)
		// 获取集群状态信息
		cephStatusList, err = c.CephStatus(fioConfig.RunTime - 2)
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

func WaitClusterClean(c *cluster_client.Cluster) error {

	_, err := c.Master.ExecCmd("ceph osd pool set rb_2disk_pool size 1")
	if err != nil {
		return err
	}
	_, err = c.Master.ExecCmd("ceph daemon osd.6 config set osd_op_queue_mclock_recov_lim 99999")
	if err != nil {
		return err
	}

	_, err = c.Master.ExecCmd("ceph daemon osd.11 config set osd_op_queue_mclock_recov_lim 99999")
	if err != nil {
		return err
	}
	count := 0
	for {
		time.Sleep(5 * time.Second)

		// 等待集群clean
		cephStatus, err := c.CurrentCephStatus()
		if err != nil {
			return err
		}
		if cephStatus.RecoveringBytesPerSec != 0 {
			continue
		}

		// 等待数据删除
		osdPerf, err := c.CurrentOSDPerf("osd.6")
		if err != nil {
			return err
		}
		if osdPerf.OSD.NumpgRemoving != 0 {
			continue
		}
		if osdPerf.OSD.NumpgStray != 0 {
			continue
		}

		// 等待数据删除
		osdPerf, err = c.CurrentOSDPerf("osd.11")
		if err != nil {
			return err
		}
		if osdPerf.OSD.NumpgRemoving != 0 {
			continue
		}
		if osdPerf.OSD.NumpgStray != 0 {
			continue
		}
		count++
		if count > 3 {
			break
		}
	}
	_, err = c.Master.ExecCmd("ceph osd pool set rb_2disk_pool size 2")
	if err != nil {
		return err
	}
	return nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage:\n     ./detect ipaddr")
		return
	}
	ipAddr := os.Args[1:]

	totalTime := 120 // 单位秒

	// 连接集群
	logrus.SetLevel(logrus.InfoLevel)
	cluster, err := cluster_client.NewCluster(ipAddr)
	if err != nil {
		return
	}
	defer func() { _ = cluster.Close() }()

	bsList := []string{"4k", "16k", "64k", "256k", "512k", "1M", "2M", "4M"}
	ioDepthList := []int{1, 4, 16, 64, 256, 512}

	//bsList := []string{"4M"}
	//ioDepthList := []int{1}
	recoveryLimits := []float64{0, 79, 158, 316, 500, 700, 999, 1250, 1500, 2000}

	var fioResultList FioResultList
	for _, bs := range bsList {
		for _, iodepth := range ioDepthList {
			for _, recoveryLimit := range recoveryLimits {
				err = WaitClusterClean(cluster)
				if err != nil {
					return
				}

				_, err = cluster.Master.ExecCmd("ceph daemon osd.6 config set osd_op_queue_mclock_recov_lim " + fmt.Sprintf("%f", recoveryLimit))
				if err != nil {
					return
				}

				_, err = cluster.Master.ExecCmd("ceph daemon osd.11 config set osd_op_queue_mclock_recov_lim " + fmt.Sprintf("%f", recoveryLimit))
				if err != nil {
					return
				}

				fioConfig := FioConfig{
					Limit:   recoveryLimit,
					RunTime: totalTime,
					BS:      bs,
					IODepth: iodepth,
				}
				bsRes, err := fioConfig.Exec(cluster)
				if err != nil {
					return
				}
				fmt.Println("-------------------", *bsRes.FioConfig,
					(*bsRes).DMClockJobList.ExpectCost(),
					(*bsRes).DMClockJobList.ActualCost(),
					(*bsRes).CephStatusList.WriteOpPerSec(),
				)
				fioResultList = append(fioResultList, *bsRes)
			}
		}
	}
	csv := fioResultList.ToCsv()

	fmt.Println(csv)
}
