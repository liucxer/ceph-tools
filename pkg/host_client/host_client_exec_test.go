package host_client_test

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/liucxer/ceph-tools/pkg/host_client"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewHost(t *testing.T) {
	client, err := host_client.NewHostClient("10.0.20.28")
	require.NoError(t, err)
	defer client.Close()

	out, err := client.ExecCmd("ls /tmp")
	require.NoError(t, err)
	spew.Dump(string(out))
	fmt.Println(string(out))
}
