package initialize

import (
	"context"
	"errors"
	"fmt"
	pb "github.com/nbzhu/ad-api-gateway-proto"
	"github.com/nbzhu/ad-api-gateway/global"
	"github.com/nbzhu/flowRestrictor/frClient"
	"github.com/nbzhu/flowRestrictor/frPkg"
	"golang.org/x/sync/singleflight"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"sync"
)

var whitelist = map[string]struct{}{
	"/ocean.Api/QueueLen": {},
}

var sfgRestrictor singleflight.Group

var (
	fdCache sync.Map // map[string]*confFieldDesc
	sfGroup singleflight.Group
)

type confFieldDesc struct {
	confFD              protoreflect.FieldDescriptor // conf
	accessTokenMapFD    protoreflect.FieldDescriptor // conf.access_token_map
	priorityFD          protoreflect.FieldDescriptor // conf.priority
	highPriorityLenFD   protoreflect.FieldDescriptor // conf.high_priority_len
	mediumPriorityLenFD protoreflect.FieldDescriptor // conf.medium_priority_len
	lowPriorityLenFD    protoreflect.FieldDescriptor // conf.low_priority_len

	// AppConf
	appConfQpsFD         protoreflect.FieldDescriptor // AppConf.qps
	appConfAccessTokenFD protoreflect.FieldDescriptor // AppConf.access_token
}

func UnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if _, ok := whitelist[info.FullMethod]; ok {
		return handler(ctx, req)
	}
	pm, ok := req.(proto.Message)
	if !ok {
		return nil, errors.New("参数异常,func=" + info.FullMethod)
	}
	conf, err := extractConfFromProtoMessage(pm)
	if err != nil {
		return nil, err
	}
	frs, err := initFqQueue(info, conf)
	if err != nil {
		return nil, fmt.Errorf("初始化队列失败:%s ,func=%s", err.Error(), info.FullMethod)
	}
	if len(frs) == 0 {
		return nil, fmt.Errorf("初始化队列失败 ,func=%s", info.FullMethod)
	}
	var resp interface{}
	var errInner error
	var wg sync.WaitGroup
	wg.Add(1)
	frwt := chooseFr(frs, conf)
	ctx = context.WithValue(ctx, "access_token", frwt.accessToken)
	ctx = context.WithValue(ctx, "auth_uni_key", frwt.authUniKey)
	if err = frClient.TryToDo(frwt.fr, frPkg.Priority(conf.Priority), frPkg.QueueData{
		Func: func() error {
			resp, errInner = handler(ctx, req)
			return errInner
		},
		FinalFunc: func(err error) {
			wg.Done()
		},
		Title: info.FullMethod,
		Ctx:   ctx,
	}); err != nil {
		return nil, errors.New("当前请求队列已满,func=" + info.FullMethod)
	}
	wg.Wait()
	if errInner != nil {
		global.Log("错误日志", map[string]interface{}{"Method": info.FullMethod, "req": req, "resp": resp, "err": errInner.Error()})
	}
	return resp, errInner
}

func chooseFr(frs []frWithToken, conf *pb.Conf) frWithToken {
	if len(frs) == 0 {
		return frs[0]
	}
	if conf.AuthUniKey != "" {
		for _, m := range conf.AccessTokenMap {
			for _, fr := range frs {
				if m.AccessToken == fr.accessToken {
					fr.authUniKey = m.AuthUniKey
					return fr
				}
			}
		}
	}
	var frwt frWithToken
	for _, item := range frs {
		if frwt.accessToken == "" {
			frwt = item
			continue
		}
		if len(item.fr.Chs[frPkg.Priority(conf.Priority)]) < len(frwt.fr.Chs[frPkg.Priority(conf.Priority)]) {
			frwt = item
		}
	}
	return frwt
}

type frWithToken struct {
	fr          *frPkg.Restrictor
	accessToken string
	authUniKey  string
}

func initFqQueue(info *grpc.UnaryServerInfo, conf *pb.Conf) ([]frWithToken, error) {
	var frs = make([]frWithToken, 0)
	for devKey, appConf := range conf.AccessTokenMap {
		queueName := global.GetUniKey(info.FullMethod, devKey)
		out, err, _ := sfgRestrictor.Do(queueName, func() (interface{}, error) {
			fr, ok := global.GetFr(queueName)
			if ok {
				return frWithToken{
					fr:          fr,
					accessToken: appConf.AccessToken,
				}, nil
			}
			global.Log("初始化fr", map[string]interface{}{"queueName": queueName, "FullMethod": info.FullMethod, "conf": conf})
			if conf.HighPriorityLen == 0 {
				conf.HighPriorityLen = 999
			}
			if conf.MediumPriorityLen == 0 {
				conf.MediumPriorityLen = 3999
			}
			if conf.LowPriorityLen == 0 {
				conf.LowPriorityLen = 9999
			}
			fr = frClient.New(int(appConf.Qps), frPkg.PriorityStruct{
				HighPriorityLen:   int(conf.HighPriorityLen),
				MediumPriorityLen: int(conf.MediumPriorityLen),
				LowPriorityLen:    int(conf.LowPriorityLen),
			}).SetRestrictorType(frPkg.RestrictorTypeSlidingWindow).SetNoticeRetryTimes(0).SetMaxRetryTimes(0)
			err := global.RegisterFr(queueName, fr)
			if err != nil {
				return nil, err
			}
			return frWithToken{
				fr:          fr,
				accessToken: appConf.AccessToken,
				authUniKey:  appConf.AuthUniKey,
			}, nil
		})
		if err != nil {
			return nil, err
		}
		frs = append(frs, out.(frWithToken))
	}
	return frs, nil
}

func extractConfFromProtoMessage(m proto.Message) (*pb.Conf, error) {
	if m == nil {
		return nil, errors.New("conf 参数异常0")
	}
	mr := m.ProtoReflect()
	desc, err := getOrBuildConfDesc(mr)
	if err != nil {
		return nil, errors.New("conf 参数异常，" + err.Error())
	}
	val := mr.Get(desc.confFD)
	if !val.IsValid() || !val.Message().IsValid() {
		return nil, errors.New("conf is empty")
	}
	confMsg := val.Message()

	var conf pb.Conf
	if desc.priorityFD != nil && confMsg.Has(desc.priorityFD) {
		conf.Priority = pb.Priority(confMsg.Get(desc.priorityFD).Enum())
	}

	if desc.highPriorityLenFD != nil && confMsg.Has(desc.highPriorityLenFD) {
		conf.HighPriorityLen = uint32(confMsg.Get(desc.highPriorityLenFD).Uint())
	}
	if desc.mediumPriorityLenFD != nil && confMsg.Has(desc.mediumPriorityLenFD) {
		conf.MediumPriorityLen = uint32(confMsg.Get(desc.mediumPriorityLenFD).Uint())
	}
	if desc.lowPriorityLenFD != nil && confMsg.Has(desc.lowPriorityLenFD) {
		conf.LowPriorityLen = uint32(confMsg.Get(desc.lowPriorityLenFD).Uint())
	}

	if desc.accessTokenMapFD == nil {
		return nil, errors.New("conf 不包含 access_token_map 字段")
	}
	mp := confMsg.Get(desc.accessTokenMapFD).Map()
	if mp.Len() == 0 {
		return nil, errors.New("access_token_map 不能为空")
	}

	conf.AccessTokenMap = make(map[string]*pb.AppConf)
	var rangeErr error
	mp.Range(func(k protoreflect.MapKey, v protoreflect.Value) bool {
		key := k.String()
		if !v.IsValid() || !v.Message().IsValid() {
			rangeErr = fmt.Errorf("access_token_map[%d] value invalid", key)
			return true
		}

		am := v.Message()
		appConf := &pb.AppConf{}

		if desc.appConfQpsFD != nil && am.Has(desc.appConfQpsFD) {
			appConf.Qps = uint32(am.Get(desc.appConfQpsFD).Uint())
		}
		if desc.appConfAccessTokenFD != nil && am.Has(desc.appConfAccessTokenFD) {
			appConf.AccessToken = am.Get(desc.appConfAccessTokenFD).String()
		}
		if appConf.Qps == 0 {
			rangeErr = fmt.Errorf("access_token_map[%d] qps 不能为空", key)
			conf.AccessTokenMap[key] = appConf
			return true
		}
		conf.AccessTokenMap[key] = appConf
		return true
	})
	if rangeErr != nil {
		return nil, rangeErr
	}
	return &conf, nil

}

func getOrBuildConfDesc(m protoreflect.Message) (*confFieldDesc, error) {
	if m == nil {
		return nil, errors.New("nil message")
	}
	msgName := string(m.Descriptor().FullName())

	if v, ok := fdCache.Load(msgName); ok {
		if v == nil {
			return nil, errors.New("缓存中无conf参数")
		}
		return v.(*confFieldDesc), nil
	}

	val, err, _ := sfGroup.Do(msgName, func() (interface{}, error) {
		if vv, ok2 := fdCache.Load(msgName); ok2 {
			return vv, nil
		}

		fd := m.Descriptor().Fields().ByName("conf")
		if fd == nil {
			return nil, errors.New("无conf参数")
		}

		confMsgDesc := fd.Message()
		if confMsgDesc == nil {
			return nil, errors.New("conf 字段不是 message 类型")
		}

		// 找 conf 下的字段
		priorityFD := confMsgDesc.Fields().ByName("priority")
		accessTokenMapFD := confMsgDesc.Fields().ByName("access_token_map")
		highFD := confMsgDesc.Fields().ByName("high_priority_len")
		mediumFD := confMsgDesc.Fields().ByName("medium_priority_len")
		lowFD := confMsgDesc.Fields().ByName("low_priority_len")

		// 准备 AppConf 的字段描述符（map 的 value）
		var appConfQpsFD, appConfAccessTokenFD protoreflect.FieldDescriptor
		if accessTokenMapFD != nil && accessTokenMapFD.Message() != nil {
			// accessTokenMapFD.Message() 是 map entry 的 message descriptor，通常含有 "key" 和 "value"。
			entry := accessTokenMapFD.Message()
			valueFD := entry.Fields().ByName("value")
			if valueFD != nil && valueFD.Message() != nil {
				// valueFD.Message() 是 AppConf 的 message descriptor
				appConfMsgDesc := valueFD.Message()
				appConfQpsFD = appConfMsgDesc.Fields().ByName("qps")
				appConfAccessTokenFD = appConfMsgDesc.Fields().ByName("access_token")
			}
		}

		desc := &confFieldDesc{
			confFD:               fd,
			accessTokenMapFD:     accessTokenMapFD,
			priorityFD:           priorityFD,
			highPriorityLenFD:    highFD,
			mediumPriorityLenFD:  mediumFD,
			lowPriorityLenFD:     lowFD,
			appConfQpsFD:         appConfQpsFD,
			appConfAccessTokenFD: appConfAccessTokenFD,
		}
		fdCache.Store(msgName, desc)
		return desc, nil
	})
	if err != nil {
		return nil, err
	}
	if val == nil {
		return nil, errors.New("无conf参数2")
	}
	return val.(*confFieldDesc), nil
}
