package ceph_test

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/liucxer/ceph-tools/pkg/ceph"
	"github.com/liucxer/ceph-tools/pkg/host_client"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGetAllIps(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	hostClient, err := host_client.NewHostClient("10.0.8.44")
	require.NoError(t, err)
	defer hostClient.Close()

	resp, err := ceph.GetAllIps(hostClient)
	require.NoError(t, err)
	spew.Dump(resp)
}
