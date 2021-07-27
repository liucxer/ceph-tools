package ceph_client

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"time"
)

type CephClientSt struct {
	Host     string
	Port     string
	Token    string
	UserName string
	Password string
}

func (t *CephClientSt) HttpRequest(httpMethod, path string, req interface{}, resp interface{}) error {
	if t.Token == "" && path != "/api/auth" {
		err := t.Auth()
		if err != nil {
			return err
		}
	}

	endPoint := "https://" + t.Host + ":" + t.Port + path
	logrus.Debugf("httpMethod:%s, endPoint:%s", httpMethod, endPoint)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	httpClient := http.Client{
		Timeout:   30 * time.Second,
		Transport: tr,
	}

	bts, err := json.Marshal(req)
	if err != nil {
		return err
	}

	request, err := http.NewRequest(httpMethod, endPoint, bytes.NewBuffer(bts))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/vnd.ceph.api.v1.0+json")
	request.Header.Set("Cookie", "token="+t.Token)
	res, err := httpClient.Do(request)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	bts, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if res.StatusCode >= 400 {
		return errors.New(string(bts))
	}

	logrus.Debugf("httpResp:%s", string(bts))
	err = json.Unmarshal(bts, resp)
	if err != nil {
		return err
	}

	return nil

}

func (t *CephClientSt) Auth() error {
	req := &AuthReq{UserName: t.UserName, Password: t.Password}
	resp := &AuthResp{}
	err := t.HttpRequest("POST", "/api/auth", req, resp)
	if err != nil {
		return err
	}

	t.Token = resp.Token
	return err
}

func (t *CephClientSt) HealthFull(req *HealthFullReq, resp *HealthFullResp) error {
	return t.HttpRequest("GET", "/api/health/full", req, resp)
}

func (t *CephClientSt) HealthMinimal(req *HealthMinimalReq, resp *HealthMinimalResp) error {
	return t.HttpRequest("GET", "/api/health/minimal", req, resp)
}
