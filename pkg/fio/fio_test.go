package fio_test

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/liucxer/ceph-tools/pkg/fio"
	"github.com/liucxer/ceph-tools/pkg/host_client"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestFioConfig_Exec(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	hostClient, err := host_client.NewHostClient("10.0.20.27")
	require.NoError(t, err)
	defer hostClient.Close()

	conf := fio.Fio{
		OpType:    "read",
		Runtime:   10,
		BlockSize: "4M",
		IoDepth:   1,
		Pool:      "hdd_2_pool",
		RbdName:   "image",
	}

	fioResult, err := conf.Exec(hostClient)
	require.NoError(t, err)
	spew.Dump(fioResult)
}
