package host_client_test

import (
	"github.com/liucxer/ceph-tools/pkg/host_client"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestHostClient_Download(t *testing.T) {
	client := host_client.NewHostClient("10.0.20.28")
	err := client.Download("/Users/liucx/Desktop/ceph.log", "/var/log/ceph/ceph.log")
	require.NoError(t, err)
}
