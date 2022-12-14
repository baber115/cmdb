package api

import (
	"codeup.aliyun.com/baber/go/cmdb/apps/resource"
	"github.com/emicklei/go-restful/v3"
	"github.com/infraboard/mcube/exception"
	"github.com/infraboard/mcube/http/response"
)

func (h *handler) SearchResource(r *restful.Request, w *restful.Response) {
	query, err := resource.NewSearchRequestFromHTTP(r.Request)
	if err != nil {
		response.Failed(w, exception.NewBadRequest("new request error, %s", err))
		return
	}

	set, err := h.service.Search(r.Request.Context(), query)
	if err != nil {
		response.Failed(w, err)
		return
	}
	response.Success(w, set)
}

//func (h *handler) AddTag(r *restful.Request, w *restful.Response) {
//	req := resource.NewUpdateTagRequest(r.PathParameter("id"), resource.UpdateAction_ADD)
//	if err := request.GetDataFromRequest(r.Request, &req.Tags); err != nil {
//		response.Failed(w, err)
//		return
//	}
//	set, err := h.service.UpdateTag(r.Request.Context(), req)
//	if err != nil {
//		response.Failed(w, err)
//		return
//	}
//	response.Success(w, set)
//}
//
//func (h *handler) RemoveTag(r *restful.Request, w *restful.Response) {
//	req := resource.NewUpdateTagRequest(r.PathParameter("id"), resource.UpdateAction_REMOVE)
//	if err := request.GetDataFromRequest(r.Request, &req.Tags); err != nil {
//		response.Failed(w, err)
//		return
//	}
//	set, err := h.service.UpdateTag(r.Request.Context(), req)
//	if err != nil {
//		response.Failed(w, err)
//		return
//	}
//	response.Success(w, set)
//}
