package ocean

import (
	"context"
	"encoding/json"
	pb "github.com/nbzhu/ad-api-gateway-proto"
	"github.com/nbzhu/ad-api-gateway/global"
	"google.golang.org/protobuf/encoding/protojson"
)

type Api struct {
	pb.UnimplementedApiServer
}

func (s *Api) getAccessToken(ctx context.Context) string {
	return ctx.Value("access_token").(string)
}

func (s *Api) getAuthUniKey(ctx context.Context) string {
	return ctx.Value("auth_uni_key").(string)
}

func (s *Api) protoJson() protojson.UnmarshalOptions {
	return protojson.UnmarshalOptions{DiscardUnknown: true}
}

func (s *Api) toJson(data interface{}) string {
	if data == nil {
		return ""
	}
	switch v := data.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	default:
		b, err := json.Marshal(data)
		if err != nil {
			global.Log("toJsonErr", map[string]interface{}{"原始值": data, "err": err})
			return ""
		}
		return string(b)
	}
}
