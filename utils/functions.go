package utils

import (
	"encoding/json"
	"net/url"
	"strconv"
)

func InterfaceToString(val interface{}) string {
	switch val.(type) {
	case string:
		return val.(string)
	case int:
		return strconv.FormatInt(int64(val.(int)), 10)
	case int64:
		return strconv.FormatInt(val.(int64), 10)
	case uint64:
		return strconv.FormatUint(val.(uint64), 10)
	case float32:
		return strconv.FormatFloat(float64(val.(float32)), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(val.(float64), 'f', -1, 64)
	default:
		bytes, _ := json.Marshal(val)
		return string(bytes)
	}
}

// 将struct转成url的参数，并去除有json omitempty且值为空的参数，类似param1=1&param2=2
func StructToUrlQuery(m interface{}) (string, error) {
	query := ""

	b, err := json.Marshal(m)
	if err != nil {
		return query, err
	}

	var f map[string]interface{}
	err = json.Unmarshal(b, &f)
	if err != nil {
		return query, err
	}

	v := url.Values{}
	for key, value := range f {
		v.Set(key, InterfaceToString(value))
	}
	query = v.Encode()
	return query, nil
}
