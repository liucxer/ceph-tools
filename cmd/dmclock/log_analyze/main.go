package log_analyze

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type DMClockJob struct {
	JobID              string    `json:"job_id"`
	JobType            string    `json:"job_type"`
	Size               int64     `json:"size"`
	ExpectCost         float64   `json:"expect_cost"`
	OutQueueActualCost float64   `json:"out_queue_actual_cost"`
	ActualCost         float64   `json:"actual_cost"`
	StartTime          time.Time `json:"start_time"`
	EndTime            time.Time `json:"end_time"`
}

type DMClockJobList []DMClockJob

func (l DMClockJobList) ExpectCost() float64 {
	var expectCostTotal float64
	for _, v := range l {
		expectCostTotal = expectCostTotal + v.ExpectCost
	}
	return expectCostTotal / float64(len(l))
}
func (l DMClockJobList) ActualCost() float64 {
	var actualCostTotal float64
	for _, v := range l {
		actualCostTotal = actualCostTotal + v.ActualCost
	}
	return actualCostTotal / float64(len(l))
}

func (l DMClockJobList) Len() int {
	return len(l)
}

func (l DMClockJobList) Less(i, j int) bool {
	return l[i].JobID < l[j].JobID
}

func (l DMClockJobList) Swap(i, j int) {
	item := l[i]
	l[i] = l[j]
	l[j] = item
}

func (l DMClockJobList) ToCsv() string {
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

func GetOSDFile(osdDir string) ([]string, error) {
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

func (s JobStr) DMClockJob() (DMClockJob, error) {
	var (
		err error
		job DMClockJob
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

	//job.ActualCost, err = s.ActualCost()
	//if err != nil {
	//	return job, err
	//}

	job.OutQueueActualCost, err = s.OutQueueActualCost()
	if err != nil {
		return job, err
	}

	job.StartTime, _ = s.StartTime()

	job.EndTime, _ = s.EndTime()

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
	//if !strings.Contains(string(s), "actual_cost:") {
	//	return 0, nil
	//}
	//a := strings.Split(string(s), "actual_cost:")
	//
	//b := strings.Split(a[1], ",")
	//
	//v2, err := strconv.ParseFloat(b[0], 64)
	//if err != nil {
	//	logrus.Errorf("strconv.ParseFloat err:%v", err)
	//}
	return 0, nil
}

func decodeLog(jobMap map[string]DMClockJob, content string) error {
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
		job, err := jobStr.DMClockJob()
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

func HandleOneOSDFile(jobMap map[string]DMClockJob, osdFilePath string) error {
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

func LogAnalyze(osdLogDir string) (*DMClockJobList, error) {
	osdFiles, err := GetOSDFile(osdLogDir)
	if err != nil {
		return nil, err
	}
	jobMap := map[string]DMClockJob{}
	// 循环处理单个osd file
	for _, osdFile := range osdFiles {
		if strings.Contains(osdFile, ".swp") {
			continue
		}
		err := HandleOneOSDFile(jobMap, osdFile)
		if err != nil {
			return nil, err
		}
	}

	var jobList DMClockJobList

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
		if v.ExpectCost == 0 || v.ActualCost == 0 {
			continue
		}
		jobList = append(jobList, v)
	}

	sort.Sort(jobList)
	return &jobList, err
}

//func main() {
//
//	// 找到所有osd file
//	osdDir := "/Users/liucx/Desktop/ceph"
//
//	csv := jobList.ToCsv()
//
//	tmpFile := "/Users/liucx/Desktop/1.csv"
//	err := ioutil.WriteFile(tmpFile, []byte(csv), os.ModePerm)
//	if err != nil {
//		logrus.Errorf("ioutil.WriteFile err:%v", err)
//		return
//	}
//	fmt.Println(tmpFile)
//}
