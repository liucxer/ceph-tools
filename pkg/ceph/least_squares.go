package ceph

import (
	"context"
	"github.com/liucxer/ceph-tools/pkg/interfacer"
	"time"
)

func (conf *CephConf) InitLineMetaData(worker interfacer.Worker) error {
	var (
		err error
	)

	var jobCostList JobCostList
	ctx, cancelFn := context.WithCancel(context.Background())

	// 收集CostList数据
	for _, osdNum := range conf.OsdNum {
		itemOsdNum := osdNum
		go func(ctx context.Context) {
			for {
				select {
				case <-ctx.Done():
					return
				default:
					itemJobCostList, _ := GetJobCostList(conf.Worker, itemOsdNum)
					time.Sleep(10 * time.Second)
					jobCostList = append(jobCostList, itemJobCostList...)
				}
			}
		}(ctx)
	}
	// 执行fio动作

	cancelFn()

	// 线性回归
	return err
}
