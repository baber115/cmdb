package host

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"codeup.aliyun.com/baber/go/cmdb/apps/resource"
	"codeup.aliyun.com/baber/go/cmdb/utils"
	"github.com/go-playground/validator/v10"
	"github.com/infraboard/mcube/flowcontrol/tokenbucket"
	"github.com/infraboard/mcube/http/request"
	pb_request "github.com/infraboard/mcube/pb/request"
	"github.com/infraboard/mcube/types/ftime"
)

const (
	AppName = "host"
)

var (
	validate = validator.New()
)

func NewQueryHostRequestFromHTTP(r *http.Request) *QueryHostRequest {
	qs := r.URL.Query()
	page := request.NewPageRequestFromHTTP(r)
	kw := qs.Get("keywords")

	return &QueryHostRequest{
		Page:     page,
		Keywords: kw,
	}
}

func NewHostSet() *HostSet {
	return &HostSet{
		Items: []*Host{},
	}
}

func NewDefaultHost() *Host {
	return &Host{
		Base: &resource.Base{
			ResourceType: resource.Type_HOST,
		},
		Information: &resource.Information{},
		Describe:    &Describe{},
	}
}

func NewDescribeHostRequestWithID(id string) *DescribeHostRequest {
	return &DescribeHostRequest{
		DescribeBy: DescribeBy_HOST_ID,
		Value:      id,
	}
}

func (h *Host) GenHash() error {
	h.Base.ResourceHash = h.Information.Hash()
	h.Base.DescribeHash = utils.Hash(h.Describe)
	return nil
}

func (req *QueryHostRequest) OffSet() int64 {
	return int64(req.Page.PageSize) * int64(req.Page.PageNumber-1)
}

func (h *Host) Put(req *UpdateHostData) {
	oldRH, oldDH := h.Base.ResourceHash, h.Base.DescribeHash

	h.Information = req.Information
	h.Describe = req.Describe
	h.Information.UpdateAt = ftime.Now().Timestamp()
	h.GenHash()

	if h.Base.ResourceHash != oldRH {
		h.Base.ResourceHashChanged = true
	}
	if h.Base.DescribeHash != oldDH {
		h.Base.DescribeHashChanged = true
	}
}

func (h *Host) Patch(req *UpdateHostData) error {
	oldRH, oldDH := h.Base.ResourceHash, h.Base.DescribeHash

	err := ObjectPatch(h.Information, req.Information)
	if err != nil {
		return err
	}

	err = ObjectPatch(h.Describe, req.Describe)
	if err != nil {
		return err
	}

	h.Information.UpdateAt = ftime.Now().Timestamp()
	h.GenHash()

	if h.Base.ResourceHash != oldRH {
		h.Base.ResourceHashChanged = true
	}
	if h.Base.DescribeHash != oldDH {
		h.Base.DescribeHashChanged = true
	}

	return nil
}

func ObjectPatch(old, new interface{}) error {
	newByte, err := json.Marshal(new)
	if err != nil {
		return err
	}
	return json.Unmarshal(newByte, old)
}

func (s *HostSet) ResourceIds() (ids []string) {
	for i := range s.Items {
		ids = append(ids, s.Items[i].Base.Id)
	}
	return
}

func (s *HostSet) UpdateTag(tags []*resource.Tag) {
	for i := range tags {
		for j := range s.Items {
			if s.Items[j].Base.Id == tags[i].ResourceId {
				s.Items[j].Information.AddTag(tags[i])
			}
		}
	}
}

func (s *HostSet) Add(item any) {
	s.Items = append(s.Items, item.(*Host))
}

func (s *HostSet) Length() int64 {
	return int64(len(s.Items))
}

func NewUpdateHostDataByIns(ins *Host) *UpdateHostData {
	return &UpdateHostData{
		Information: ins.Information,
		Describe:    ins.Describe,
	}
}

func (d *Describe) KeyPairNameToString() string {
	return strings.Join(d.KeyPairName, ",")
}

func (d *Describe) SecurityGroupsToString() string {
	return strings.Join(d.SecurityGroups, ",")
}

func (d *Describe) LoadKeyPairNameString(s string) {
	if s != "" {
		d.KeyPairName = strings.Split(s, ",")
	}
}

func (d *Describe) LoadSecurityGroupsString(s string) {
	if s != "" {
		d.SecurityGroups = strings.Split(s, ",")
	}
}

func NewDeleteHostRequestWithID(id string) *ReleaseHostRequest {
	return &ReleaseHostRequest{Id: id}
}

func NewUpdateHostRequest(id string) *UpdateHostRequest {
	return &UpdateHostRequest{
		Id:             id,
		UpdateMode:     pb_request.UpdateMode_PUT,
		UpdateHostData: &UpdateHostData{},
	}
}

func (req *UpdateHostRequest) Validate() error {
	return validate.Struct(req)
}

// 分页器
type Pagger interface {
	HasNext() bool
	Scan(context.Context, *HostSet) error
	SetPageSize(int64)
}

func (req *DescribeHostRequest) Where() (string, interface{}) {
	switch req.DescribeBy {
	default:
		return "r.id = ?", req.Value
	}
}

// 面向组合, 用他来实现一个模板, 除了Scan的其他方法都实现

type BasePagerV2 struct {
	// 令牌桶
	hasNext bool
	tb      *tokenbucket.Bucket

	// 控制分页的核心参数
	pageNumber int64
	pageSize   int64
}

type Set interface {
	// 往Set里面添加元素, 任何类型都可以
	Add(any)
	// 当前的集合里面有多个元素
	Length() int64
}

func NewBasePagerV2() *BasePagerV2 {
	return &BasePagerV2{
		hasNext:    true,
		tb:         tokenbucket.NewBucketWithRate(1, 1),
		pageNumber: 1,
		pageSize:   20,
	}
}

func (p *BasePagerV2) SetPageSize(ps int64) {
	p.pageSize = ps
}

func (p *BasePagerV2) SetRate(rate float64) {
	p.tb.SetRate(rate)
}

func (p *BasePagerV2) HasNext() bool {
	// 等待分配令牌
	p.tb.Wait(1)

	return p.hasNext
}

func (p *BasePagerV2) Offset() int64 {
	return (p.pageNumber - 1) * p.pageSize
}

func (p *BasePagerV2) PageSize() int64 {
	return p.pageSize
}

func (p *BasePagerV2) PageNumber() int64 {
	return p.pageNumber
}

func (p *BasePagerV2) CheckHasNext(current int64) {
	// 可以根据当前一页是满页来决定是否有下一页
	if current < p.pageSize {
		p.hasNext = false
	} else {
		// 直接调整指针到下一页
		p.pageNumber++
	}
}
