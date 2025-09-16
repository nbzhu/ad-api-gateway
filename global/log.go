package global

import (
	"encoding/json"
	"log"
)

func Log(key string, m map[string]interface{}) {
	s, _ := json.Marshal(&m)
	log.Println(key + " " + string(s))
}
