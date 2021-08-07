package main

import (
	"encoding/json"
	"fmt"
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
}

type FioResult struct {
	FioConfig  *FioConfig
	ReadIops   float64
	WriteIops  float64
}

type FioResultList []FioResult

func (item FioResult) ToCsv() string {
	itemStr := fmt.Sprintf("%s,%d,%s,%s,%s,%s,%d,%f,%f,",
		item.FioConfig.DiskType,
		item.FioConfig.Runtime,
		item.FioConfig.OpType,
		item.FioConfig.DataPool,
		item.FioConfig.DataVolume,
		item.FioConfig.BlockSize,
		item.FioConfig.IoDepth,
		item.ReadIops,
		item.WriteIops,
	)

	return itemStr
}

func (list FioResultList) ToCsv() string {
	var res = ""
	header := "diskType, runtime, opType, pool, volume, blockSize, ioDepth" +
		"readIops,writeIops"
	res = res + header + "\n"
	for _, item := range list {
		itemStr := fmt.Sprintf("%s,%d,%s,%s,%s,%s,%d,%f,%f",
			item.FioConfig.DiskType,
			item.FioConfig.Runtime,
			item.FioConfig.OpType,
			item.FioConfig.DataPool,
			item.FioConfig.DataVolume,
			item.FioConfig.BlockSize,
			item.FioConfig.IoDepth,
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

	res.FioConfig = fioConfig
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

	logrus.Infof("RunOneJob res:%s", res.ToCsv())
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
