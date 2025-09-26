package ocean

import (
	"bytes"
	"context"
	"encoding/json"
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
	resp.CommonResp = &proto.CommonResp{AuthUniKey: s.getAuthUniKey(ctx)}
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

func (s *Api) FileUploadTaskCreate(ctx context.Context, req *proto.FileUploadTaskCreateReq) (*proto.FileUploadTaskCreateResp, error) {
	params := map[string]interface{}{
		"account_id":   req.Params.AccountId,
		"account_type": req.Params.AccountType,
		"filename":     req.Params.Filename,
		"video_url":    req.Params.VideoUrl,
		"is_aigc":      req.Params.IsAigc,
	}
	if len(req.Params.Labels) != 0 {
		params["labels"] = req.Params.Labels
	}
	reqBody, _ := json.Marshal(params)
	body, code, err := global.Http.Post(ctx, "https://api.oceanengine.com/open_api/2/file/upload_task/create/", reqBody,
		map[string]string{"Access-Token": s.getAccessToken(ctx), "Content-Type": "application/json"})
	if err != nil {
		return nil, err
	}
	global.Log("log", map[string]interface{}{"req": params, "AuthUniKey": s.getAuthUniKey(ctx), "resp": string(body)})
	resp := &proto.FileUploadTaskCreateResp{}
	if err = s.protoJson().Unmarshal(body, resp); err != nil {
		return nil, fmt.Errorf("反序列化失败:%s。httpCode=%d,原始数据为:%s", err.Error(), code, string(body))
	}
	resp.CommonResp = &proto.CommonResp{AuthUniKey: s.getAuthUniKey(ctx)}
	return resp, nil
}

func (s *Api) FileUploadTaskList(ctx context.Context, req *proto.FileUploadTaskListReq) (*proto.FileUploadTaskListResp, error) {
	taskIds, _ := json.Marshal(req.Params.TaskIds)
	body, code, err := global.Http.Get(ctx, "https://api.oceanengine.com/open_api/2/file/video/upload_task/list/",
		map[string]string{
			"account_id":   strconv.FormatUint(req.Params.AccountId, 10),
			"account_type": req.Params.AccountType,
			"task_ids":     string(taskIds),
		},
		map[string]string{"Access-Token": s.getAccessToken(ctx)})
	if err != nil {
		return nil, err
	}
	//global.Log("log", map[string]interface{}{"req": req, "resp": string(body)})
	resp := &proto.FileUploadTaskListResp{}
	if err = s.protoJson().Unmarshal(body, resp); err != nil {
		return nil, fmt.Errorf("反序列化失败:%s。httpCode=%d,原始数据为:%s", err.Error(), code, string(body))
	}
	return resp, nil
}
