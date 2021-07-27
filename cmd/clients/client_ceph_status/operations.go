package client_ceph_status

import (
	context "context"

	github_com_go_courier_courier "github.com/go-courier/courier"
	github_com_go_courier_metax "github.com/go-courier/metax"
)

type CreateNode struct {
	CreateNodeBody GithubComLiucxerSrvCephStatusPkgMgrsNodeCreateNodeBody `in:"body"`
}

func (CreateNode) Path() string {
	return "/ceph-status/v0/node"
}

func (CreateNode) Method() string {
	return "POST"
}

// @StatusErr[InternalServerError][500999001][InternalServerError]
func (req *CreateNode) Do(ctx context.Context, c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) github_com_go_courier_courier.Result {

	ctx = github_com_go_courier_metax.ContextWith(ctx, "operationID", "cephStatus.CreateNode")
	return c.Do(ctx, req, metas...)

}

func (req *CreateNode) InvokeContext(ctx context.Context, c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) (*GithubComLiucxerSrvCephStatusPkgModelsNode, github_com_go_courier_courier.Metadata, error) {
	resp := new(GithubComLiucxerSrvCephStatusPkgModelsNode)

	meta, err := req.Do(ctx, c, metas...).Into(resp)

	return resp, meta, err
}

func (req *CreateNode) Invoke(c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) (*GithubComLiucxerSrvCephStatusPkgModelsNode, github_com_go_courier_courier.Metadata, error) {
	return req.InvokeContext(context.Background(), c, metas...)
}

type ListCephStatus struct {
	EndTime   GithubComGoCourierSqlxV2DatatypesTimestamp `in:"query" name:"endTime,omitempty"`
	NodeID    GithubComLiucxerSrvCephStatusPkgToolsSFID  `in:"query" name:"nodeID"`
	Offset    int64                                      `in:"query" default:"0" name:"offset,omitempty" validate:"@int64[0,]"`
	Size      int64                                      `in:"query" default:"10" name:"size,omitempty" validate:"@int64[-1,]"`
	StartTime GithubComGoCourierSqlxV2DatatypesTimestamp `in:"query" name:"startTime,omitempty"`
}

func (ListCephStatus) Path() string {
	return "/ceph-status/v0/ceph-status"
}

func (ListCephStatus) Method() string {
	return "GET"
}

// @StatusErr[InternalServerError][500999001][InternalServerError]
func (req *ListCephStatus) Do(ctx context.Context, c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) github_com_go_courier_courier.Result {

	ctx = github_com_go_courier_metax.ContextWith(ctx, "operationID", "cephStatus.ListCephStatus")
	return c.Do(ctx, req, metas...)

}

func (req *ListCephStatus) InvokeContext(ctx context.Context, c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) (*[]GithubComLiucxerSrvCephStatusPkgModelsCephStatus, github_com_go_courier_courier.Metadata, error) {
	resp := new([]GithubComLiucxerSrvCephStatusPkgModelsCephStatus)

	meta, err := req.Do(ctx, c, metas...).Into(resp)

	return resp, meta, err
}

func (req *ListCephStatus) Invoke(c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) (*[]GithubComLiucxerSrvCephStatusPkgModelsCephStatus, github_com_go_courier_courier.Metadata, error) {
	return req.InvokeContext(context.Background(), c, metas...)
}

type ListNode struct {
	Offset int64 `in:"query" default:"0" name:"offset,omitempty" validate:"@int64[0,]"`
	Size   int64 `in:"query" default:"10" name:"size,omitempty" validate:"@int64[-1,]"`
}

func (ListNode) Path() string {
	return "/ceph-status/v0/node"
}

func (ListNode) Method() string {
	return "GET"
}

func (req *ListNode) Do(ctx context.Context, c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) github_com_go_courier_courier.Result {

	ctx = github_com_go_courier_metax.ContextWith(ctx, "operationID", "cephStatus.ListNode")
	return c.Do(ctx, req, metas...)

}

func (req *ListNode) InvokeContext(ctx context.Context, c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) (*[]GithubComLiucxerSrvCephStatusPkgModelsNode, github_com_go_courier_courier.Metadata, error) {
	resp := new([]GithubComLiucxerSrvCephStatusPkgModelsNode)

	meta, err := req.Do(ctx, c, metas...).Into(resp)

	return resp, meta, err
}

func (req *ListNode) Invoke(c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) (*[]GithubComLiucxerSrvCephStatusPkgModelsNode, github_com_go_courier_courier.Metadata, error) {
	return req.InvokeContext(context.Background(), c, metas...)
}

type Liveness struct {
}

func (Liveness) Path() string {
	return "/ceph-status/liveness"
}

func (Liveness) Method() string {
	return "GET"
}

func (req *Liveness) Do(ctx context.Context, c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) github_com_go_courier_courier.Result {

	ctx = github_com_go_courier_metax.ContextWith(ctx, "operationID", "cephStatus.Liveness")
	return c.Do(ctx, req, metas...)

}

func (req *Liveness) InvokeContext(ctx context.Context, c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) (*map[string]string, github_com_go_courier_courier.Metadata, error) {
	resp := new(map[string]string)

	meta, err := req.Do(ctx, c, metas...).Into(resp)

	return resp, meta, err
}

func (req *Liveness) Invoke(c github_com_go_courier_courier.Client, metas ...github_com_go_courier_courier.Metadata) (*map[string]string, github_com_go_courier_courier.Metadata, error) {
	return req.InvokeContext(context.Background(), c, metas...)
}
