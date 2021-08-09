package ceph_test

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/liucxer/ceph-tools/pkg/ceph"
	"github.com/liucxer/ceph-tools/pkg/host_client"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
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
