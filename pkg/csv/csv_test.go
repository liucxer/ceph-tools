package csv_test

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/liucxer/ceph-tools/pkg/csv"
	"github.com/liucxer/ceph-tools/pkg/fio"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestObjectToCsv(t *testing.T) {
	type FioConfig struct {
		DiskType   string `json:"diskType"`
		Runtime    int64  `json:"runtime"`
		OpType     string `json:"opType"`
		DataPool   string `json:"dataPool"`
		DataVolume string `json:"dataVolume"`
		BlockSize  string `json:"blockSize"`
		IoDepth    int64  `json:"ioDepth"`
	}

	type FioResult struct {
		FioConfig
		fio.FioResult
		ExpectCost float64 `json:"expectCost"`
		ActualCost float64 `json:"actualCost"`
	}

	var res FioResult
	res.FioResult.WriteIops = 8
	res.FioResult.ReadIops = 7
	res.ActualCost = 9
	res.ActualCost = 10

	nameStr, valueStr, err := csv.ObjectToCsv(res)
	require.NoError(t, err)
	spew.Dump(valueStr)
	fmt.Println(valueStr)
	spew.Dump(nameStr)
	fmt.Println(nameStr)

	nameStr, valueStr, err = csv.ObjectToCsv(&res)
	require.NoError(t, err)
	spew.Dump(valueStr)
	fmt.Println(valueStr)
	spew.Dump(nameStr)
	fmt.Println(nameStr)
}

func TestObjectListToCsv(t *testing.T) {
	type FioConfig struct {
		DiskType   string `json:"diskType"`
		Runtime    int64  `json:"runtime"`
		OpType     string `json:"opType"`
		DataPool   string `json:"dataPool"`
		DataVolume string `json:"dataVolume"`
		BlockSize  string `json:"blockSize"`
		IoDepth    int64  `json:"ioDepth"`
	}

	type FioResult struct {
		FioConfig
		fio.FioResult
		ExpectCost float64 `json:"expectCost"`
		ActualCost float64 `json:"actualCost"`
	}

	var res FioResult
	res.FioResult.WriteIops = 8
	res.FioResult.ReadIops = 7
	res.ActualCost = 9
	res.ActualCost = 10

	var res1 FioResult
	res1.FioResult.WriteIops = 1
	res1.FioResult.ReadIops = 2
	res1.ActualCost = 3
	res1.ActualCost = 4

	valueStr, err := csv.ObjectListToCsv([]FioResult{res, res1})
	require.NoError(t, err)
	fmt.Println(valueStr)
}
