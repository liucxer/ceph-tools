package ceph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/liucxer/ceph-tools/pkg/interfacer"
	"strconv"
	"strings"
	"time"
)

func GetOSDIp(worker interfacer.Worker, osdID int64) (string, error) {
	var (
		err error
	)

	type Resp struct {
		BackAddr string `json:"back_addr"`
	}

	bts, err := worker.ExecCmd("ceph osd metadata " + strconv.Itoa(int(osdID)))
	if err != nil {
		return "", err
	}

	var res Resp
	err = json.Unmarshal(bts, &res)
	if err != nil {
		return "", err
	}

	tmpList := strings.Split(res.BackAddr, ":")
	if len(tmpList) < 2 {
		return "", errors.New("res.BackAddr error")
	}

	return tmpList[1], nil
}

func GetDiskGroupByPoolName(worker interfacer.Worker, poolName string) (string, error) {
	var (
		err error
	)

	type Resp struct {
		Steps []struct {
			Op       string `json:"op"`
			ItemName string `json:"item_name"`
		} `json:"steps"`
	}

	bts, err := worker.ExecCmd("ceph osd crush rule dump  " + poolName + "_ruleset")
	if err != nil {
		return "", err
	}

	var res Resp
	err = json.Unmarshal(bts, &res)
	if err != nil {
		return "", err
	}

	for _, step := range res.Steps {
		if step.Op == "take" {
			return step.ItemName, nil
		}
	}
	return "", errors.New("not found DiskGroup")
}

type Node struct {
	ID       int64   `json:"id"`
	Name     string  `json:"name"`
	Type     string  `json:"type"`
	Children []int64 `json:"children"`
}

func GetOsdIDsByDiskGroupName(worker interfacer.Worker, diskGroupName string) ([]int64, error) {
	var (
		err error
	)

	bts, err := worker.ExecCmd("ceph osd crush tree -f json")
	if err != nil {
		return nil, err
	}

	type Resp struct {
		Nodes []Node `json:"nodes"`
	}

	var res Resp
	err = json.Unmarshal(bts, &res)
	if err != nil {
		return nil, err
	}

	nodeIDMap := map[int64]Node{}
	nodeNameMap := map[string]Node{}
	for _, node := range res.Nodes {
		nodeIDMap[node.ID] = node
		nodeNameMap[node.Name] = node
	}

	var diskGroupID int64
	if value, ok := nodeNameMap[diskGroupName]; ok {
		diskGroupID = value.ID
	}

	nodeList := GetSub(diskGroupID, nodeIDMap)

	nodeIDs := []int64{}
	for _, node := range nodeList {
		nodeIDs = append(nodeIDs, node.ID)
	}
	return nodeIDs, nil
}

func GetSub(nodeID int64, nodeIDMap map[int64]Node) []Node {
	node := nodeIDMap[nodeID]

	if len(node.Children) == 0 {
		return []Node{node}
	}

	var res []Node
	for _, itemID := range node.Children {
		tmpList := GetSub(itemID, nodeIDMap)
		res = append(res, tmpList...)
	}
	return res
}

// 根据磁盘组找到 osd
type JobCost struct {
	ExpectCost float64 `json:"expect_cost"`
	ActualCost float64 `json:"actual_cost(ms)"`
	Type       string  `json:"type"`
	Bytes      int64   `json:"bytes"`
}

type JobCostList []JobCost

func (list JobCostList) BaseLineActualCost() float64 {
	if len(list) == 0 {
		return 0
	}

	actualCost := []float64{}
	for _, item := range list {
		if item.ActualCost < 1 {
			continue
		}
		actualCost = append(actualCost, item.ActualCost)
	}
	if len(actualCost) > 10 {
		return actualCost[len(actualCost)/10]
	} else {
		return actualCost[0]
	}

}

func (list JobCostList) AvgExpectCost() float64 {
	if len(list) == 0 {
		return 0
	}
	sum := float64(0)
	for _, item := range list {
		sum = sum + item.ExpectCost
	}

	res, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", sum/float64(len(list))), 64)
	return res
}

func (list JobCostList) AvgActualCost() float64 {
	if len(list) == 0 {
		return 0
	}

	sum := float64(0)
	for _, item := range list {
		sum = sum + item.ActualCost
	}
	res, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", sum/float64(len(list))), 64)
	return res
}

func (list JobCostList) TotalExpectCost() float64 {
	sum := float64(0)
	for _, item := range list {
		sum = sum + item.ExpectCost
	}
	return sum
}

func (list JobCostList) TotalActualCost() float64 {
	sum := float64(0)
	for _, item := range list {
		sum = sum + item.ActualCost
	}
	return sum
}

func (list JobCostList) Coefficient() float64 {
	return list.AvgActualCost() / list.AvgExpectCost()
}

/* ceph tell osd.0 dump_recent_ops_cost */
func GetJobCostList(worker interfacer.Worker, osdNum int64) (JobCostList, error) {
	var (
		err error
	)

	bts, err := worker.ExecCmd("ceph tell osd." + strconv.Itoa(int(osdNum)) + " dump_recent_ops_cost")
	if err != nil {
		return nil, err
	}

	var resp JobCostList
	err = json.Unmarshal(bts, &resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func AsyncJobCostListFunc(ctx context.Context, worker interfacer.Worker, osdNum []int64) (JobCostList, error) {
	var (
		err error
		res JobCostList
	)
	for _, osdNum := range osdNum {
		itemOsdNum := osdNum
		go func(ctx context.Context) {
			for {
				select {
				case <-ctx.Done():
					return
				default:
					item, _ := GetJobCostList(worker, itemOsdNum)
					res = append(res, item...)
					time.Sleep(10 * time.Second)
				}
			}
		}(ctx)
	}

	return res, err
}

func Get4KRandWriteIops(worker interfacer.Worker, osdNum int64) (float64, error) {
	var (
		err error
	)

	type Resp struct {
		BytesWritten int64   `json:"bytes_written"`
		BlockSize    int64   `json:"blocksize"`
		ElapsedSec   float64 `json:"elapsed_sec"`
		BytesPerSec  float64 `json:"bytes_per_sec"`
		Iops         float64 `json:"iops"`
		Latency      float64 `json:"latency(ms)"`
	}

	cmdStr := "ceph tell osd." + strconv.Itoa(int(osdNum)) + " cache drop"
	bts, err := worker.ExecCmd(cmdStr)
	if err != nil {
		return 0, err
	}

	cmdStr = "ceph tell osd." + strconv.Itoa(int(osdNum)) + " bench 12288000 4096 4194304 100"
	bts, err = worker.ExecCmd(cmdStr)
	if err != nil {
		return 0, err
	}

	var resp Resp
	err = json.Unmarshal(bts, &resp)
	if err != nil {
		return 0, err
	}

	return resp.Iops, nil
}
