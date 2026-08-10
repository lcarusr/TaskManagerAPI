package main

import (
	"encoding/json"

	"github.com/go-kratos/kratos/v3/encoding"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// jsonCodec 覆盖 kratos 默认的 json codec：
// - protobuf message 使用 protojson（支持 enum 字符串解析，如 "todo" / "in_progress"）
// - 其他类型回退标准库 encoding/json
type jsonCodec struct{}

var _ encoding.Codec = jsonCodec{}

// protoJSONMarshalOptions 输出空字段并使用 proto 原始字段名（snake_case），
// 使响应体与作业数据模型示例一致（created_at / updated_at）。
var protoJSONMarshalOptions = protojson.MarshalOptions{
	EmitUnpopulated: true,
	UseProtoNames:    true,
}

func (jsonCodec) Marshal(v any) ([]byte, error) {
	if m, ok := v.(proto.Message); ok {
		return protoJSONMarshalOptions.Marshal(m)
	}
	return json.Marshal(v)
}

func (jsonCodec) Unmarshal(data []byte, v any) error {
	if len(data) == 0 {
		return nil
	}
	if m, ok := v.(proto.Message); ok {
		return protojson.Unmarshal(data, m)
	}
	return json.Unmarshal(data, v)
}

func (jsonCodec) Name() string {
	return "json"
}

func init() {
	encoding.RegisterCodec(jsonCodec{})
}
