package type_convert

import (
	"errors"
	"strconv"
	"strings"
	"time"
)


func SliceStringToInt64(data []string) []int64 {
	res := []int64{}
	for i := range data {
		each, err := strconv.ParseInt(data[i], 10, 64)
		if err == nil {
			res = append(res, each)
		}
	}
	return res
}

func SliceInt64ToString(data []int64) []string {
	res := []string{}
	for i := range data {
		each := strconv.FormatInt(data[i], 10)
		res = append(res, each)
	}
	return res
}

func SliceInt32ToInt64(data []int32) []int64 {
	res := []int64{}
	for _, i := range data {
		res = append(res, int64(i))
	}
	return res
}

func SliceStringToInt32(data []string) []int32 {
	res := []int32{}
	for i := range data {
		each, err := strconv.ParseInt(data[i], 10, 32)
		if err == nil {
			res = append(res, int32(each))
		}
	}
	return res
}

func SliceStringToInterface(data []string) []interface{} {
	res := []interface{}{}
	for i := range data {
		res = append(res, data[i])
	}
	return res
}

func SliceInt32ToInterface(data []int32) []interface{} {
	res := []interface{}{}
	for i := range data {
		res = append(res, data[i])
	}
	return res
}

func IsEmptyInt32Slice(data []int32) bool {
	return data == nil || len(data) == 0
}

func IsEmptyStringSlice(data []string) bool {
	return data == nil || len(data) == 0
}

func IsEmptyInt64Slice(data []int64) bool {
	return data == nil || len(data) == 0
}

func IsEmptyInt64SlicePoint(data *int64) bool {
	return data == nil
}

func IsEmptyStringSlicePoint(data *string) bool {
	return data == nil
}

func Int64SliceUnique(data []int64) []int64 {
	res := make([]int64, 0, len(data))
	temp := map[int64]struct{}{}
	for _, item := range data {
		if _, ok := temp[item]; !ok {
			temp[item] = struct{}{}
			res = append(res, item)
		}
	}
	return res
}

func StringSliceUnique(data []string) []string {
	res := make([]string, 0, len(data))
	temp := map[string]struct{}{}
	for _, item := range data {
		if _, ok := temp[item]; !ok {
			temp[item] = struct{}{}
			res = append(res, item)
		}
	}
	return res
}

func SliceInt64ToInterface(data []int64) []interface{} {
	res := []interface{}{}
	for i := range data {
		res = append(res, data[i])
	}
	return res
}

func SliceInt64ToInterfacePoint(data *int64) []interface{} {
	res := []interface{}{}
	res = append(res, data)
	return res
}

func SliceInt64ToStringInterface(data *int64) []interface{} {
	res := []interface{}{}
	res = append(res, strconv.FormatInt(*data, 10))
	return res
}

func SliceStringToInterfacePoint(data *string) []interface{} {
	res := []interface{}{}
	res = append(res, data)
	return res
}

// todo: 改成StringSliceIndexOf
func IndexOf(element string, data []string) (*int, error) {
	for k, v := range data {
		if element == v {
			return &k, nil
		}
	}
	return nil, errors.New("not found") //not found.
}

func StringSliceIndexOf(element string, data []string) (int, bool) {
	if len(data) == 0 {
		return -1, false
	}
	for k, v := range data {
		if element == v {
			return k, true
		}
	}
	return -1, false //not found.
}

func Int32SliceIndexOf(element int32, data []int32) (int, bool) {
	for k, v := range data {
		if element == v {
			return k, true
		}
	}
	return 0, false //not found.
}

func Int64SliceIndexOf(element int64, data []int64) (int, bool) {
	for k, v := range data {
		if element == v {
			return k, true
		}
	}
	return 0, false //not found.
}

func InSliceInt32(element int32, slice []int32) bool {
	for _, v := range slice {
		if element == v {
			return true
		}
	}
	return false
}

func ReverseStringSlice(data []string) []string {
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
	return data
}

func DeduplicateSliceString(data []string) []string {
	res := make([]string, 0, 0)
	m := make(map[string]interface{}, 0)
	for _, d := range data {
		if _, ok := m[d]; !ok {
			res = append(res, d)
			m[d] = nil
		}
	}
	return res
}

func IntersectSliceString(slice1, slice2 []string) []string {
	resultMap := make(map[string]bool, len(slice1))
	var resultSlice []string
	for _, str := range slice1 {
		resultMap[str] = false
	}
	for _, str := range slice2 {
		if _, ok := resultMap[str]; ok {
			resultMap[str] = true
		}
	}
	for key, val := range resultMap {
		if val == true {
			resultSlice = append(resultSlice, key)
		}
	}
	return resultSlice
}



// 将输入的"[1,2,3]"类型的字符串参数转成整型数组
func ConvertRequestArrayParams2Ints(arrayParams string) []int64 {
	ints := make([]int64, 0)
	if len(arrayParams) == 0 {
		return nil
	}
	intStrings := strings.Split(strings.Trim(arrayParams, "[]"), ",")
	if len(intStrings) == 0 {
		return nil
	}
	for _, intString := range intStrings {
		if v, err := strconv.ParseInt(intString, 10, 64); err == nil {
			ints = append(ints, v)
		}
	}
	return ints
}

// 将输入的'["1","2","3"]'类型的字符串参数转成整型数组
func ConvertRequestStringArray2Ints(intStrings []string) []int64 {
	ints := make([]int64, 0)
	if len(intStrings) == 0 {
		return nil
	}
	for _, intString := range intStrings {
		if v, err := strconv.ParseInt(intString, 10, 64); err == nil {
			ints = append(ints, v)
		}
	}
	return ints
}

func ConvertRequestArrayParams2Int32s(arrayParams string) []int32 {
	ints := make([]int32, 0)
	if len(arrayParams) == 0 {
		return nil
	}
	intStrings := strings.Split(strings.Trim(arrayParams, "[]"), ",")
	if len(intStrings) == 0 {
		return nil
	}
	for _, intString := range intStrings {
		if v, err := strconv.ParseInt(intString, 10, 64); err == nil {
			ints = append(ints, int32(v))
		}
	}
	return ints
}

func ConvertRequestArrayParams2Int64s(arrayParams string) []int64 {
	ints := make([]int64, 0)
	if len(arrayParams) == 0 {
		return nil
	}
	intStrings := strings.Split(strings.Trim(arrayParams, "[]"), ",")
	if len(intStrings) == 0 {
		return nil
	}
	for _, intString := range intStrings {
		if v, err := strconv.ParseInt(intString, 10, 64); err == nil {
			ints = append(ints, v)
		}
	}
	return ints
}

func ConvertRequestArrayParams2Strings(arrayParams string) []string {
	raw := strings.Trim(arrayParams, "[] ")
	if len(raw) == 0 {
		return nil
	}
	rawArray := strings.Split(raw, ",")
	if len(rawArray) == 0 {
		return nil
	}
	convertedArray := make([]string, len(rawArray))
	for idx, str := range rawArray {
		convertedArray[idx] = strings.Trim(str, "\" ")
	}
	return convertedArray
}

func ConvertStringSliceToInt64Slice(stringSlice []string) ([]int64, error) {
	var int64List []int64
	for _, str := range stringSlice {
		id, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return nil, err
		}
		int64List = append(int64List, id)
	}
	return int64List, nil
}
func NewInt(a int) *int {
	return &a
}

func NewInt8(a int8) *int8 {
	return &a
}

func NewInt32(a int32) *int32 {
	return &a
}

func NewInt64(a int64) *int64 {
	return &a
}

func NewString(a string) *string {
	return &a
}

func NewBool(a bool) *bool {
	return &a
}

func NewFloat64(a float64) *float64 {
	return &a
}

func NewTime(a time.Time) *time.Time {
	return &a
}

func Int64ToStr(i int64) string {
	return strconv.FormatInt(i, 10)
}

func StrToInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

func StrToInt(s string) (int, error) {
	val, err := strconv.ParseInt(s, 10, 64)
	return int(val), err
}

func StrToFloat64(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

func ToStr(id int64) string {
	return strconv.FormatInt(id, 10)
}

type StrList []string

func (s StrList) Contains(val string) bool {
	for _, v := range s {
		if v == val {
			return true
		}
	}
	return false
}

func GetInt64Value(data map[string]interface{}, name string) int64 {
	if intVal, ok := data[name]; ok {
		switch v := intVal.(type) {
		case int:
			return int64(v)
		case *int:
			if v != nil {
				return int64(*v)
			}
			return 0
		case int64:
			return v
		case *int64:
			if v != nil {
				return *v
			}
			return 0
		case float64:
			return int64(v)
		case *float64:
			if v != nil {
				return int64(*v)
			}
			return 0
		case string:
			ret, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return 0
			}
			return ret
		case *string:
			if v != nil {
				ret, err := strconv.ParseInt(*v, 10, 64)
				if err != nil {
					return 0
				}
				return ret
			}
			return 0
		default:
			return 0
		}
	}
	return 0
}

func GetStringValue(data map[string]interface{}, name string) string {
	if intVal, ok := data[name]; ok {
		switch v := intVal.(type) {
		case string:
			return v
		case *string:
			if v != nil {
				return *v
			}
			return ""
		default:
			return ""
		}
	}
	return ""
}
