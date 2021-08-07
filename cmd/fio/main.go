package main

import (
	"encoding/json"
	"fmt"
	"github.com/liucxer/ceph-tools/pkg/cluster_client"
	"github.com/liucxer/ceph-tools/pkg/fio"
	"github.com/liucxer/confmiddleware/conflogger"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
)

type FioConfig struct {
	DiskType       string  `json:"diskType"`
	Runtime        int64   `json:"runtime"`
	OpType         string  `json:"opType"`
	DataPool       string  `json:"dataPool"`
	DataVolume     string  `json:"dataVolume"`
	BlockSize      string  `json:"blockSize"`
	IoDepth        int64   `json:"ioDepth"`
}

type FioResult struct {
	FioConfig  *FioConfig
	ExpectCost float64
	ActualCost float64
	ReadIops   float64
	WriteIops  float64
}

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

type FioResultList []FioResult
func (list FioResultList) ToCsv() string {
	var res = ""
	header := "diskType,runtime,opType,pool,volume,blockSize,ioDepth,readIops,writeIops"
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

type ExecConfig struct {
	DiskType   string   `json:"diskType"`
	IpAddr     string   `json:"ipAddr"`
	Runtime    int64    `json:"runtime"`
	DataPool   string   `json:"dataPool"`
	DataVolume string   `json:"dataVolume"`
	OpType     []string `json:"opType"`
	BlockSize  []string `json:"blockSize"`
	IoDepth    []int64  `json:"ioDepth"`

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
		Name:  "fio",
		Level: "Debug",
	}
	logger.SetDefaults()
	logger.Init()
}

func (execConfig *ExecConfig) RunOneJob(fioConfig *FioConfig) (*FioResult, error) {
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
	fioResult, err := fioObject.Exec(execConfig.cluster.Master)
	if err != nil {
		return nil, err
	}

	res.FioConfig = fioConfig
	res.ReadIops = fioResult.ReadIops
	res.WriteIops = fioResult.WriteIops

	logrus.Infof("RunOneJob res:%s", res.ToCsv())
	return &res, err
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
	fioResultList, err := execConfig.Run()
	if err != nil {
		return
	}

	csv := fioResultList.ToCsv()

	fmt.Println(csv)
}