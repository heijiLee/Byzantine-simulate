package codec

import (
	"codec/abstraction"
	"fmt"

	"github.com/ethereum/go-ethereum/rlp"
)

type rlpCodec struct{} //rlp 포맷 parsing/serializing

func (rlpCodec) Parse(data []byte, opts ParseOptions) (*abstraction.AbstractMessage, error) {
	var raw []byte
	if err := rlp.DecodeBytes(data, &raw); err == nil {
		return (jsonCodec{}).Parse(raw, ParseOptions{Format: FormatJSON, OverrideMsgType: opts.OverrideMsgType})
	}
	var decoded interface{}
	if err := rlp.DecodeBytes(data, &decoded); err != nil {
		return nil, fmt.Errorf("rlp decode: %w", err)
	}
	js, err := jsonFromInterface(decoded)
	if err != nil {
		return nil, err
	}
	return (jsonCodec{}).Parse(js, ParseOptions{Format: FormatJSON, OverrideMsgType: opts.OverrideMsgType})
} //rlp 바이트를 AbstractMessage로 변환

func (rlpCodec) Serialize(am *abstraction.AbstractMessage, _ SerializeOptions) ([]byte, error) {
	js, err := (jsonCodec{}).Serialize(am, SerializeOptions{Format: FormatJSON}) //JSON 바이트로 변환
	if err != nil {
		return nil, err
	}
	return rlp.EncodeToBytes(js) //
} //AbstractMessage를 rlp 바이트로 변환
