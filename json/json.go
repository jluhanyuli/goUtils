package json


import (
	jsoniter "github.com/json-iterator/go"
)

//经过自测验证 jsoniter 速度大于decode/json
// StructToJsonString 将Struct转换为Json格式的字符串
func StructToJsonString(v interface{}) (string, error) {
	str, err := MarshalToString(v)
	if err != nil {
		return "", err
	}
	return str, nil
}

func UnmarshalFromString(str string, v interface{}) error {
	return jsoniter.UnmarshalFromString(str, v)
}

func MarshalToString(v interface{}) (string, error) {
	return jsoniter.MarshalToString(v)
}
