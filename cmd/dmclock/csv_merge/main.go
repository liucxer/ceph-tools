package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"strconv"
	"strings"
)

type ResultKey struct {
	DiskType      string  `json:"diskType"`
	OpType        string  `json:"opType"`
	BlockSize     string  `json:"blockSize"`
	IoDepth       int64   `json:"ioDepth"`
	RecoveryLimit float64 `json:"recoveryLimit"`
}
type ResultValue struct {
	ExpectCost float64 `json:"expectCost"`
	ActualCost float64 `json:"actualCost"`
	ReadIops   float64 `json:"readIops"`
	WriteIops  float64 `json:"writeIops"`
}

type Result struct {
	ResultKey
	ResultValue
}

func DataCost(costFile string) ([]Result, error) {
	var (
		res []Result
		err error
		bts []byte
	)

	bts, err = ioutil.ReadFile(costFile)
	if err != nil {
		logrus.Errorf("ioutil.ReadFile err:%v", err)
		return res, err
	}

	list := strings.Split(string(bts), "\n")
	for _, line := range list {
		if line == "" {
			continue
		}
		items := strings.Split(line, ",")
		if items[0] == "diskType" {
			continue
		}

		var resItem Result

		resItem.DiskType = items[0]
		resItem.OpType = items[2]
		resItem.BlockSize = items[5]

		ioDepth, err := strconv.Atoi(items[6])
		if err != nil {
			logrus.Errorf("strconv.Atoi, err:%v", err)
			return nil, err
		}
		resItem.IoDepth = int64(ioDepth)
		resItem.RecoveryLimit, err = strconv.ParseFloat(strings.Trim(items[7], " "), 64)
		if err != nil {
			logrus.Errorf("strconv.ParseFloat, err:%v", err)
			return nil, err
		}

		if items[7] == "NaN" || items[8] == "NaN" || items[9] == "NaN" {
			continue
		}

		expectCostStr := strings.Trim(strings.Trim(items[8], " "), "\r")
		resItem.ExpectCost, err = strconv.ParseFloat(expectCostStr, 64)
		if err != nil {
			logrus.Errorf("strconv.ParseFloat, err:%v", err)
			return nil, err
		}
		actualCostStr := strings.Trim(strings.Trim(items[9], " "), "\r")
		resItem.ActualCost, err = strconv.ParseFloat(actualCostStr, 64)
		if err != nil {
			logrus.Errorf("strconv.ParseFloat, err:%v", err)
			return nil, err
		}
		res = append(res, resItem)
	}

	return res, err
}

func DataIops(iopsFile string) ([]Result, error) {
	var (
		res []Result
		err error
		bts []byte
	)

	bts, err = ioutil.ReadFile(iopsFile)
	if err != nil {
		logrus.Errorf("ioutil.ReadFile err:%v", err)
		return res, err
	}

	list := strings.Split(string(bts), "\n")
	for _, line := range list {
		if line == "" {
			continue
		}
		items := strings.Split(line, ",")
		if items[0] == "diskType" || items[0] == "DiskType" {
			continue
		}

		var resItem Result

		resItem.DiskType = items[0]
		resItem.OpType = items[1]
		resItem.BlockSize = items[2]

		ioDepth, err := strconv.Atoi(items[3])
		if err != nil {
			logrus.Errorf("strconv.Atoi, err:%v", err)
			return nil, err
		}
		resItem.IoDepth = int64(ioDepth)
		resItem.RecoveryLimit, err = strconv.ParseFloat(strings.Trim(items[4], " "), 64)
		if err != nil {
			logrus.Errorf("strconv.ParseFloat, err:%v", err)
			return nil, err
		}

		//if items[7] == "NaN" || items[8] == "NaN" || items[9] == "NaN" {
		//	continue
		//}

		readIopsStr := strings.Trim(strings.Trim(items[5], " "), "\r")
		if readIopsStr != "" {
			resItem.ReadIops, err = strconv.ParseFloat(readIopsStr, 64)
			if err != nil {
				logrus.Errorf("strconv.ParseFloat, err:%v", err)
				return nil, err
			}
		}

		writeCostStr := strings.Trim(strings.Trim(items[10], " "), "\r")
		if writeCostStr != "" {
			resItem.WriteIops, err = strconv.ParseFloat(writeCostStr, 64)
			if err != nil {
				logrus.Errorf("strconv.ParseFloat, err:%v", err)
				return nil, err
			}
		}

		res = append(res, resItem)
	}

	return res, err
}

func DataMerge(costData, IopsData []Result) []Result {
	costDataMap := map[ResultKey]ResultValue{}
	iopsMap := map[ResultKey]ResultValue{}
	for _, v := range costData {
		resultKey := ResultKey{
			DiskType:      v.DiskType,
			OpType:        v.OpType,
			BlockSize:     v.BlockSize,
			IoDepth:       v.IoDepth,
			RecoveryLimit: v.RecoveryLimit,
		}

		resultValue := ResultValue{
			ExpectCost: v.ExpectCost,
			ActualCost: v.ActualCost,
		}
		costDataMap[resultKey] = resultValue
	}

	for _, v := range IopsData {
		resultKey := ResultKey{
			DiskType:      v.DiskType,
			OpType:        v.OpType,
			BlockSize:     v.BlockSize,
			IoDepth:       v.IoDepth,
			RecoveryLimit: v.RecoveryLimit,
		}

		resultValue := ResultValue{
			ReadIops:  v.ReadIops,
			WriteIops: v.WriteIops,
		}
		iopsMap[resultKey] = resultValue
	}

	for key, value := range iopsMap {
		if v, ok := costDataMap[key]; ok {
			value.ActualCost = v.ActualCost
			value.ExpectCost = v.ExpectCost
		}
		iopsMap[key] = value
	}

	var res []Result
	for key, value := range iopsMap {
		if value.ActualCost != 0 && value.ExpectCost != 0 &&
			(value.WriteIops != 0 || value.ReadIops != 0) {
			item := Result{}
			item.ResultKey = key
			item.ResultValue = value
			res = append(res, item)
		}
	}
	return res
}

type FioResultList []Result

func (list FioResultList) ToCsv() string {
	var res = ""
	header := "diskType, opType,blockSize, ioDepth, recoveryLimit," +
		"expectCost,actualCost,readIops,writeIops"
	res = res + header + "\n"
	for _, item := range list {
		itemStr := fmt.Sprintf("%s,%s,%s,%d,%f,%f,%f,%f,%f",
			item.DiskType,
			item.OpType,
			item.BlockSize,
			item.IoDepth,
			item.RecoveryLimit,
			item.ExpectCost,
			item.ActualCost,
			item.ReadIops,
			item.WriteIops,
		)
		res = res + itemStr + "\n"
	}

	return res
}

func main() {
	costResult, err := DataCost("/Users/liucx/Documents/gopath/src/github.com/liucxer/ceph-tools/cmd/dmclock/csv_merge/data/cost.csv")
	if err != nil {
		return
	}
	iopsResult, err := DataIops("/Users/liucx/Documents/gopath/src/github.com/liucxer/ceph-tools/cmd/dmclock/csv_merge/data/ipos.csv")
	res := DataMerge(costResult, iopsResult)

	list := FioResultList(res)

	fmt.Println(list.ToCsv())
}
