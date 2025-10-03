package codec

import (
	"bytes"
	"encoding/json"
	"fmt"
	"unicode/utf8"

	"codec/abstraction"
)

type Format string

const (
	FormatAuto     Format = "auto"     //자동 감지
	FormatGeneric  Format = "generic"  //Phase(k=v,...) 형태의 문자열 포맷
	FormatJSON     Format = "json"     //JSON
	FormatProtobuf Format = "protobuf" //Protocol Buffers(binary)
	FormatRLP      Format = "rlp"      //Ethereum RLP
	FormatMsgPack  Format = "msgpack"  //MessagePack
	FormatBCS      Format = "bcs"      //BCS(Binary Canonical Serialization)
)

type ParseOptions struct {
	Format               Format                  //명시된 포맷
	OverrideMsgType      string                  //메시지 타입명 덮어씀
	ProtoMessageFullName string                  //protobuf 메시지 full name
	DescriptorProvider   ProtoDescriptorProvider //protobuf 동적 parsing에 필요한 descriptor
	ProtoDiscardUnknown  bool                    //protobuf → JSON 변환 시 지원되지 않는 필드 무시
}

type SerializeOptions struct {
	Format               Format                  //출력 포맷
	ProtoMessageFullName string                  //protobuf로 직렬화할 때 대상 메시지 full name
	DescriptorProvider   ProtoDescriptorProvider //protobuf 메시지 동적 생성에 필요한 descriptor
	ProtoDiscardUnknown  bool                    //JSON→protobuf 역매핑 시 지원되지 않는 필드 무시
}

type Codec interface {
	Parse(data []byte, opts ParseOptions) (*abstraction.AbstractMessage, error)       //바이트 → AbstractMessage
	Serialize(am *abstraction.AbstractMessage, opts SerializeOptions) ([]byte, error) //AbstractMessage → 바이트
}

func DetectFormat(data []byte) Format {
	trim := bytes.TrimSpace(data) // 앞뒤 공백 제거
	if len(trim) == 0 {           //비어 있을 시
		return FormatGeneric //human-readable generic으로 간주
	}
	if (trim[0] == '{' || trim[0] == '[') && utf8.Valid(trim) { //시작 문자가 '{' 또는 '['이고 UTF-8 유효할 시
		var js json.RawMessage
		if json.Unmarshal(trim, &js) == nil { //parsing 시도하여 성공 시
			return FormatJSON //JSON으로 간주
		}
	}
	if len(trim) > 0 && (trim[0] >= 0xc0 || trim[0] <= 0xbf) {
		return FormatRLP
	} //RLP: 0xc0~0xff 범위 prefix
	if len(trim) > 0 && ((trim[0] >= 0x80 && trim[0] <= 0x9f) || (trim[0] >= 0xa0 && trim[0] <= 0xbf)) {
		return FormatMsgPack
	} //MsgPack: 0x80~0x9f, 0xa0~0xbf 범위
	return FormatProtobuf //그 외 protobuf(binary)로 간주
} //입력 바이트 검사하여 포맷 추정

func Parse(data []byte, opts ParseOptions) (*abstraction.AbstractMessage, error) {
	format := opts.Format                     //옵션에 명시된 포맷 확인
	if format == "" || format == FormatAuto { // 빈 값 또는 auto일 시
		format = DetectFormat(data) //입력으로 포맷 추정
	}
	switch format {
	case FormatGeneric:
		return (genericCodec{}).Parse(data, opts)
	case FormatJSON:
		return (jsonCodec{}).Parse(data, opts)
	case FormatProtobuf:
		return (protoCodec{}).Parse(data, opts)
	case FormatRLP:
		return (rlpCodec{}).Parse(data, opts)
	case FormatMsgPack:
		return (msgpackCodec{}).Parse(data, opts)
	case FormatBCS:
		return (bcsCodec{}).Parse(data, opts)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format) //지원되지 않는 포맷
	}
} //포맷에 맞는 codec으로 parsing

func Serialize(am *abstraction.AbstractMessage, opts SerializeOptions) ([]byte, error) {
	format := opts.Format                     //출력 포맷 확인
	if format == "" || format == FormatAuto { //지정 안 되어있을 시
		format = FormatGeneric //human-readable generic 사용
	}
	switch format {
	case FormatGeneric:
		return (genericCodec{}).Serialize(am, opts)
	case FormatJSON:
		return (jsonCodec{}).Serialize(am, opts)
	case FormatProtobuf:
		return (protoCodec{}).Serialize(am, opts)
	case FormatRLP:
		return (rlpCodec{}).Serialize(am, opts)
	case FormatMsgPack:
		return (msgpackCodec{}).Serialize(am, opts)
	case FormatBCS:
		return (bcsCodec{}).Serialize(am, opts)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format) //지원되지 않는 포맷
	}
} //포맷에 맞는 codec으로 직렬화
