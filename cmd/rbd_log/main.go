package main

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"
)

type Log struct {
	LogTime  time.Time
	OpType   string
	Prio     int
	Cost     int
	CostTime int64
}

// 2021-06-24 04:36:23.022 7fe001c0b700  1 osd.0 op_wq(0) _process OpQueueItem(3.108 PGOpItem(op=osd_op(client.6898201.0:15 3.108 3.76756908 (undecoded) ondisk+write+known_if_redirected e4863) v8) prio 128 cost 48 e4863) [OSD::ShardedOpWQ::_process:11170]
func decodeLine(content string, next string) (Log, error) {
	var res Log
	var err error

	items := strings.Split(content, " ")
	res.LogTime, err = time.Parse("2006-01-02 15:04:05.999", items[0]+" "+items[1])
	if err != nil {
		logrus.Errorf("time.Parse err:%v", err)
		return res, err
	}

	tmp := strings.Split(content, "OpQueueItem")

	// optype
	tmp1 := strings.Split(tmp[1], " ")
	tmp2 := strings.Split(tmp1[1], "(")
	if tmp2[0] == "PGOpItem" {
		tpys := strings.Split(tmp2[1], "=")
		res.OpType = tpys[1]
	} else {
		res.OpType = tmp2[0]
	}

	// Prio int64
	tmp1 = strings.Split(content, " prio ")
	tmp2 = strings.Split(tmp1[1], " ")
	res.Prio, err = strconv.Atoi(tmp2[0])
	if err != nil {
		logrus.Errorf("strconv.Atoi err:%v", err)
		return res, err
	}
	// Cos
	//t int64
	tmp1 = strings.Split(content, "cost")
	tmp2 = strings.Split(tmp1[1], " ")
	res.Cost, err = strconv.Atoi(tmp2[1])
	if err != nil {
		logrus.Errorf("strconv.Atoi err:%v", err)
		return res, err
	}

	if next != "" && strings.Contains(next, "OSD::ShardedOpWQ::_process:11390") {
		// CostTime time.Time
		tmp1 = strings.Split(next, "cost:")
		tmp2 = strings.Split(tmp1[1], " ")
		cost, err := strconv.ParseFloat(tmp2[1], 64)
		if err != nil {
			logrus.Errorf("strconv.Atoi err:%v", err)
			return res, err
		}
		res.CostTime = int64(cost * 1000)
	}

	return res, err
}

func decodeLog(content string) ([]Log, error) {
	var res []Log
	var err error
	var log Log

	lines := strings.Split(content, "\n")

	newLines := []string{}
	// 数据清洗，只保留OSD::ShardedOpWQ::_process:11170, OSD::ShardedOpWQ::_process:11390
	for i := 0; i < len(lines); i++ {
		if !strings.Contains(lines[i], "OSD::ShardedOpWQ::_process:11170") &&
			!strings.Contains(lines[i], "OSD::ShardedOpWQ::_process:11390") {
			continue
		}
		newLines = append(newLines, lines[i])
	}

	for i := 0; i < len(newLines); i++ {
		if !strings.Contains(newLines[i], "OSD::ShardedOpWQ::_process:11170") {
			continue
		}

		if i == len(newLines)-1 {
			log, err = decodeLine(newLines[i], "")
		} else {
			log, err = decodeLine(newLines[i], newLines[i+1])
		}
		if err != nil {
			logrus.Errorf("decodeLine err:%v", err)
			return nil, err
		}
		res = append(res, log)
	}

	return res, err
}

type staticLog struct {
	OpType    string `json:"op_type"`
	Count     int64  `json:"count"`
	TotalTime int64  `json:"total_time"`
	AvgTime   int64  `json:"avg_time"`
	TotalCost int64  `json:"total_cost"`
	AvgCost   int64  `json:"avg_cost"`
}

func staticLogs(logs []Log) map[string]staticLog {
	staticMap := map[string]staticLog{}

	for _, v := range logs {
		s, ok := staticMap[v.OpType]
		if ok {
			s.Count++
			s.TotalTime += v.CostTime
			s.TotalCost += int64(v.Cost)
			staticMap[v.OpType] = s
		} else {
			staticMap[v.OpType] = staticLog{
				OpType:    v.OpType,
				Count:     1,
				TotalTime: v.CostTime,
				TotalCost: int64(v.Cost),
			}
		}
	}

	for _, item := range staticMap {
		v := item
		v.AvgCost = v.TotalCost / v.Count
		v.AvgTime = v.TotalTime / v.Count
		staticMap[item.OpType] = v
	}

	return staticMap
}

func getOSDFile(osdDir string) ([]string, error) {
	var (
		res []string
		err error
	)
	files, err := os.ReadDir(osdDir)
	if err != nil {
		logrus.Errorf("os.ReadDir err:%v", err)
		return res, err
	}

	for _, v := range files {
		if v.IsDir() {
			continue
		}
		if !strings.Contains(v.Name(), "ceph-osd") {
			continue
		}
		res = append(res, osdDir+"/"+v.Name())
	}
	return res, nil
}

func handleOneOSDFile(osdFilePath string) ([]Log, error) {
	var (
		err error
		res []Log
		bts []byte
	)
	bts, err = ioutil.ReadFile(osdFilePath)
	if err != nil {
		logrus.Errorf("ioutil.ReadFile err:%v", err)
		return res, err
	}

	res, err = decodeLog(string(bts))
	if err != nil {
		logrus.Errorf("decodeLog err:%v", err)
		return res, err
	}
	return res, nil
}

func main() {
	// 找到所有osd file
	osdDir := "/Users/liucx/Desktop/ceph"
	osdFiles, err := getOSDFile(osdDir)
	if err != nil {
		return
	}

	var logs []Log
	// 循环处理单个osd file
	for _, osdFile := range osdFiles {
		item, err := handleOneOSDFile(osdFile)
		if err != nil {
			return
		}
		logs = append(logs, item...)
	}

	// 统计
	mapRes := staticLogs(logs)

	bts, err := json.Marshal(mapRes)
	if err != nil {
		return
	}

	println(string(bts))

}
