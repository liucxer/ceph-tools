package ceph_test

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/liucxer/ceph-tools/pkg/ceph"
	"github.com/liucxer/ceph-tools/pkg/host_client"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"strconv"
	"testing"
)

func TestGetOSDIp(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	hostClient, err := host_client.NewHostClient("10.0.8.44")
	require.NoError(t, err)
	defer hostClient.Close()

	resp, err := ceph.GetOSDIp(hostClient, 1)
	require.NoError(t, err)
	spew.Dump(resp)
}

func TestGetDiskGroupByPoolName(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	hostClient, err := host_client.NewHostClient("10.0.20.27")
	require.NoError(t, err)
	defer hostClient.Close()

	resp, err := ceph.GetDiskGroupByPoolName(hostClient, "hdd_1_pool")
	require.NoError(t, err)
	spew.Dump(resp)
}

func TestGetOsdIDsByDiskGroupName(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	hostClient, err := host_client.NewHostClient("10.0.20.27")
	require.NoError(t, err)
	defer hostClient.Close()

	resp, err := ceph.GetOsdIDsByDiskGroupName(hostClient, "hdd_1_disk_group")
	require.NoError(t, err)
	spew.Dump(resp)
}

func TestGetJobCostList(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	hostClient, err := host_client.NewHostClient("10.0.20.28")
	require.NoError(t, err)
	defer hostClient.Close()

	resp, err := ceph.GetJobCostList(hostClient, 9)
	require.NoError(t, err)
	spew.Dump(resp)
}

func TestGetSub(t *testing.T) {
	res := ""
	for i := 0; i < 50; i++ {
		res += strconv.Itoa(i) + ","
	}
	fmt.Println(res)
}

// 0,1,2,3,4,5,6,7,8,10,13,16,17,18,20,21,22,23,24,26,27,29,30,31,32,33,35,36,38,39,40,41,42