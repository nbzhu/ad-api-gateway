package ocean

import (
	"context"
	"fmt"
	proto "github.com/nbzhu/ad-api-gateway-proto"
	"github.com/nbzhu/ad-api-gateway/global"
	"strconv"
)

func (s *Api) VideoCoverSuggest(ctx context.Context, req *proto.VideoCoverSuggestReq) (*proto.VideoCoverSuggestResp, error) {
	body, code, err := global.Http.Get(ctx, "https://ad.oceanengine.com/open_api/2/tools/video_cover/suggest/",
		map[string]string{
			"advertiser_id": strconv.FormatUint(req.Params.AdvertiserId, 10),
			"video_id":      req.Params.VideoId},
		map[string]string{"Access-Token": s.getAccessToken(ctx)})
	if err != nil {
		return nil, err
	}
	//global.Log("log", map[string]interface{}{"req": req, "resp": string(body)})
	resp := &proto.VideoCoverSuggestResp{}
	if err = s.protoJson().Unmarshal(body, resp); err != nil {
		return nil, fmt.Errorf("反序列化失败:%s。httpCode=%d,原始数据为:%s", err.Error(), code, string(body))
	}
	return resp, nil
}
