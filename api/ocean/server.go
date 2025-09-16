package ocean

import (
	"context"
	pb "github.com/nbzhu/ad-api-gateway-proto"
)

type Api struct {
	pb.UnimplementedApiServer
}

func (s *Api) getAccessToken(ctx context.Context) string {
	return ctx.Value("access_token").(string)
}
