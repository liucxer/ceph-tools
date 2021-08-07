package ceph

import (
	"encoding/json"
	"errors"
	"github.com/liucxer/ceph-tools/pkg/interfacer"
	"strconv"
	"strings"
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
/*ceph osd crush tree -f json*/
