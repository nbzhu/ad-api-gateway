package ocean

import (
	"context"
	pb "github.com/nbzhu/ad-api-gateway-proto"
	"google.golang.org/protobuf/encoding/protojson"
)

type Api struct {
	pb.UnimplementedApiServer
}

func (s *Api) getAccessToken(ctx context.Context) string {
	return ctx.Value("access_token").(string)
}

func (s *Api) protoJson() protojson.UnmarshalOptions {
	return protojson.UnmarshalOptions{DiscardUnknown: true}
}
