package fio

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/liucxer/ceph-tools/pkg/interfacer"
	"strconv"
)

type Fio struct {
	OpType    string `json:"opType"`
	Runtime   int64  `json:"runtime"`
	BlockSize string `json:"blockSize"`
	IoDepth   int64  `json:"ioDepth"`
	Pool      string `json:"pool"`
	RbdName   string `json:"rbdName"`
}

func (conf *Fio) Config() string {
	cfgData := `[global]
ioengine=rbd
clientname=admin
invalidate=0
time_based
direct=1
group_reporting

[fioJob]
rw=` + conf.OpType + `
runtime=` + strconv.Itoa(int(conf.Runtime)) + `
bs=` + conf.BlockSize + `
iodepth=` + strconv.Itoa(int(conf.IoDepth)) + `
pool=` + conf.Pool + `
rbdname=` + conf.RbdName
	return cfgData
}

func (conf *Fio) ConfigFileName() string {
	return uuid.New().String() + "_" +
		conf.OpType + "_" +
		conf.Pool + "_" +
		conf.RbdName + "_" +
		conf.BlockSize + "_" +
		strconv.Itoa(int(conf.IoDepth)) + ".conf"
}

type fioOpItem struct {
	BW    float64 `json:"bw"`
	Iops  float64 `json:"iops"`
	LatNS struct {
		Mean float64 `json:"mean"`
	} `json:"lat_ns"`
	ClatNS struct {
		Mean       float64 `json:"mean"`
		Percentile struct {
			Key95 float64 `json:"95.000000"`
			Key99 float64 `json:"99.000000"`
		} `json:"percentile"`
	} `json:"clat_ns"`
}

type fioResp struct {
	Jobs []struct {
		Jobname string    `json:"jobname"`
		Read    fioOpItem `json:"read"`
		Write   fioOpItem `json:"write"`
	} `json:"jobs"`
}

type FioResultItem struct {
	Bandwidth float64 `json:"bandwidth"`
	Iops      float64 `json:"iops"`
	Clat      float64 `json:"clat"`
	Clat95    float64 `json:"clat95"`
	Clat99    float64 `json:"clat99"`
}

type FioResult struct {
	ReadBandwidth float64 `json:"readBandwidth"`
	ReadIops      float64 `json:"readIops"`
	ReadLat       float64 `json:"readLat"`
	ReadClat95    float64 `json:"readClat95"`
	ReadClat99    float64 `json:"readClat99"`

	WriteBandwidth float64 `json:"writeBandwidth"`
	WriteIops      float64 `json:"writeIops"`
	WriteLat       float64 `json:"writeLat"`
	WriteClat95    float64 `json:"writeClat95"`
	WriteClat99    float64 `json:"writeClat99"`
}

func (conf *Fio) Exec(worker interfacer.Worker) (*FioResult, error) {
	var (
		res FioResult
		err error
	)

	// 创建配置文件
	bsFilePath := conf.ConfigFileName()
	_, err = worker.ExecCmd("echo '" + conf.Config() + "' > " + bsFilePath)
	if err != nil {
		return nil, err
	}
	defer func() {
		// 删除配置文件
		_, err = worker.ExecCmd("rm " + bsFilePath)
		if err != nil {
			return
		}
	}()

	bts, err := worker.ExecCmd("fio " + bsFilePath + " --output-format=json")
	if err != nil {
		return nil, err
	}

	fioRes := fioResp{}
	err = json.Unmarshal(bts, &fioRes)
	if err != nil {
		return nil, err
	}

	for _, item := range fioRes.Jobs {
		if item.Jobname != "fioJob" {
			continue
		}
		res.WriteIops, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", item.Write.Iops), 64)
		res.WriteBandwidth = item.Write.BW
		res.WriteLat, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", item.Write.LatNS.Mean/float64(1000000)), 64)
		res.WriteClat95, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", item.Write.ClatNS.Percentile.Key95/float64(1000000)), 64)
		res.WriteClat99, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", item.Write.ClatNS.Percentile.Key99/float64(1000000)), 64)
		res.ReadIops, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", item.Read.Iops), 64)
		res.ReadBandwidth = item.Read.BW
		res.ReadLat, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", item.Read.LatNS.Mean/float64(1000000)), 64)
		res.ReadClat95, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", item.Read.ClatNS.Percentile.Key95/float64(1000000)), 64)
		res.ReadClat99, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", item.Read.ClatNS.Percentile.Key99/float64(1000000)), 64)
	}

	return &res, err
}
