package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/liucxer/ceph-tools/pkg/ceph"
	"github.com/liucxer/ceph-tools/pkg/cluster_client"
	"github.com/liucxer/ceph-tools/pkg/csv"
	"github.com/liucxer/ceph-tools/pkg/fio"
	"github.com/liucxer/confmiddleware/conflogger"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"strconv"
	"time"
)

type FioConfig struct {
	WithRecovery   bool   `json:"withRecovery"`
	RecoveryPool   string `json:"recoveryPool"`
	RecoveryVolume string `json:"recoveryVolume"`

	WithJobCost bool `json:"withJobCost"`

	DiskType   string `json:"diskType"`
	Runtime    int64  `json:"runtime"`
	DataPool   string `json:"dataPool"`
	DataVolume string `json:"dataVolume"`
	OpType     string `json:"opType"`
	BlockSize  string `json:"blockSize"`
	IoDepth    int64  `json:"ioDepth"`
}

type ExecResult struct {
	FioConfig
	fio.FioResult
	ExpectCost float64 `json:"expectCost"`
	ActualCost float64 `json:"actualCost"`
}

type ExecConfig struct {
	*cluster_client.Cluster
	*ceph.CephConf

	WithRecovery   bool   `json:"withRecovery"`
	RecoveryPool   string `json:"recoveryPool"`
	RecoveryVolume string `json:"recoveryVolume"`

	WithJobCost bool    `json:"withJobCost"`
	OsdNum      []int64 `json:"osdNum"`

	DiskType   string `json:"diskType"`
	IpAddr     string `json:"ipAddr"`
	Runtime    int64  `json:"runtime"`
	DataPool   string `json:"dataPool"`
	DataVolume string `json:"dataVolume"`

	OpType    []string `json:"opType"`
	BlockSize []string `json:"blockSize"`
	IoDepth   []int64  `json:"ioDepth"`
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
		Name:  "fio",
		Level: "Debug",
	}
	logger.SetDefaults()
	logger.Init()
}

func (execConfig *ExecConfig) RunOneJob(fioConfig *FioConfig) (*ExecResult, error) {
	var (
		err error
		res ExecResult
	)

	fioObject := fio.Fio{
		OpType:    fioConfig.OpType,
		Runtime:   fioConfig.Runtime,
		BlockSize: fioConfig.BlockSize,
		IoDepth:   fioConfig.IoDepth,
		Pool:      fioConfig.DataPool,
		RbdName:   fioConfig.DataVolume,
	}

	var totalCount = 0
	var totalExpectCost = float64(0)
	var totalActualCost = float64(0)
	ctx, cancelFn := context.WithCancel(context.Background())

	if execConfig.WithRecovery {
		// 等待对应的osd clean
		err = execConfig.WaitOsdClean()
		if err != nil {
			cancelFn()
			return &res, err
		}

		time.Sleep(10 * time.Second)
		_, err = execConfig.Master.ExecCmd("ceph osd pool set " + execConfig.RecoveryPool + " size 2")
		if err != nil {
			cancelFn()
			return &res, err
		}
	}

	if execConfig.WithJobCost {
		for _, osdNum := range execConfig.OsdNum {
			itemOsdNum := osdNum
			go func(ctx context.Context) {
				for {
					select {
					case <-ctx.Done():
						return
					default:
						jobCostList, _ := ceph.GetJobCostList(execConfig.Master, itemOsdNum)
						totalCount += len(jobCostList)
						totalExpectCost += jobCostList.TotalExpectCost()
						totalActualCost += jobCostList.TotalActualCost()
						time.Sleep(10 * time.Second)
					}
				}
			}(ctx)
		}
	}

	fioResult, err := fioObject.Exec(execConfig.Master)
	cancelFn()
	if err != nil {
		return nil, err
	}

	res.FioResult = *fioResult
	res.FioConfig = *fioConfig
	res.ExpectCost, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", totalExpectCost/float64(totalCount)), 64)
	res.ActualCost, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", totalActualCost/float64(totalCount)), 64)

	name, value, err := csv.ObjectToCsv(res)
	if err != nil {
		return nil, err
	}
	logrus.Infof("RunOneJob res:%+v", res)
	logrus.Infof("RunOneJob name:%s", name)
	logrus.Infof("RunOneJob value:%s", value)
	return &res, err
}

func (execConfig *ExecConfig) WaitOsdClean() error {
	var (
		err error
	)
	for _, osdNum := range execConfig.OsdNum {
		_, err = execConfig.OsdNumMap[osdNum].ExecCmd("systemctl restart ceph-osd@" + strconv.Itoa(int(osdNum)))
		if err != nil {
			return err
		}
	}

	// 设置limit 最大, 这样recovery恢复最快
	for _, osdNum := range execConfig.OsdNum {
		_, err = execConfig.OsdNumMap[osdNum].ExecCmd("ceph daemon osd." + strconv.Itoa(int(osdNum)) + " config set osd_op_queue_mclock_recov_lim 99999")
		if err != nil {
			return err
		}
	}

	_, err = execConfig.Master.ExecCmd("ceph osd pool set " + execConfig.RecoveryPool + " size 1")
	if err != nil {
		return err
	}

	count := 0
	for {
		time.Sleep(5 * time.Second)

		// 等待OSDClean
		for _, osdNum := range execConfig.OsdNum {
			osdStatus, err := execConfig.OsdStatus(osdNum)
			if err != nil {
				return err
			}
			if !osdStatus.ActiveClean {
				count = 0
				continue
			}
		}

		// 等待数据删除
		for _, osdNum := range execConfig.OsdNum {
			osdPerf, err := execConfig.OsdNumMap[osdNum].OSDPerf("osd." + strconv.Itoa(int(osdNum)))
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

func (execConfig *ExecConfig) Run() (*[]ExecResult, error) {
	var (
		err           error
		fioResultList []ExecResult
	)
	for _, opType := range execConfig.OpType {
		for _, blockSize := range execConfig.BlockSize {
			for _, ioDepth := range execConfig.IoDepth {
				fioConfig := &FioConfig{
					WithRecovery:   execConfig.WithRecovery,
					RecoveryPool:   execConfig.RecoveryPool,
					RecoveryVolume: execConfig.RecoveryVolume,
					WithJobCost:    execConfig.WithJobCost,
					DiskType:       execConfig.DiskType,
					Runtime:        execConfig.Runtime,
					OpType:         opType,
					DataPool:       execConfig.DataPool,
					DataVolume:     execConfig.DataVolume,
					BlockSize:      blockSize,
					IoDepth:        ioDepth,
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

func NewExecConfig(configPath string) (*ExecConfig, error) {
	execConfig := ExecConfig{}
	err := execConfig.ReadConfig(configPath)
	if err != nil {
		return &execConfig, nil
	}

	execConfig.Cluster, err = cluster_client.NewCluster([]string{execConfig.IpAddr})
	if err != nil {
		return &execConfig, nil
	}
	//defer func() { _ = execConfig.Cluster.Close() }()

	execConfig.CephConf, err = ceph.NewCephConf(execConfig.Master, execConfig.OsdNum)
	if err != nil {
		return &execConfig, nil
	}

	logrus.Debugf("NewExecConfig. execConfig:%+v", execConfig)
	return &execConfig, nil
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage:\n     ./cmd config.json")
		return
	}

	execConfig, err := NewExecConfig(os.Args[1])
	if err != nil {
		return
	}

	fioResultList, err := execConfig.Run()
	if err != nil {
		return
	}

	res, err := csv.ObjectListToCsv(*fioResultList)
	if err != nil {
		return
	}

	fmt.Println(res)
}
