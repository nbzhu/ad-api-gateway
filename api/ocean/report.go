package ocean

import (
	"context"
	"errors"
	"fmt"
	proto "github.com/nbzhu/ad-api-gateway-proto"
	"github.com/nbzhu/ad-api-gateway/global"
	"strconv"
)

func (s *Api) ReportCustomGet(ctx context.Context, req *proto.ReportReq) (*proto.ReportResp, error) {
	if req == nil || req.Params == nil || req.Params.AdvertiserId == 0 {
		return nil, errors.New("advertiser_id 必填")
	}
	if len(req.Params.OrderBy) == 0 {
		return nil, errors.New("排序字段 必填")
	}
	params := map[string]string{
		"advertiser_id": strconv.FormatInt(req.Params.AdvertiserId, 10),
		"dimensions":    s.toJson(req.Params.Dimensions),
		"metrics":       s.toJson(req.Params.Metrics),
		"start_time":    req.Params.StartTime,
		"end_time":      req.Params.EndTime,
	}

	orderBy := make([]map[string]interface{}, 0, len(req.Params.OrderBy))
	for _, item := range req.Params.OrderBy {
		orderBy = append(orderBy, map[string]interface{}{
			"field": item.Field,
			"type":  item.Type,
		})
	}
	params["order_by"] = s.toJson(orderBy)

	if len(req.Params.Filters) > 0 {
		fs := make([]map[string]interface{}, 0, len(req.Params.Filters))
		for _, f := range req.Params.Filters {
			fs = append(fs, map[string]interface{}{
				"field":    f.Field,
				"type":     int32(f.Type),
				"operator": int32(f.Operator),
				"values":   f.Values,
			})
		}
		params["filters"] = s.toJson(fs)
	}

	if req.Params.Page != 0 {
		params["page"] = string(req.Params.Page)
	}
	if req.Params.PageSize != 0 {
		params["page_size"] = string(req.Params.PageSize)
	}

	body, code, err := global.Http.Get(ctx, "https://api.oceanengine.com/open_api/v3.0/report/custom/get/", params,
		map[string]string{"Access-Token": s.getAccessToken(ctx)})
	if err != nil {
		return nil, err
	}

	resp := &proto.ReportResp{}
	if err = s.protoJson().Unmarshal(body, resp); err != nil {
		return nil, fmt.Errorf("反序列化失败: %s。httpCode=%d, 原始数据为: %s", err.Error(), code, string(body))
	}
	return resp, nil
}
