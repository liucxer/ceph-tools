package host_client_test

import (
	"ceph-tools/pkg/host_client"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewHost(t *testing.T) {
	client := host_client.NewHostClient("10.0.20.28")
	out, err := client.ExecCmd("ls /tmp")
	require.NoError(t, err)
	spew.Dump(string(out))
	fmt.Println(string(out))
}
