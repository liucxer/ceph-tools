package ceph_client_test

import (
	"ceph-tools/ceph_client"
	"github.com/davecgh/go-spew/spew"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCephClientSt_Auth(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	client := ceph_client.CephClientSt{
		Host:     "10.0.20.29",
		Port:     "8443",
		UserName: "admin",
		Password: "xv07b7uhkm1",
	}

	HealthMinimalReq := &ceph_client.HealthMinimalReq{}
	HealthMinimalResp := &ceph_client.HealthMinimalResp{}
	err := client.HealthMinimal(HealthMinimalReq, HealthMinimalResp)
	require.NoError(t, err)
	spew.Dump(HealthMinimalResp)
}
