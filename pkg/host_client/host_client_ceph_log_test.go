package host_client_test

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/liucxer/ceph-tools/pkg/host_client"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestHostClient_BackupAndClearCephLog(t *testing.T) {
	hostClient, err := host_client.NewHostClient("10.0.20.28")
	require.NoError(t, err)
	defer hostClient.Close()

	err = hostClient.BackupAndClearCephLog()
	require.NoError(t, err)
}

func TestHostClient_ReadBackupCephLog(t *testing.T) {
	hostClient, err := host_client.NewHostClient("10.0.20.28")
	require.NoError(t, err)
	defer hostClient.Close()

	res, err := hostClient.ReadBackupCephLog()
	require.NoError(t, err)
	spew.Dump(len(res))
}

func TestHostClient_OsdLogFiles(t *testing.T) {
	hostClient, err := host_client.NewHostClient("10.0.20.28")
	require.NoError(t, err)
	defer hostClient.Close()

	logFiles, err := hostClient.OsdLogFiles()
	require.NoError(t, err)
	spew.Dump(logFiles)

	backupLogFiles, err := hostClient.OsdBackupLogFiles()
	require.NoError(t, err)
	spew.Dump(backupLogFiles)
}
