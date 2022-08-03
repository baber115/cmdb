package cvm

import (
	"context"

	"codeup.aliyun.com/baber/go/cmdb/apps/host"
	"github.com/infraboard/mcube/logger"
	"github.com/infraboard/mcube/logger/zap"
	tx_cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
)

func NewPagerV2(op *CVMOperator) *pagerV2 {
	return &pagerV2{
		req:         tx_cvm.NewDescribeInstancesRequest(),
		op:          op,
		log:         zap.L().Named("CVM pagger"),
		BasePagerV2: host.NewBasePagerV2(),
	}
}

type pagerV2 struct {
	req *tx_cvm.DescribeInstancesRequest
	op  *CVMOperator
	log logger.Logger

	*host.BasePagerV2
}

// 修改req，执行真正的下一页的offset
func (p *pagerV2) nextReq() *tx_cvm.DescribeInstancesRequest {
	offset := p.Offset()
	pageSize := p.PageSize()
	p.req.Offset = &offset
	p.req.Limit = &pageSize

	return p.req
}

func (p *pagerV2) Scan(ctx context.Context, h host.Set) error {
	p.log.Debugf("p.pageNummber = %d", p.PageNumber())
	response, err := p.op.Query(ctx, p.nextReq())
	if err != nil {
		return err
	}

	// 把查询出来的数据clone给hostSet
	for i := range response.Items {
		h.Add(response.Items[i])
	}

	// 可以根据当前一页是满页来决定是否有下一页
	p.CheckHasNext(h.Length())

	return nil
}
