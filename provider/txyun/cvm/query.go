package cvm

import (
	"context"

	"codeup.aliyun.com/baber/go/cmdb/apps/host"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
)

// 查看实例列表: https://cloud.tencent.com/document/api/213/15728
func (c *CVMOperator) Query(ctx context.Context, req *cvm.DescribeInstancesRequest) (*host.HostSet, error) {
	response, err := c.client.DescribeInstancesWithContext(ctx, req)
	if err != nil {
		return nil, err
	}
	c.log.Debugf(response.ToJsonString())

	set := c.transferSet(response.Response.InstanceSet)
	return set, nil
}
