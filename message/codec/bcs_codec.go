package codec

import (
	"fmt"

	"codec/abstraction"

	bcs "github.com/fardream/go-bcs/bcs"
)

type bcsCodec struct{} //bcs 포맷 parsing/serializing

func (bcsCodec) Parse(data []byte, opts ParseOptions) (*abstraction.AbstractMessage, error) {
	var raw []byte
	if _, err := bcs.Unmarshal(data, &raw); err == nil {
		return (jsonCodec{}).Parse(raw, ParseOptions{Format: FormatJSON, OverrideMsgType: opts.OverrideMsgType})
	}
	var decoded map[string]interface{}
	if _, err := bcs.Unmarshal(data, &decoded); err != nil {
		return nil, fmt.Errorf("bcs decode: %w", err)
	}
	js, err := jsonFromInterface(decoded)
	if err != nil {
		return nil, err
	}
	return (jsonCodec{}).Parse(js, ParseOptions{Format: FormatJSON, OverrideMsgType: opts.OverrideMsgType})
} //bcs 바이트를 AbstractMessage로 변환

func (bcsCodec) Serialize(am *abstraction.AbstractMessage, _ SerializeOptions) ([]byte, error) {
	js, err := (jsonCodec{}).Serialize(am, SerializeOptions{Format: FormatJSON}) //JSON 바이트로 변환
	if err != nil {
		return nil, err
	}
	return bcs.Marshal(js) //bcs payload를 []byte(JSON) 형태로 serializing
} //AbstractMessage를 BCS 바이트로 변환
