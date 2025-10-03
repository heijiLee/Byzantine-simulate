package codec

import (
	"fmt"

	"codec/abstraction"

	"github.com/vmihailenco/msgpack/v5"
)

type msgpackCodec struct{} //MessagePack 포맷 parsing/serializing

func (msgpackCodec) Parse(data []byte, opts ParseOptions) (*abstraction.AbstractMessage, error) {
	var decoded map[string]interface{}
	if err := msgpack.Unmarshal(data, &decoded); err != nil {
		return nil, fmt.Errorf("msgpack decode: %w", err)
	}
	js, err := jsonFromInterface(decoded)
	if err != nil {
		return nil, err
	}
	return (jsonCodec{}).Parse(js, ParseOptions{Format: FormatJSON, OverrideMsgType: opts.OverrideMsgType})
} //MessagePack 바이트를 AbstractMessage로 변환

func (msgpackCodec) Serialize(am *abstraction.AbstractMessage, _ SerializeOptions) ([]byte, error) {
	js, err := (jsonCodec{}).Serialize(am, SerializeOptions{Format: FormatJSON}) //JSON 바이트로 변환
	if err != nil {
		return nil, err
	}
	var obj map[string]interface{}
	if err := unmarshalJSON(js, &obj); err != nil {
		return nil, err
	}
	return msgpack.Marshal(obj)
} //AbstractMessage를 MessagePack 바이트로 변환
