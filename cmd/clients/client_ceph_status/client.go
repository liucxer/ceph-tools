package client_ceph_status

import (
	context "context"

	github_com_go_courier_courier "github.com/go-courier/courier"
)

type ClientCephStatus interface {
	WithContext(context.Context) ClientCephStatus
	Context() context.Context
	CreateNode(req *CreateNode, metas ...github_com_go_courier_courier.Metadata) (*GithubComLiucxerSrvCephStatusPkgModelsNode, github_com_go_courier_courier.Metadata, error)
	ListCephStatus(req *ListCephStatus, metas ...github_com_go_courier_courier.Metadata) (*[]GithubComLiucxerSrvCephStatusPkgModelsCephStatus, github_com_go_courier_courier.Metadata, error)
	ListNode(req *ListNode, metas ...github_com_go_courier_courier.Metadata) (*[]GithubComLiucxerSrvCephStatusPkgModelsNode, github_com_go_courier_courier.Metadata, error)
	Liveness(metas ...github_com_go_courier_courier.Metadata) (*map[string]string, github_com_go_courier_courier.Metadata, error)
}

func NewClientCephStatus(c github_com_go_courier_courier.Client) *ClientCephStatusStruct {
	return &(ClientCephStatusStruct{
		Client: c,
	})
}

type ClientCephStatusStruct struct {
	Client github_com_go_courier_courier.Client
	ctx    context.Context
}

func (c *ClientCephStatusStruct) WithContext(ctx context.Context) ClientCephStatus {
	cc := new(ClientCephStatusStruct)
	cc.Client = c.Client
	cc.ctx = ctx
	return cc
}

func (c *ClientCephStatusStruct) Context() context.Context {
	if c.ctx != nil {
		return c.ctx
	}
	return context.Background()
}

func (c *ClientCephStatusStruct) CreateNode(req *CreateNode, metas ...github_com_go_courier_courier.Metadata) (*GithubComLiucxerSrvCephStatusPkgModelsNode, github_com_go_courier_courier.Metadata, error) {
	return req.InvokeContext(c.Context(), c.Client, metas...)
}

func (c *ClientCephStatusStruct) ListCephStatus(req *ListCephStatus, metas ...github_com_go_courier_courier.Metadata) (*[]GithubComLiucxerSrvCephStatusPkgModelsCephStatus, github_com_go_courier_courier.Metadata, error) {
	return req.InvokeContext(c.Context(), c.Client, metas...)
}

func (c *ClientCephStatusStruct) ListNode(req *ListNode, metas ...github_com_go_courier_courier.Metadata) (*[]GithubComLiucxerSrvCephStatusPkgModelsNode, github_com_go_courier_courier.Metadata, error) {
	return req.InvokeContext(c.Context(), c.Client, metas...)
}

func (c *ClientCephStatusStruct) Liveness(metas ...github_com_go_courier_courier.Metadata) (*map[string]string, github_com_go_courier_courier.Metadata, error) {
	return (&Liveness{}).InvokeContext(c.Context(), c.Client, metas...)
}
