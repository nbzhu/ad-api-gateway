package ocean

import (
	"context"
	"fmt"
	proto "github.com/nbzhu/ad-api-gateway-proto"
	"github.com/nbzhu/ad-api-gateway/global"
)

func (s *Api) ProjectCreate(ctx context.Context, req *proto.ProjectCreateReq) (*proto.ProjectCreateResp, error) {
	body, code, err := global.Http.Post(ctx, "https://api.oceanengine.com/open_api/v3.0/project/create/",
		req.Params.Body,
		map[string]string{"Access-Token": s.getAccessToken(ctx), "Content-Type": "application/json"})
	if err != nil {
		return nil, err
	}
	resp := &proto.ProjectCreateResp{}
	if err = s.protoJson().Unmarshal(body, resp); err != nil {
		return nil, fmt.Errorf("反序列化失败:%s。httpCode=%d,原始数据为:%s", err.Error(), code, string(body))
	}
	return resp, nil
}
