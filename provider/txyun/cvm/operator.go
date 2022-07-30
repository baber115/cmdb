package cvm

import (
	"time"

	"codeup.aliyun.com/baber/go/cmdb/apps/host"
	"codeup.aliyun.com/baber/go/cmdb/apps/resource"
	"github.com/infraboard/cmdb/utils"
	"github.com/infraboard/mcube/logger"
	"github.com/infraboard/mcube/logger/zap"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
)

func NewCVMOperator(conn *cvm.Client) *CVMOperator {
	return &CVMOperator{
		client: conn,
		log:    zap.L().Named("Tx CVM"),
	}
}

type CVMOperator struct {
	client *cvm.Client
	log    logger.Logger
}

func (c *CVMOperator) transferSet(items []*cvm.Instance) *host.HostSet {
	set := host.NewHostSet()
	for i := range items {
		set.Add(c.transferOne(items[i]))
	}

	return set
}

func (c *CVMOperator) transferOne(item *cvm.Instance) *host.Host {
	h := host.NewDefaultHost()
	h.Base.Vendor = resource.Vendor_TENCENT
	h.Base.Region = c.client.GetRegion()
	h.Base.Zone = utils.PtrStrV(item.Placement.Zone)
	h.Base.CreateAt = c.parseTime(utils.PtrStrV(item.CreatedTime))
	h.Base.Id = utils.PtrStrV(item.InstanceId)

	h.Information.ExpireAt = c.parseTime(utils.PtrStrV(item.ExpiredTime))
	h.Information.Type = utils.PtrStrV(item.InstanceType)
	h.Information.Name = utils.PtrStrV(item.InstanceName)
	h.Information.Status = utils.PtrStrV(item.InstanceState)
	h.Information.Tags = transferTags(item.Tags)
	h.Information.PublicIp = utils.SlicePtrStrv(item.PublicIpAddresses)
	h.Information.PrivateIp = utils.SlicePtrStrv(item.PrivateIpAddresses)
	h.Information.PayType = utils.PtrStrV(item.InstanceChargeType)
	//h.Information.SyncAccount = c.GetAccountId()

	h.Describe.Cpu = utils.PtrInt64(item.CPU)
	h.Describe.Memory = utils.PtrInt64(item.Memory)
	h.Describe.OsName = utils.PtrStrV(item.OsName)
	h.Describe.SerialNumber = utils.PtrStrV(item.Uuid)
	h.Describe.ImageId = utils.PtrStrV(item.ImageId)
	if item.InternetAccessible != nil {
		h.Describe.InternetMaxBandwidthOut = utils.PtrInt64(item.InternetAccessible.InternetMaxBandwidthOut)
	}
	h.Describe.KeyPairName = utils.SlicePtrStrv(item.LoginSettings.KeyIds)
	h.Describe.SecurityGroups = utils.SlicePtrStrv(item.SecurityGroupIds)
	return h
}

func transferTags(tags []*cvm.Tag) (ret []*resource.Tag) {
	for i := range tags {
		ret = append(ret, resource.NewThirdTag(
			utils.PtrStrV(tags[i].Key),
			utils.PtrStrV(tags[i].Value)),
		)
	}
	return
}

func (o *CVMOperator) parseTime(t string) int64 {
	if t == "" {
		return 0
	}

	ts, err := time.Parse("2006-01-02T15:04:05Z", t)
	if err != nil {
		o.log.Errorf("parse time %s error, %s", t, err)
		return 0
	}

	return ts.UnixNano() / 1000000
}
