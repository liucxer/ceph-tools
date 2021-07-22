package host_client_test

import (
	"ceph-tools/pkg/host_client"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestHostClient_Upload(t *testing.T) {
	client := host_client.NewHostClient("10.0.20.28")
	err := client.Upload("/tmp", "/Users/liucx/Desktop/ceph.log")
	require.NoError(t, err)

	out, err := client.ExecCmd("ls /tmp")
	require.NoError(t, err)
	spew.Dump(string(out))
	fmt.Println(string(out))
}
