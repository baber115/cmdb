package impl

import (
	"context"
	"fmt"
	"strings"

	"codeup.aliyun.com/baber/go/cmdb/apps/resource"
	"github.com/infraboard/mcube/exception"
	"github.com/infraboard/mcube/sqlbuilder"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *service) Search(ctx context.Context, req *resource.SearchRequest) (*resource.ResourceSet, error) {
	// sql里的里具体是left join 还是right join，取决于是否需要关联tag表
	// 当筛选tag时，需要关联右表，如果右表数据少的话，效率更高
	join := "LEFT"
	if req.HasTag() {
		join = "RIGHT"
	}
	query := sqlbuilder.NewQuery(fmt.Sprintf(sqlQueryResource, join))
	s.BuildQuery(req, query)

	/*
		==========
		count语句
		==========
	*/
	set := resource.NewResourceSet()
	countSQL, args := query.BuildFromNewBase(fmt.Sprintf(sqlCountResource, join))
	countStmt, err := s.db.Prepare(countSQL)
	if err != nil {
		s.log.Debugf("count sql, %s, %v", countSQL, args)
		return nil, status.Errorf(codes.Unimplemented, "method Search not implemented")
	}
	defer countStmt.Close()
	err = countStmt.QueryRow(args).Scan(&set.Total)
	if err != nil {
		s.log.Debugf("count sql, %s, %v", countSQL, args)
		return nil, status.Errorf(codes.Unimplemented, "method Search not implemented")
	}

	/*
		==========
		分页查询
		==========
	*/
	if req.HasTag() {
		// 如果有tag，就以tag表的创建时间排序
		query.Order("t.created_at").Desc()
	} else {
		query.Order("r.created_at").Desc()
	}
	querySQL, args := query.
		GroupBy("r.id").
		Limit(req.Page.ComputeOffset(), uint(req.Page.PageSize)).
		BuildQuery()
	s.log.Debugf("sql: %s, args: %v", querySQL, args)

	queryStmt, err := s.db.PrepareContext(ctx, querySQL)
	if err != nil {
		return nil, exception.NewInternalServerError("prepare query resource error, %s", err.Error())
	}
	defer queryStmt.Close()

	rows, err := queryStmt.QueryContext(ctx, args...)
	if err != nil {
		return nil, exception.NewInternalServerError(err.Error())
	}
	defer rows.Close()

	var (
		publicIPList, privateIPList string
	)

	for rows.Next() {
		ins := resource.NewDefaultResource()
		base := ins.Base
		info := ins.Information
		err := rows.Scan(
			&base.Id, &base.ResourceType, &base.Vendor, &base.Region, &base.Zone, &base.CreateAt, &info.ExpireAt,
			&info.Category, &info.Type, &info.Name, &info.Description,
			&info.Status, &info.UpdateAt, &base.SyncAt, &info.SyncAccount,
			&publicIPList, &privateIPList, &info.PayType, &base.DescribeHash, &base.ResourceHash,
			&base.CredentialId, &base.Domain, &base.Namespace, &base.Env, &base.UsageMode,
		)
		if err != nil {
			return nil, exception.NewInternalServerError("query resource error, %s", err.Error())
		}
		// ip地址格式化
		info.LoadPrivateIPString(privateIPList)
		info.LoadPublicIPString(publicIPList)
		set.Add(ins)
	}

	// 补充资源的标签
	if req.WithTags {
		tags, err := QueryTag(ctx, s.db, set.ResourceIds())
		if err != nil {
			return nil, err
		}
		set.UpdateTag(tags)
	}

	return set, nil
}

func (s *service) QueryTag(ctx context.Context, req *resource.QueryTagRequest) (*resource.TagSet, error) {
	return nil, status.Errorf(codes.Unimplemented, "method QueryTag not implemented")
}

func (s *service) UpdateTag(ctx context.Context, req *resource.UpdateTagRequest) (*resource.Resource, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateTag not implemented")
}

// 构造where语句
func (s *service) BuildQuery(req *resource.SearchRequest, query *sqlbuilder.Builder) {
	if req.Keywords != "" {
		if req.ExactMatch {
			// 精确匹配
			query.Where("r.name = ? OR r.id = ? OR r.private_ip = ? OR r.public_ip = ?",
				req.Keywords,
				req.Keywords,
				req.Keywords,
				req.Keywords,
			)
		} else {
			// 模糊匹配
			query.Where("r.name LIKE ? OR r.id = ? OR r.private_ip LIKE ? OR r.public_ip LIKE ?",
				"%"+req.Keywords+"%",
				req.Keywords,
				req.Keywords+"%",
				req.Keywords+"%",
			)
		}
	}
	if req.Domain != "" {
		query.Where("r.domain = ?", req.Domain)
	}
	if req.Namespace != "" {
		query.Where("r.namespace = ?", req.Namespace)
	}
	if req.Env != "" {
		query.Where("r.env = ?", req.Env)
	}
	if req.UsageMode != nil {
		query.Where("r.usage_mode = ?", req.UsageMode)
	}
	if req.Vendor != nil {
		query.Where("r.vendor = ?", req.Vendor)
	}
	if req.SyncAccount != "" {
		query.Where("r.sync_accout = ?", req.SyncAccount)
	}
	if req.Type != nil {
		query.Where("r.resource_type = ?", req.Type)
	}
	if req.Status != "" {
		query.Where("r.status = ?", req.Status)
	}

	// Tag过滤
	for i := range req.Tags {
		selector := req.Tags[i]
		if selector.Key == "" {
			continue
		}

		// 添加Key过滤条件
		query.Where("t.t_key LIKE ?", strings.ReplaceAll(selector.Key, ".*", "%"))

		// 添加Value过滤条件
		condtions := []string{}
		args := []interface{}{}
		for _, v := range selector.Values {
			condtions = append(condtions, fmt.Sprintf("t.t_value %s ?", selector.Opertor))
			args = append(args, strings.ReplaceAll(v, ".*", "%"))
		}
		if len(condtions) > 0 {
			vwhere := fmt.Sprintf("( %s )", strings.Join(condtions, selector.RelationShip()))
			query.Where(vwhere, args...)
		}
	}

}
