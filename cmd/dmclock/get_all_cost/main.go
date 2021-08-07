package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/liucxer/ceph-tools/cmd/dmclock/log_analyze"
	"github.com/liucxer/ceph-tools/pkg/ceph"
	"github.com/liucxer/ceph-tools/pkg/cluster_client"
	"github.com/liucxer/ceph-tools/pkg/fio"
	"github.com/liucxer/ceph-tools/pkg/host_client"
	"github.com/liucxer/confmiddleware/conflogger"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"strconv"
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

type FioResult struct {
	FioConfig  *FioConfig
	ExpectCost float64
	ActualCost float64
	ReadIops   float64
	WriteIops  float64
}

type FioResultList []FioResult

func (item FioResult) ToCsv() string {
	itemStr := fmt.Sprintf("%s,%d,%s,%s,%s,%s,%d,%f,%f,%f,%f,%f",
		item.FioConfig.DiskType,
		item.FioConfig.Runtime,
		item.FioConfig.OpType,
		item.FioConfig.DataPool,
		item.FioConfig.DataVolume,
		item.FioConfig.BlockSize,
		item.FioConfig.IoDepth,
		item.FioConfig.RecoveryLimit,
		item.ExpectCost,
		item.ActualCost,
		item.ReadIops,
		item.WriteIops,
	)

	return itemStr
}

func (list FioResultList) ToCsv() string {
	var res = ""
	header := "diskType, runtime, opType, pool, volume, blockSize, ioDepth, recoveryLimit" +
		"expectCost,actualCost,readIops,writeIops"
	res = res + header + "\n"
	for _, item := range list {
		itemStr := fmt.Sprintf("%s,%d,%s,%s,%s,%s,%d,%f,%f,%f,%f,%f",
			item.FioConfig.DiskType,
			item.FioConfig.Runtime,
			item.FioConfig.OpType,
			item.FioConfig.DataPool,
			item.FioConfig.DataVolume,
			item.FioConfig.BlockSize,
			item.FioConfig.IoDepth,
			item.FioConfig.RecoveryLimit,
			item.ExpectCost,
			item.ActualCost,
			item.ReadIops,
			item.WriteIops,
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

	fioObject := fio.Fio{
		OpType:    fioConfig.OpType,
		Runtime:   fioConfig.Runtime,
		BlockSize: fioConfig.BlockSize,
		IoDepth:   fioConfig.IoDepth,
		Pool:      fioConfig.DataPool,
		RbdName:   fioConfig.DataVolume,
	}
	fioResult, err := fioObject.Exec(c.Master)
	if err != nil {
		return nil, err
	}

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
	res.ExpectCost = dmClockJobList.ExpectCost()
	res.ActualCost = dmClockJobList.ActualCost()
	res.ReadIops = fioResult.ReadIops
	res.WriteIops = fioResult.WriteIops
	return &res, err
}

func (conf *FioConfig) WaitOsdClean(c *cluster_client.Cluster) error {
	var (
		err error
	)
	for _, osdNum := range conf.OsdNum {
		_, err = globalConfig.OsdMap[osdNum].ExecCmd("systemctl restart ceph-osd@" + strconv.Itoa(int(osdNum)))
		if err != nil {
			return err
		}
	}

	// 设置limit 最大, 这样recovery恢复最快
	for _, osdNum := range conf.OsdNum {
		_, err = globalConfig.OsdMap[osdNum].ExecCmd("ceph daemon osd." + strconv.Itoa(int(osdNum)) + " config set osd_op_queue_mclock_recov_lim 99999")
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
			osdPerf, err := globalConfig.OsdMap[osdNum].OSDPerf("osd." + strconv.Itoa(int(osdNum)))
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

	cluster *cluster_client.Cluster
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

func init() {
	var logger = conflogger.Log{
		Name:  "getAllCost",
		Level: "Debug",
	}
	logger.SetDefaults()
	logger.Init()
}

func (execConfig *ExecConfig) RunOneJob(fioConfig *FioConfig) (*FioResult, error) {
	var (
		err error
		res *FioResult
	)
	// 等待对应的osd clean
	err = fioConfig.WaitOsdClean(execConfig.cluster)
	if err != nil {
		return res, err
	}

	// 设置limit
	for _, osdNum := range execConfig.OsdNum {
		_, err = globalConfig.OsdMap[osdNum].ExecCmd("ceph daemon osd." + strconv.Itoa(int(osdNum)) + " config set osd_op_queue_mclock_recov_lim " + fmt.Sprintf("%f", fioConfig.RecoveryLimit))
		if err != nil {
			return res, err
		}
	}

	time.Sleep(10 * time.Second)
	_, err = execConfig.cluster.Master.ExecCmd("ceph osd pool set " + execConfig.RecoveryPool + " size 2")
	if err != nil {
		return res, err
	}

	// 开始执行任务
	res, err = fioConfig.Exec(execConfig.cluster)
	if err != nil {
		return res, err
	}

	fmt.Println("RunOneJob res:", res.ToCsv())
	return res, err
}

func (execConfig *ExecConfig) Run() (*FioResultList, error) {
	var (
		err           error
		fioResultList FioResultList
	)
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
					bsRes, err := execConfig.RunOneJob(fioConfig)
					if err != nil {
						logrus.Errorf("fioConfig Result:%+v, failure, err:%v", fioConfig, err)
					} else {
						logrus.Warningf("fioConfig Result:%+v, success", fioConfig)
						fioResultList = append(fioResultList, *bsRes)
					}
				}
			}
		}
	}
	return &fioResultList, err
}

type GlobalConfig struct {
	OsdMap map[int64]*host_client.HostClient
}

var (
	globalConfig = GlobalConfig{}
)

func InitGlobalConfig(cluster *cluster_client.Cluster, osdNums []int64) error {
	var (
		err error
	)

	globalConfig.OsdMap = map[int64]*host_client.HostClient{}
	for _, osdNum := range osdNums {
		ip, err := ceph.GetOSDIp(cluster.Master, osdNum)
		if err != nil {
			return err
		}
		hostClient, err := host_client.NewHostClient(ip)
		if err != nil {
			return err
		}
		globalConfig.OsdMap[osdNum] = hostClient
	}

	for key, osd := range globalConfig.OsdMap {
		fmt.Println(key, osd.IpAddr)
	}
	return err
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

	// 连接集群
	cluster, err := cluster_client.NewCluster(execConfig.IpAddr)
	if err != nil {
		return
	}
	defer func() { _ = cluster.Close() }()

	err = InitGlobalConfig(cluster, execConfig.OsdNum)
	if err != nil {
		return
	}

	execConfig.cluster = cluster
	fioResultList, err := execConfig.Run()
	if err != nil {
		return
	}

	csv := fioResultList.ToCsv()

	fmt.Println(csv)
}
