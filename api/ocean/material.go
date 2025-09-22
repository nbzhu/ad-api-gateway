package ocean

import (
	"bytes"
	"context"
	"fmt"
	proto "github.com/nbzhu/ad-api-gateway-proto"
	"github.com/nbzhu/ad-api-gateway/global"
	"mime/multipart"
	"net/http"
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

func (s *Api) FileImageAd(ctx context.Context, req *proto.FileImageAdReq) (*proto.FileImageAdResp, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	err := writer.WriteField("advertiser_id", strconv.FormatUint(req.Params.AdvertiserId, 10))
	if err != nil {
		return nil, err
	}
	err = writer.WriteField("upload_type", "UPLOAD_BY_URL")
	if err != nil {
		return nil, err
	}
	err = writer.WriteField("filename", req.Params.Filename)
	if err != nil {
		return nil, err
	}
	err = writer.WriteField("image_url", req.Params.ImageUrl)
	if err != nil {
		return nil, err
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequest("POST", "https://api.oceanengine.com/open_api/2/file/image/ad/", body)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.Header.Set("Access-Token", s.getAccessToken(ctx))
	respBody, code, err := global.Http.Do(request)
	if err != nil {
		global.Log("log", map[string]interface{}{"req": req, "err": err, "code": code, "resp": string(respBody)})
		return nil, err
	}
	resp := &proto.FileImageAdResp{}
	if err = s.protoJson().Unmarshal(respBody, resp); err != nil {
		return nil, fmt.Errorf("反序列化失败:%s。httpCode=%d,原始数据为:%s", err.Error(), code, string(respBody))
	}
	return resp, nil
}
