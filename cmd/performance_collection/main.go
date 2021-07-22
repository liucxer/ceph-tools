package main

import (
	"ceph-tools/ceph_client"
	"ceph-tools/host"
	"fmt"
	"sort"
	"time"
)

type CephStatus struct {
	Time              time.Time                     `json:"time"`
	HealthMinimalResp ceph_client.HealthMinimalResp `json:"health_minimal_resp"`
}

var mapCephStatus = map[int64]CephStatus{}

func DoCollect(clientHost *host.Host) error {
	client := ceph_client.CephClientSt{
		Host:     "10.0.20.29",
		Port:     "8443",
		UserName: "admin",
		Password: "xv07b7uhkm1",
	}

	HealthMinimalReq := &ceph_client.HealthMinimalReq{}
	HealthMinimalResp := &ceph_client.HealthMinimalResp{}
	err := client.HealthMinimal(HealthMinimalReq, HealthMinimalResp)
	if err != nil {
		return err
	}

	var cephStatus CephStatus

	cephStatus.Time = time.Now()
	cephStatus.HealthMinimalResp = *HealthMinimalResp
	mapCephStatus[cephStatus.Time.Unix()] = cephStatus
	return nil
}

type CephStatusList []CephStatus

func (list CephStatusList) Len() int {
	return len(list)
}

func (list CephStatusList) Less(i, j int) bool {
	return (list)[i].Time.Unix() < (list)[j].Time.Unix()
}

func (list CephStatusList) Swap(i, j int) {
	item := (list)[i]
	(list)[i] = (list)[j]
	(list)[j] = item
}

func (list CephStatusList) ToCsv() string {
	var res = ""
	header := "time," +
		"recovering_bytes_per_sec" +
		"read_bytes_sec,write_bytes_sec,read_op_per_sec,write_op_per_sec"
	res = res + header + "\n"
	for _, item := range list {
		itemStr := fmt.Sprintf("%s, %d,%d,%d,%d,%d",
			item.Time.Format("2006-01-02 15:04:05"),
			item.HealthMinimalResp.ClientPerf.RecoveringBytesPerSec,
			item.HealthMinimalResp.ClientPerf.ReadBytesSec,
			item.HealthMinimalResp.ClientPerf.WriteBytesSec,
			item.HealthMinimalResp.ClientPerf.ReadOpPerSec,
			item.HealthMinimalResp.ClientPerf.WriteOpPerSec,
		)
		res = res + itemStr + "\n"
	}

	return res
}

func main() {
	//logrus.SetLevel(logrus.DebugLevel)
	totalTime := 1 * 60 // 10分钟
	clientHost := host.NewHost("10.0.20.29")

	// 收集10分钟数据
	go func() {
		for i := 0; i < totalTime*2; i++ {
			go DoCollect(clientHost)
			time.Sleep(500 * time.Millisecond)
		}
	}()

	// 数据写入
	// 1.先确保rbd进程全部关闭
	// 2.开启协程 执行rbd写入
	// 3.10分钟之后 关闭rbd进程
	go func() {
		clientHost.ExecCmd("killall rbd")
		go func() {
			clientHost.ExecCmd("rbd -p data_pool bench foo1 --io-type write --io-size 4K --io-threads 64 --io-total 500G --io-pattern seq")
		}()
		time.Sleep(time.Duration(totalTime) * time.Second)
		clientHost.ExecCmd("killall rbd")
	}()

	// recovery
	// 第5分钟开启recovery
	// 第10分钟关闭recovery
	go func() {
		time.Sleep(time.Duration(totalTime) * time.Second / 2)
		clientHost.ExecCmd("ceph osd pool set rb_pool size 3")
		time.Sleep(time.Duration(totalTime) * time.Second / 2)
		clientHost.ExecCmd("ceph osd pool set rb_pool size 2")
	}()

	time.Sleep(time.Duration(totalTime) * time.Second)
	var cephStatusList CephStatusList
	for _, cephStatus := range mapCephStatus {
		cephStatusList = append(cephStatusList, cephStatus)
	}
	sort.Sort(cephStatusList)
	csv := cephStatusList.ToCsv()
	fmt.Println(csv)
}
