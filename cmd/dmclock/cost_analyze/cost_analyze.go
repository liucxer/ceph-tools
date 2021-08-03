package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"math"
	"sort"
	"strconv"
	"strings"
)

type Result struct {
	DiskType      string  `json:"diskType"`
	OpType        string  `json:"opType"`
	BlockSize     string  `json:"blockSize"`
	IoDepth       int64   `json:"ioDepth"`
	RecoveryLimit float64 `json:"recoveryLimit"`
	ExpectCost    float64 `json:"expectCost"`
	ActualCost    float64 `json:"actualCost"`
}

type ResultList []Result

func (l ResultList) Swap(i, j int) {
	item := l[i]
	l[i] = l[j]
	l[j] = item
}

func (l ResultList) Less(i, j int) bool {
	if l[i].DiskType != l[j].DiskType {
		return l[i].DiskType < l[j].DiskType
	}

	if l[i].OpType != l[j].OpType {
		return l[i].OpType < l[j].OpType
	}

	if l[i].BlockSize != l[j].BlockSize {
		return l[i].BlockSize < l[j].BlockSize
	}

	if l[i].IoDepth != l[j].IoDepth {
		return l[i].IoDepth < l[j].IoDepth
	}
	return false
}

func (l ResultList) Len() int {
	return len(l)
}

type ResultKey struct {
	DiskType  string `json:"diskType"`
	OpType    string `json:"opType"`
	BlockSize string `json:"blockSize"`
	IoDepth   int64  `json:"ioDepth"`
}
type ResultValue struct {
	RecoveryLimit float64 `json:"recoveryLimit"`
	ExpectCost    float64 `json:"expectCost"`
	ActualCost    float64 `json:"actualCost"`
}

type ResultValueList []ResultValue

func (list ResultValueList) Best(toleranceScope float64) ResultValue {
	minResult := list[0]
	for _, item := range list {
		if item.ActualCost < minResult.ActualCost {
			minResult = item
		}
	}

	for _, item := range list {
		subData := item.ActualCost - minResult.ActualCost
		absData := math.Abs(subData / minResult.ActualCost)
		if absData < toleranceScope && item.RecoveryLimit > minResult.RecoveryLimit {
			minResult.RecoveryLimit = item.RecoveryLimit
		}
	}

	return minResult
}

/*
之前的公式是：
系数 = （ActualCost/1000）/（ExpectCost/285）
系数 < 3的时候limit为500
系数在3～5之间，limit为316
系数 > 5，limit为158
*/

func ReadCostFile(costFilePath string, toleranceScope float64) ([]Result, error) {
	var (
		res []Result
		err error
		bts []byte
	)

	bts, err = ioutil.ReadFile(costFilePath)
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

func GetBestLimit(costFile string, toleranceScope float64) (ResultList, error) {
	resultList, err := ReadCostFile(costFile, toleranceScope)
	if err != nil {
		return resultList, nil
	}

	resultMap := map[ResultKey]ResultValueList{}
	for _, result := range resultList {
		resultKey := ResultKey{
			DiskType:  result.DiskType,
			OpType:    result.OpType,
			BlockSize: result.BlockSize,
			IoDepth:   result.IoDepth,
		}
		resultValue := ResultValue{
			RecoveryLimit: result.RecoveryLimit,
			ExpectCost:    result.ExpectCost,
			ActualCost:    result.ActualCost,
		}

		_, ok := resultMap[resultKey]
		if ok {
			resultMap[resultKey] = append(resultMap[resultKey], resultValue)
		} else {
			resultMap[resultKey] = []ResultValue{}
			resultMap[resultKey] = append(resultMap[resultKey], resultValue)
		}
	}

	bestResultMap := map[ResultKey]ResultValue{}
	for key, value := range resultMap {
		bestResultMap[key] = value.Best(toleranceScope)
	}

	resList := ResultList{}
	for key, value := range bestResultMap {
		resList = append(resList, Result{
			DiskType:      key.DiskType,
			OpType:        key.OpType,
			BlockSize:     key.BlockSize,
			IoDepth:       key.IoDepth,
			RecoveryLimit: value.RecoveryLimit,
			ExpectCost:    value.ExpectCost,
			ActualCost:    value.ActualCost,
		})
	}

	sort.Sort(resList)
	return resultList, nil
}

func main() {
	toleranceScope := 0.3
	costFile := "/Users/liucx/Desktop/QOS/cost和limit测试/all.csv"

	resList, err := GetBestLimit(costFile, toleranceScope)
	if err != nil {
		return
	}
	fmt.Println("DiskType, OpType, BlockSize, IoDepth, RecoveryLimit, ExpectCost, ActualCost")
	for _, value := range resList {
		fmt.Println(value.DiskType, ",",
			value.OpType, ",",
			value.BlockSize, ",",
			value.IoDepth, ",",
			value.RecoveryLimit, ",",
			value.ExpectCost, ",",
			value.ActualCost)
	}
	//resultKey := ResultKey{
	//	DiskType:  "hdd",
	//	OpType:    "randwrite",
	//	BlockSize: "4M",
	//	IoDepth:   64,
	//}
	//logrus.Infof("item:%+v",resultMap[resultKey])
}
