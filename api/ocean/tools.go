package ocean

import (
	"context"
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
