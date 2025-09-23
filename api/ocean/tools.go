package ocean

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	proto "github.com/nbzhu/ad-api-gateway-proto"
	"github.com/nbzhu/ad-api-gateway/global"
	"github.com/nbzhu/flowRestrictor/frPkg"
	"strconv"
)

func (s *Api) QueueLen(ctx context.Context, req *proto.QueueLenReq) (*proto.QueueLenResp, error) {
	uniKey := global.GetUniKey(fmt.Sprintf("/ocean.Api/%s", req.Method), req.DevKey)
	fr, ok := global.GetFr(uniKey)
	if !ok {
		return nil, errors.New("队列异常,uniKey=" + uniKey)
	}
	ch, ok2 := fr.Chs[frPkg.Priority(req.Priority)]
	if !ok2 {
		return nil, errors.New("队列对应优先级不存在,uniKey=" + uniKey)
	}
	return &proto.QueueLenResp{Length: int32(len(ch))}, nil
}

func (s *Api) Awemes(ctx context.Context, req *proto.AwemesReq) (*proto.AwemesResp, error) {
	body, code, err := global.Http.Get(ctx, "https://ad.oceanengine.com/open_api/2/tools/ies_account_search/",
		map[string]string{"advertiser_id": strconv.FormatUint(req.Params.AdvertiserId, 10)},
		map[string]string{"Access-Token": s.getAccessToken(ctx)})
	if err != nil {
		return nil, err
	}
	resp := &proto.AwemesResp{}
	if err = s.protoJson().Unmarshal(body, resp); err != nil {
		return nil, fmt.Errorf("反序列化失败:%s。httpCode=%d,原始数据为:%s", err.Error(), code, string(body))
	}
	return resp, nil
}

func (s *Api) MicroAppList(ctx context.Context, req *proto.MicroAppListReq) (*proto.MicroAppListResp, error) {
	if req.Params == nil {
		return nil, errors.New("参数必填")
	}
	query := map[string]string{"advertiser_id": strconv.FormatInt(req.Params.AdvertiserId, 10)}
	if req.Params.Filtering != nil {
		filter := map[string]interface{}{}
		if req.Params.Filtering.SearchKey != "" {
			filter["search_key"] = req.Params.Filtering.SearchKey
		}
		if req.Params.Filtering.AuditStatus != "" {
			filter["audit_status"] = req.Params.Filtering.AuditStatus
		}
		if req.Params.Filtering.SearchType != "" {
			filter["search_type"] = req.Params.Filtering.SearchType
		}
		if req.Params.Filtering.CreateTime != nil {
			filter["create_time"] = req.Params.Filtering.CreateTime
		}
		filtering, _ := json.Marshal(filter)
		query["filtering"] = string(filtering)
	}

	if req.Params.Page != 0 {
		query["page"] = strconv.Itoa(int(req.Params.Page))
	}
	if req.Params.PageSize != 0 {
		query["page_size"] = strconv.Itoa(int(req.Params.PageSize))
	}

	body, code, err := global.Http.Get(ctx, "https://api.oceanengine.com/open_api/v3.0/tools/micro_app/list/", query,
		map[string]string{"Access-Token": s.getAccessToken(ctx)})
	if err != nil {
		return nil, err
	}
	//global.Log("log", map[string]interface{}{"req": req, "resp": string(body)})
	resp := &proto.MicroAppListResp{}
	if err = s.protoJson().Unmarshal(body, resp); err != nil {
		return nil, fmt.Errorf("反序列化失败:%s。httpCode=%d,原始数据为:%s", err.Error(), code, string(body))
	}
	return resp, nil
}

func (s *Api) MicroAppDetail(ctx context.Context, req *proto.MicroAppDetailReq) (*proto.MicroAppDetailResp, error) {
	if req.Params == nil || req.Params.Filtering == nil || req.Params.Filtering.InstanceId == 0 {
		return nil, errors.New("资产id 必填")
	}
	query := map[string]string{"advertiser_id": strconv.FormatInt(req.Params.AdvertiserId, 10)}
	filter := map[string]interface{}{
		"instance_id": req.Params.Filtering.InstanceId,
	}
	filtering, _ := json.Marshal(filter)
	query["filtering"] = string(filtering)
	if req.Params.Page != 0 {
		query["page"] = strconv.Itoa(int(req.Params.Page))
	}
	if req.Params.PageSize != 0 {
		query["page_size"] = strconv.Itoa(int(req.Params.PageSize))
	}

	body, code, err := global.Http.Get(ctx, "https://api.oceanengine.com/open_api/v3.0/tools/asset_link/list/", query,
		map[string]string{"Access-Token": s.getAccessToken(ctx)})
	if err != nil {
		return nil, err
	}
	resp := &proto.MicroAppDetailResp{}
	if err = s.protoJson().Unmarshal(body, resp); err != nil {
		return nil, fmt.Errorf("反序列化失败:%s。httpCode=%d,原始数据为:%s", err.Error(), code, string(body))
	}
	return resp, nil
}
