package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/liucxer/ceph-tools/cmd/dmclock/log_analyze"
	"github.com/liucxer/ceph-tools/pkg/cluster_client"
	"github.com/liucxer/confmiddleware/conflogger"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"strconv"
	"sync"
	"time"
)

type FioConfig struct {
	OsdNum         []int64 `json:"osdNum"`
	DiskType       string  `json:"diskType"`
	Runtime        int64   `json:"runtime"`
	OpType         string  `json:"opType"`
	DataPool       string  `json:"dataPool"`
	RecoveryPool   string  `json:"recoveryPool"`
	DataVolume     string  `json:"dataVolume"`
	RecoveryVolume string  `json:"recoveryVolume"`
	BlockSize      string  `json:"blockSize"`
	IoDepth        int64   `json:"ioDepth"`
	RecoveryLimit  float64 `json:"recoveryLimit"`
}

func (exec *FioConfig) Config() string {
	cfgData := `[global]
ioengine=rbd
clientname=admin
invalidate=0
time_based
direct=1
group_reporting

[write]
rw=` + exec.OpType + `
runtime=` + strconv.Itoa(int(exec.Runtime)) + `
bs=` + exec.BlockSize + `
iodepth=` + strconv.Itoa(int(exec.IoDepth)) + `
pool=` + exec.DataPool + `
rbdname=` + exec.DataVolume
	return cfgData
}
func (conf *FioConfig) ConfigFileName() string {
	return uuid.New().String() + "_" + conf.DiskType + "_" +
		conf.OpType + "_" +
		conf.DataPool + "_" +
		conf.DataVolume + "_" +
		conf.BlockSize + "_" +
		strconv.Itoa(int(conf.IoDepth)) + "_" +
		fmt.Sprintf("%f", conf.RecoveryLimit) + ".conf"
}

type FioResult struct {
	FioConfig      *FioConfig
	DMClockJobList *log_analyze.DMClockJobList
}

type FioResultList []FioResult

func (list FioResultList) ToCsv() string {
	var res = ""
	header := "diskType, runtime, opType, pool, volume, blockSize, ioDepth, recoveryLimit" +
		"expectCost,actualCost,"
	res = res + header + "\n"
	for _, item := range list {
		itemStr := fmt.Sprintf("%s,%d,%s,%s,%s,%s,%d,%f ,%f, %f",
			item.FioConfig.DiskType,
			item.FioConfig.Runtime,
			item.FioConfig.OpType,
			item.FioConfig.DataPool,
			item.FioConfig.DataVolume,
			item.FioConfig.BlockSize,
			item.FioConfig.IoDepth,
			item.FioConfig.RecoveryLimit,
			item.DMClockJobList.ExpectCost(),
			item.DMClockJobList.ActualCost(),
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
	err = c.ClearOsdLog(fioConfig.OsdNum)
	if err != nil {
		return nil, err
	}
	tmpDir := os.TempDir() + uuid.NewString() + "ceph"

	err = os.Mkdir(tmpDir, os.ModePerm)
	if err != nil {
		logrus.Errorf("os.Mkdir err. [err:%v, localLogDir:%s]", err, tmpDir)
		return nil, err
	}
	defer func() {
		err = os.RemoveAll(tmpDir)
		if err != nil {
			logrus.Errorf("os.RemoveAll err. [err:%v, tmpDir:%s]", err, tmpDir)
		}
	}()

	// 创建配置文件
	bsFilePath := fioConfig.ConfigFileName()
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

	wg.Wait()

	// 收集日志
	err = c.CollectOsdLog(tmpDir, fioConfig.OsdNum)
	if err != nil {
		return nil, err
	}

	// 统计分析
	dmClockJobList, err := log_analyze.LogAnalyze(tmpDir)
	if err != nil {
		return nil, err
	}

	res.FioConfig = fioConfig
	res.DMClockJobList = dmClockJobList
	return &res, err
}

func (conf *FioConfig) WaitOsdClean(c *cluster_client.Cluster) error {
	var (
		err error
	)

	for _, osdNum := range conf.OsdNum {
		_, err = c.Master.ExecCmd("systemctl restart ceph-osd@" + strconv.Itoa(int(osdNum)))
		if err != nil {
			return err
		}
	}

	// 设置limit 最大, 这样recovery恢复最快
	for _, osdNum := range conf.OsdNum {
		_, err = c.Master.ExecCmd("ceph daemon osd." + strconv.Itoa(int(osdNum)) + " config set osd_op_queue_mclock_recov_lim 99999")
		if err != nil {
			return err
		}
	}
	_, err = c.Master.ExecCmd("ceph osd pool set " + conf.RecoveryPool + " size 1")
	if err != nil {
		return err
	}

	count := 0
	for {
		time.Sleep(5 * time.Second)

		// 等待OSDClean
		for _, osdNum := range conf.OsdNum {
			osdStatus, err := c.OsdStatus(osdNum)
			if err != nil {
				return err
			}
			if !osdStatus.ActiveClean {
				count = 0
				continue
			}
		}

		// 等待数据删除
		for _, osdNum := range conf.OsdNum {
			osdPerf, err := c.CurrentOSDPerf("osd." + strconv.Itoa(int(osdNum)))
			if err != nil {
				return err
			}

			if osdPerf.OSD.NumpgRemoving != 0 {
				count = 0
				continue
			}
			if osdPerf.OSD.NumpgStray != 0 {
				count = 0
				continue
			}
		}
		count++
		if count > 3 {
			break
		}
	}

	return nil
}

type ExecConfig struct {
	DiskType string   `json:"diskType"`
	IpAddr   []string `json:"ipAddr"`
	OsdNum   []int64  `json:"osdNum"`

	Runtime        int64  `json:"runtime"`
	DataPool       string `json:"dataPool"`
	RecoveryPool   string `json:"recoveryPool"`
	DataVolume     string `json:"dataVolume"`
	RecoveryVolume string `json:"recoveryVolume"`

	OpType        []string  `json:"opType"`
	BlockSize     []string  `json:"blockSize"`
	IoDepth       []int64   `json:"ioDepth"`
	RecoveryLimit []float64 `json:"recoveryLimit"`
}

func ReadExecConfig(configFilePath string) (*ExecConfig, error) {
	bts, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		logrus.Errorf("ioutil.ReadFile err:%v", err)
		return nil, err
	}

	var conf ExecConfig
	err = json.Unmarshal(bts, &conf)
	if err != nil {
		logrus.Errorf("json.Unmarshal err:%v", err)
		return nil, err
	}
	return &conf, err
}

func init() {
	var logger = conflogger.Log{
		Name:  "getAllCost",
		Level: "Debug",
	}
	logger.SetDefaults()
	logger.Init()
}

func RunOneJob(cluster *cluster_client.Cluster, execConfig *ExecConfig, fioConfig *FioConfig) (*FioResult, error) {
	var (
		err error
		res *FioResult
	)
	// 等待对应的osd clean
	err = fioConfig.WaitOsdClean(cluster)
	if err != nil {
		return res, err
	}

	// 设置limit
	for _, osdNum := range execConfig.OsdNum {
		_, err = cluster.Master.ExecCmd("ceph daemon osd." + strconv.Itoa(int(osdNum)) + " config set osd_op_queue_mclock_recov_lim " + fmt.Sprintf("%f", fioConfig.RecoveryLimit))
		if err != nil {
			return res, err
		}
	}

	_, err = cluster.Master.ExecCmd("ceph osd pool set " + execConfig.RecoveryPool + " size 2")
	if err != nil {
		return res, err
	}

	// 开始执行任务
	res, err = fioConfig.Exec(cluster)
	if err != nil {
		return res, err
	}

	logrus.Warningf("fioConfig Result:%+v, ExpectCost:%f, ActualCost:%f", *res.FioConfig,
		(*res).DMClockJobList.ExpectCost(),
		(*res).DMClockJobList.ActualCost(),
	)
	return res, err
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage:\n     ./cmd config.json")
		return
	}

	execConfig, err := ReadExecConfig(os.Args[1])
	if err != nil {
		return
	}

	// 连接集群
	cluster, err := cluster_client.NewCluster(execConfig.IpAddr)
	if err != nil {
		return
	}
	defer func() { _ = cluster.Close() }()

	var fioResultList FioResultList
	for _, opType := range execConfig.OpType {
		for _, blockSize := range execConfig.BlockSize {
			for _, ioDepth := range execConfig.IoDepth {
				for _, recoveryLimit := range execConfig.RecoveryLimit {
					fioConfig := &FioConfig{
						DiskType:       execConfig.DiskType,
						Runtime:        execConfig.Runtime,
						OpType:         opType,
						RecoveryPool:   execConfig.RecoveryPool,
						DataPool:       execConfig.DataPool,
						DataVolume:     execConfig.DataVolume,
						RecoveryVolume: execConfig.RecoveryVolume,
						BlockSize:      blockSize,
						IoDepth:        ioDepth,
						RecoveryLimit:  recoveryLimit,
						OsdNum:         execConfig.OsdNum,
					}
					bsRes, err := RunOneJob(cluster, execConfig, fioConfig)
					if err != nil {
						logrus.Warningf("fioConfig Result:%+v, failure, err:%v", fioConfig, err)
					} else {
						logrus.Warningf("fioConfig Result:%+v, success", fioConfig)
						fioResultList = append(fioResultList, *bsRes)
					}
				}
			}
		}
	}

	csv := fioResultList.ToCsv()

	fmt.Println(csv)
}
