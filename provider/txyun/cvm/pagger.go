package cvm

import (
	"context"

	"codeup.aliyun.com/baber/go/cmdb/apps/host"
	"github.com/infraboard/mcube/logger"
	"github.com/infraboard/mcube/logger/zap"
	tx_cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
)

var (
	_ host.Pagger = &pagger{}
)

func NewPagger(op *CVMOperator) *pagger {
	req := tx_cvm.NewDescribeInstancesRequest()
	p := &pagger{
		op:         op,
		req:        req,
		hasNext:    true,
		pageNumber: 1,
		pageSize:   20,
		log:        zap.L().Named("CVM pagger"),
	}
	p.req.Offset = p.offset()
	p.req.Limit = &p.pageSize
	return p
}

type pagger struct {
	req        *tx_cvm.DescribeInstancesRequest
	op         *CVMOperator
	hasNext    bool
	pageNumber int64
	pageSize   int64
	log        logger.Logger
}

func (p *pagger) SetPageSize(pageSize int64) {
	p.pageSize = pageSize
}

// 判断是否有下一页
func (p *pagger) HasNext() bool {
	return p.hasNext
}

// 计算翻页跳过几条数据
func (p *pagger) offset() *int64 {
	offset := (p.pageNumber - 1) * p.pageSize

	return &offset
}

// 修改req，执行真正的下一页的offset
func (p *pagger) nextReq() *tx_cvm.DescribeInstancesRequest {
	p.req.Offset = p.offset()
	p.req.Limit = &p.pageSize

	return p.req
}

func (p *pagger) Scan(ctx context.Context, h *host.HostSet) error {
	p.log.Debugf("p.pageNummber = %d", p.pageNumber)
	response, err := p.op.Query(ctx, p.nextReq())
	if err != nil {
		return err
	}
	if response.Length() < p.pageSize {
		p.hasNext = false
	}
	// 把查询出来的数据clone给hostSet
	for i := range response.Items {
		h.Add(response.Items[i])
	}

	// 修改指针到下一页
	p.pageNumber++

	return nil
}
