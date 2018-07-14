package formparser

import (
	"encoding/base64"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

const pkgName = "formparser"

// FormParser 将结构体对象转换成HTTP请求所需的KV形式, 只处理struct及*struct类型
//
// > 关键字"..." 表示该字段的子字段不继承父辈的标签, 该方式可用于struct，map类型
// 	 例如:
// 	 type Demo1 struct {
//	 		Auth 		`zwf:"..."`
// 	 }
//
// 	 type Demo2 struct {
//	 		Auth		`zwf:"auth"`
// 	 }
//
// 	 type Auth struct {
// 	  	AK *string	`zwf:"ak"`
//	 }
//
// 	 Demo1: "ak"="xxx"
// 	 Demo2: "auth.ak"="xxx"
//
//
// > 关键字"join" 可以将[]string进行按英文逗号join操作, 参见parser_test.go的TestParse例子
//
type FormParser struct {
	// 用于转换的tag名字, 类似于json序列化的json tag
	tag string

	// 用于忽略转换的符号, 类似于json序列化的"-"
	ignoreFlag string

	// 编码器
	encoders map[reflect.Kind]kindEncoder
}

func Default() *FormParser {
	return New("zwf", "-")
}

func New(tag, ignoreFlag string) *FormParser {
	if len(tag) <= 0 {
		panic(fmt.Sprintf("%s: Missing `tag` value", pkgName))
	}
	if len(ignoreFlag) <= 0 {
		panic(fmt.Sprintf("%s: Missing `ignoreFlag` value", pkgName))
	}
	p := FormParser{
		tag:        tag,
		ignoreFlag: ignoreFlag,
	}
	return p.init()
}

// ToMap the param v should be either reflect.ValueOf(struct) or reflect.ValueOf(*struct)
func (p *FormParser) ToMap(v reflect.Value) (map[string]string, error) {
	kvs, err := p.parse(v)
	if err != nil {
		return nil, err
	}

	m := make(map[string]string)
	for _, kv := range kvs {
		m[kv.K] = kv.V
	}
	return m, err
}

func (p *FormParser) Debug(v reflect.Value) {
	kvs, err := p.parse(v)
	if err != nil {
		panic(err.Error())
	}
	for _, kv := range kvs {
		fmt.Printf("%10s : %s\n", kv.K, kv.V)
	}
}

func (p *FormParser) parse(rv reflect.Value) ([]KV, error) {
	for rv.Kind() != reflect.Struct {
		if rv.Kind() == reflect.Ptr {
			rv = rv.Elem()
			continue
		}
		return nil, errors.New("Param obj is invalid, struct or non-nil *struct is needed")
	}

	var kvs []KV
	for i := 0; i < rv.NumField(); i++ {
		// 过滤掉缺省的数据
		field := rv.Field(i)
		for field.Kind() == reflect.Ptr {
			field = field.Elem() // 消除指针
		}
		if field.Kind() == reflect.Invalid {
			continue
		}
		// 过滤掉指定标签的数据
		tagK, drop := p.fieldTag(rv.Type().Field(i))
		if drop {
			continue
		}

		// 获取字段值
		kvs = append(kvs, p.encode(field, tagK)...)
	}
	return kvs, nil
}

func (p *FormParser) fieldTag(f reflect.StructField) (tag string, drop bool) {
	tag = f.Tag.Get(p.tag)
	switch tag {
	case p.ignoreFlag:
		drop = true
	case "":
		tag = f.Name
	}
	return tag, drop
}

func (p *FormParser) encode(v reflect.Value, tagK string) []KV {
	for v.Kind() == reflect.Ptr {
		v = v.Elem() // 消除指针
	}

	e, ok := p.encoders[v.Kind()]
	if !ok || e == nil {
		panic(fmt.Sprintf("Unknown type %v", v.Kind()))
	}
	return e(v, tagK)
}

func (p *FormParser) init() *FormParser {
	p.encoders = map[reflect.Kind]kindEncoder{
		reflect.String:     p.encodeString,
		reflect.Bool:       p.encodeBool,
		reflect.Int:        p.encodeInt,
		reflect.Int8:       p.encodeInt8,
		reflect.Int16:      p.encodeInt16,
		reflect.Int32:      p.encodeInt32,
		reflect.Int64:      p.encodeInt64,
		reflect.Uint:       p.encodeUint,
		reflect.Uint8:      p.encodeUint8,
		reflect.Uint16:     p.encodeUint16,
		reflect.Uint32:     p.encodeUint32,
		reflect.Uint64:     p.encodeUint64,
		reflect.Float32:    p.encodeFloat32,
		reflect.Float64:    p.encodeFloat64,
		reflect.Complex64:  p.encodeComplex64,
		reflect.Complex128: p.encodeComplex128,
		reflect.Slice:      p.encodeSlice,
		reflect.Array:      p.encodeSlice,
		reflect.Struct:     p.encodeStruct,
		reflect.Map:        p.encodeMap,
		reflect.Invalid:    p.encodeInvalid,
	}
	return p
}

func (p *FormParser) encodeString(v reflect.Value, tagK string) (rt []KV) {
	return append(rt, KV{tagK, v.Interface().(string)})
}

func (p *FormParser) encodeBool(v reflect.Value, tagK string) (rt []KV) {
	return append(rt, KV{tagK, strconv.FormatBool(v.Interface().(bool))})
}

func (p *FormParser) encodeInt(v reflect.Value, tagK string) (rt []KV) {
	return append(rt, KV{tagK, strconv.Itoa(v.Interface().(int))})
}

func (p *FormParser) encodeInt8(v reflect.Value, tagK string) (rt []KV) {
	return append(rt, KV{tagK, strconv.FormatInt(int64(v.Interface().(int8)), 10)})
}

func (p *FormParser) encodeInt16(v reflect.Value, tagK string) (rt []KV) {
	return append(rt, KV{tagK, strconv.FormatInt(int64(v.Interface().(int16)), 10)})
}

func (p *FormParser) encodeInt32(v reflect.Value, tagK string) (rt []KV) {
	return append(rt, KV{tagK, strconv.FormatInt(int64(v.Interface().(int32)), 10)})
}

func (p *FormParser) encodeInt64(v reflect.Value, tagK string) (rt []KV) {
	return append(rt, KV{tagK, strconv.FormatInt(v.Interface().(int64), 10)})
}

func (p *FormParser) encodeUint(v reflect.Value, tagK string) (rt []KV) {
	return append(rt, KV{tagK, strconv.FormatUint(uint64(v.Interface().(uint)), 10)})
}

func (p *FormParser) encodeUint8(v reflect.Value, tagK string) (rt []KV) {
	return append(rt, KV{tagK, strconv.FormatUint(uint64(v.Interface().(uint8)), 10)})
}

func (p *FormParser) encodeUint16(v reflect.Value, tagK string) (rt []KV) {
	return append(rt, KV{tagK, strconv.FormatUint(uint64(v.Interface().(uint16)), 10)})
}

func (p *FormParser) encodeUint32(v reflect.Value, tagK string) (rt []KV) {
	return append(rt, KV{tagK, strconv.FormatUint(uint64(v.Interface().(uint32)), 10)})
}

func (p *FormParser) encodeUint64(v reflect.Value, tagK string) (rt []KV) {
	return append(rt, KV{tagK, strconv.FormatUint(v.Interface().(uint64), 10)})
}

func (p *FormParser) encodeFloat32(v reflect.Value, tagK string) (rt []KV) {
	return append(rt, KV{tagK, fmt.Sprintf("%v", v.Interface().(float32))})
}

func (p *FormParser) encodeFloat64(v reflect.Value, tagK string) (rt []KV) {
	return append(rt, KV{tagK, fmt.Sprintf("%v", v.Interface().(float64))})
}

func (p *FormParser) encodeComplex64(v reflect.Value, tagK string) (rt []KV) {
	return append(rt, KV{tagK, fmt.Sprintf("%v", v.Interface().(complex64))})
}

func (p *FormParser) encodeComplex128(v reflect.Value, tagK string) (rt []KV) {
	return append(rt, KV{tagK, fmt.Sprintf("%v", v.Interface().(complex128))})
}

func (p *FormParser) encodeSlice(v reflect.Value, tagK string) (rt []KV) {
	// 如果是[]byte，则进行base64后做成KV
	b, isBytes := v.Interface().([]byte)
	if isBytes == true {
		return append(rt, KV{tagK, base64.StdEncoding.EncodeToString(b)})
	}
	// 如果是[]string,并且tagList[1]为“join”
	strList, isStrList := v.Interface().([]string)
	if isStrList {
		tagList := strings.Split(tagK, ",")
		if len(tagList) > 1 && tagList[1] == "join" {
			return append(rt, KV{tagList[0], strings.Join(strList, ",")})
		}
	}
	// 如果是非以上情况，则将每个元素单独做成KV
	for i := 0; i < v.Len(); i++ {
		rt = append(rt, p.encode(v.Index(i), fmt.Sprintf("%s.%d", tagK, i))...)
	}
	return rt
}

func (p *FormParser) encodeStruct(v reflect.Value, tagK string) (rt []KV) {
	kvs, err := p.parse(v)
	if err != nil {
		panic(fmt.Sprintf("Parse value for tagK(%s) failed, %v", tagK, err))
	}
	for i, kv := range kvs {
		if tagK != "..." { // 不继承父辈标签
			kv.K = tagK + "." + kv.K
		}
		kvs[i] = kv
	}
	rt = kvs
	return rt
}

func (p *FormParser) encodeMap(v reflect.Value, tagK string) (rt []KV) {
	keys := v.MapKeys()
	for _, k := range keys {
		keyPair := p.encode(k, "")
		valPair := p.encode(v.MapIndex(k), "")
		for _, key := range keyPair {
			for _, val := range valPair {
				var a KV
				if tagK == "..." { // 不继承父辈标签
					a.K = key.V
					a.V = val.V
				} else {
					a.K = tagK + "." + key.V
					a.V = val.V
				}
				rt = append(rt, a)
			}
		}
	}
	return rt
}

func (p *FormParser) encodeInvalid(v reflect.Value, tagK string) (rt []KV) {
	// do nothing
	return nil
}

type kindEncoder func(v reflect.Value, tagK string) (rt []KV)

type KV struct {
	K string
	V string
}

func StringPtr(v string) *string {
	return &v
}

func IntPtr(v int) *int {
	return &v
}

func Int64Ptr(v int64) *int64 {
	return &v
}

func Float64Ptr(v float64) *float64 {
	return &v
}

func BoolPtr(v bool) *bool {
	return &v
}
