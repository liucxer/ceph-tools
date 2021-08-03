package main

import (
	"fmt"
	"github.com/liucxer/ceph-tools/pkg/cluster_client"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Job struct {
	JobID              string    `json:"job_id"`
	JobType            string    `json:"job_type"`
	Size               int64     `json:"size"`
	ExpectCost         float64   `json:"expect_cost"`
	OutQueueActualCost float64   `json:"out_queue_actual_cost"`
	ActualCost         float64   `json:"actual_cost"`
	StartTime          time.Time `json:"start_time"`
	EndTime            time.Time `json:"end_time"`
}

type JobList []Job

func (l JobList) Len() int {
	return len(l)
}

func (l JobList) Less(i, j int) bool {
	return l[i].JobID < l[j].JobID
}

func (l JobList) Swap(i, j int) {
	item := l[i]
	l[i] = l[j]
	l[j] = item
}

func (l JobList) ToCsv() string {
	var res = ""
	header := "job_type," +
		"job_id," +
		"expect_cost,out_queue_actual_cost,actual_cost"
	res = res + header + "\n"
	for _, item := range l {
		itemStr := fmt.Sprintf("%s, L%sL,%f,%f,%f",
			item.JobType,
			item.JobID,
			item.ExpectCost,
			item.OutQueueActualCost,
			item.ActualCost,
		)
		res = res + itemStr + "\n"
	}

	return res
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

/*
日志格式约定下面这个标准哈，我用这个格式解析。
*** use dmclock queue. job_id:12,expect_cost:110, ***
*** use dmclock queue. job_id:12,actual_cost:500, ***
*/

type JobStr string

func (s JobStr) Job() (Job, error) {
	var (
		err error
		job Job
	)
	job.JobID, err = s.JobID()
	if err != nil {
		return job, err
	}

	job.JobType, err = s.JobType()
	if err != nil {
		return job, err
	}

	job.Size, err = s.Size()
	if err != nil {
		return job, err
	}

	job.ExpectCost, err = s.ExpectCost()
	if err != nil {
		return job, err
	}

	job.ActualCost, err = s.ActualCost()
	if err != nil {
		return job, err
	}

	job.OutQueueActualCost, err = s.OutQueueActualCost()
	if err != nil {
		return job, err
	}

	job.StartTime, err = s.StartTime()
	if err != nil {
		return job, err
	}

	job.EndTime, err = s.EndTime()
	if err != nil {
		return job, err
	}

	return job, err
}

// 2021-07-14 23:42:42.892 7fc1c33d4700  2 use dmclock queue. item OpQueueItem(2.d PGOpItem(op=osd_op(client.617511.0:3 2.d 2.8e41aa8d (undecoded) ondisk+read+known_if_redirected e8568) v8) prio 63 cost 63 job_id 6821282842222841856 e8568), expect_cost: 1, osd_mclock_cost_per_byte 5.2e-06, osd_mclock_cost_per_io 0.025 [ceph::mClockOpClassQueue::enqueue:116]
func (s JobStr) StartTime() (time.Time, error) {
	var (
		err error
		res time.Time
	)

	if !strings.Contains(string(s), "expect_cost:") {
		return res, nil
	}

	items := strings.Split(string(s), " ")
	res, err = time.Parse("2006-01-02 15:04:05.999", items[0]+" "+items[1])
	if err != nil {
		logrus.Errorf("time.Parse err:%v", err)
		return res, err
	}

	return res, err
}

func (s JobStr) EndTime() (time.Time, error) {
	var (
		err error
		res time.Time
	)

	if !strings.Contains(string(s), "log_op_stats") {
		return res, nil
	}
	items := strings.Split(string(s), " ")
	res, err = time.Parse("2006-01-02 15:04:05.999", items[0]+" "+items[1])
	if err != nil {
		logrus.Errorf("time.Parse err:%v", err)
		return res, err
	}

	return res, err
}

// 2021-07-14 23:42:42.892 7fc1c33d4700  2 use dmclock queue. item OpQueueItem(2.d PGOpItem(op=osd_op(client.617511.0:3 2.d 2.8e41aa8d (undecoded) ondisk+read+known_if_redirected e8568) v8) prio 63 cost 63 job_id 6821282842222841856 e8568), expect_cost: 1, osd_mclock_cost_per_byte 5.2e-06, osd_mclock_cost_per_io 0.025 [ceph::mClockOpClassQueue::enqueue:116]
func (s JobStr) JobID() (string, error) {
	a := strings.Split(string(s), "client.")
	b := strings.Split(a[1], " ")
	return b[0], nil
}

func (s JobStr) JobType() (string, error) {
	res := ""
	if !strings.Contains(string(s), "OpQueueItem") {
		return "", nil
	}

	tmp := strings.Split(string(s), "OpQueueItem")

	// optype
	tmp1 := strings.Split(tmp[1], " ")
	tmp2 := strings.Split(tmp1[1], "(")
	if tmp2[0] == "PGOpItem" {
		tpys := strings.Split(tmp2[1], "=")
		res = tpys[1]
	} else {
		res = tmp2[0]
	}

	return res, nil
}

func (s JobStr) Size() (int64, error) {
	return 0, nil
}

func (s JobStr) OutQueueActualCost() (float64, error) {
	if !strings.Contains(string(s), "actual_cost ") {
		return 0, nil
	}
	a := strings.Split(string(s), "actual_cost ")
	b := strings.Split(a[1], "ms")

	v2, err := strconv.ParseFloat(b[0], 64)
	if err != nil {
		logrus.Errorf("strconv.ParseFloat err:%v", err)
	}
	return v2, err
}

func (s JobStr) ExpectCost() (float64, error) {
	if !strings.Contains(string(s), "expect_cost:") {
		return 0, nil
	}
	a := strings.Split(string(s), "expect_cost: ")
	b := strings.Split(a[1], ",")

	v2, err := strconv.ParseFloat(b[0], 64)
	if err != nil {
		logrus.Errorf("strconv.ParseFloat err:%v", err)
	}
	return v2, err
}

func (s JobStr) ActualCost() (float64, error) {
	if !strings.Contains(string(s), "actual_cost:") {
		return 0, nil
	}
	a := strings.Split(string(s), "actual_cost:")

	b := strings.Split(a[1], ",")

	v2, err := strconv.ParseFloat(b[0], 64)
	if err != nil {
		logrus.Errorf("strconv.ParseFloat err:%v", err)
	}
	return v2, err
}

func decodeLog(jobMap map[string]Job, content string) error {
	lines := strings.Split(content, "\n")

	newLines := []string{}
	// 数据清洗，只保留OSD::ShardedOpWQ::_process:11170, OSD::ShardedOpWQ::_process:11390
	for i := 0; i < len(lines); i++ {
		if !strings.Contains(lines[i], "use dmclock queue") {
			continue
		}
		newLines = append(newLines, lines[i])
	}

	for i := 0; i < len(newLines); i++ {
		jobStr := JobStr(newLines[i])
		if !strings.Contains(string(jobStr), "client.") {
			continue
		}
		if !strings.Contains(string(jobStr), "use dmclock queue") {
			continue
		}
		job, err := jobStr.Job()
		if err != nil {
			return err
		}

		if job.JobID == "" {
			continue
		}
		if j, ok := jobMap[job.JobID]; ok {
			if job.ExpectCost != 0 && j.ExpectCost == 0 {
				j.ExpectCost = job.ExpectCost
			}

			if job.JobType != "" && j.JobType == "" {
				j.JobType = job.JobType
			}

			if job.Size != 0 && j.Size == 0 {
				j.Size = job.Size
			}

			if job.OutQueueActualCost != 0 && j.OutQueueActualCost == 0 {
				j.OutQueueActualCost = job.OutQueueActualCost
			}

			if job.ActualCost != 0 && j.ActualCost == 0 {
				j.ActualCost = job.ActualCost
			}

			if job.StartTime.Hour() != 0 && j.StartTime.Hour() == 0 {
				j.StartTime = job.StartTime
			}

			if job.EndTime.Hour() != 0 && j.EndTime.Hour() == 0 {
				j.EndTime = job.EndTime
			}

			jobMap[job.JobID] = j
		} else {
			jobMap[job.JobID] = job
		}
	}

	return nil

}

func handleOneOSDFile(jobMap map[string]Job, osdFilePath string) error {
	var (
		err error
		bts []byte
	)
	bts, err = ioutil.ReadFile(osdFilePath)
	if err != nil {
		logrus.Errorf("ioutil.ReadFile err:%v", err)
		return err
	}

	err = decodeLog(jobMap, string(bts))
	if err != nil {
		logrus.Errorf("decodeLog err:%v", err)
		return err
	}
	return nil
}

func Detect(c *cluster_client.Cluster) (float64, error) {
	localLogDir := os.TempDir() + "cephlog/"
	err := os.RemoveAll(localLogDir)
	if err != nil {
		logrus.Errorf("os.RemoveAll err. [err:%v, localLogDir:%s]", err, localLogDir)
		return 0, err
	}

	err = os.Mkdir(localLogDir, os.ModePerm)
	if err != nil {
		logrus.Errorf("os.Mkdir err. [err:%v, localLogDir:%s]", err, localLogDir)
		return 0, err
	}

	err = c.CollectCephLog(localLogDir)
	if err != nil {
		return 0, err
	}

	osdFiles, err := getOSDFile(localLogDir)
	if err != nil {
		return 0, err
	}
	jobMap := map[string]Job{}
	// 循环处理单个osd file
	for _, osdFile := range osdFiles {
		if strings.Contains(osdFile, ".swp") {
			continue
		}
		err := handleOneOSDFile(jobMap, osdFile)
		if err != nil {
			return 0, err
		}
	}

	jobList := JobList{}

	for _, v := range jobMap {
		if v.StartTime.Hour() == 0 {
			continue
		}

		if v.EndTime.Hour() == 0 {
			continue
		}

		if v.JobType != "osd_op" {
			continue
		}

		v.ActualCost = float64(v.EndTime.Sub(v.StartTime).Milliseconds())
		jobList = append(jobList, v)
	}

	if len(jobList) == 0 {
		return 1, nil
	}

	modulus := []float64{}
	for _, v := range jobList {
		if v.ActualCost == 0 || v.ExpectCost == 0 {
			continue
		}
		modulu := (v.ActualCost / float64(1000)) / (v.ExpectCost / float64(285))
		modulus = append(modulus, modulu)
	}
	sort.Float64s(modulus)

	modulusLen := len(modulus)
	modulusNew := modulus[modulusLen/10 : modulusLen*9/10]

	modulusAvg := float64(0)
	modulusSum := float64(0)

	for _, v := range modulusNew {
		modulusSum = modulusSum + v
	}
	modulusAvg = modulusSum / float64(len(modulusNew))

	return modulusAvg, nil
}

func SetLimit(c *cluster_client.Cluster, limit string) (float64, error) {
	for i := 0; i < 12; i++ {
		cmdStr := "ceph daemon osd." + strconv.Itoa(i) + " config set osd_op_queue_mclock_recov_lim " + limit
		c.Master.ExecCmd(cmdStr)
	}
	return 0, nil
}

func DetectAndSetLimit(c *cluster_client.Cluster) error {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("捕获异常:", err)
		}
	}()

	detect, err := Detect(c)
	if err != nil {
		return err
	}

	if detect < 3 {
		SetLimit(c, "500")
	} else if detect >= 3 && detect <= 5 {
		SetLimit(c, "316")
	} else {
		SetLimit(c, "158")
	}
	return err
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage:\n     ./detect ipaddr")
		return
	}
	ipAddr := os.Args[1]

	logrus.SetLevel(logrus.DebugLevel)
	c, err := cluster_client.NewCluster([]string{ipAddr})
	if err != nil {
		return
	}
	defer func() { _ = c.Close() }()

	err = c.ClearCephLog()
	if err != nil {
		return
	}

	time.Sleep(10 * time.Second)

	for {
		DetectAndSetLimit(c)
		//go func() {
		//
		//}()
		time.Sleep(10 * time.Second)
		err = c.ClearCephLog()
		if err != nil {
			return
		}
	}
}
