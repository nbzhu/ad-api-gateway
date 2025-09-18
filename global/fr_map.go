package global

import (
	"fmt"
	"github.com/nbzhu/flowRestrictor/frPkg"
	"sync"
	"time"
)

var FrMap = sync.Map{}

func GetUniKey(fullMethod string, appNo uint64) string {
	return fmt.Sprintf("%s-%d", fullMethod, appNo)
}

func RegisterFr(uniKey string, fr *frPkg.Restrictor) error {
	if fr == nil {
		return fmt.Errorf("fr is nil")
	}
	FrMap.Store(uniKey, fr)
	go logChs(uniKey, fr)
	return nil
}

func GetFr(uniKey string) (*frPkg.Restrictor, bool) {
	fr, ok := FrMap.Load(uniKey)
	if !ok || fr == nil {
		return nil, false
	}
	return fr.(*frPkg.Restrictor), true
}

func logChs(uniKey string, fr *frPkg.Restrictor) {
	for true {
		var logs = make(map[string]interface{})
		lall := 0
		for priority, chs := range fr.Chs {
			l := len(chs)
			logs[fmt.Sprintf("%d", priority)] = l
			lall += l
		}
		if lall > 10 {
			Log("队列长度["+uniKey+"]", logs)
		}
		time.Sleep(time.Second * 10)
	}
}
