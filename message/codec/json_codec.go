package codec

import (
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"codec/abstraction"
)

type jsonCodec struct{} //JSON parsing/serializing

func (jsonCodec) Parse(data []byte, opts ParseOptions) (*abstraction.AbstractMessage, error) {
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("json unmarshal: %w", err)
	}
	am := &abstraction.AbstractMessage{
		Extras:     map[string][]byte{},          //표준화되지 않은 필드
		RawPayload: append([]byte(nil), data...), //원본 JSON
	} //AbstractMessage 초기화
	if opts.OverrideMsgType != "" { //타입 지정 시
		am.Type = abstraction.MsgType(opts.OverrideMsgType)
	} else if v, ok := m["type"]; ok { //JSON에 type 키 있을 시
		if s, ok2 := v.(string); ok2 { //문자열일 때만 처리
			if mapped, ok3 := PhaseSynonyms[s]; ok3 { //유의어 정규화
				am.Type = abstraction.MsgType(mapped)
			} else {
				am.Type = abstraction.MsgType(s) //그대로 사용
			}
		}
	}
	for kRaw, v := range m { //kRaw는 원본 키
		key := kRaw
		if mapped, ok := FieldSynonyms[key]; ok { //유의어 정규화
			key = mapped
		}
		switch key {
		case "Height":
			am.Height = toBigIntPtr(v)
		case "Round":
			am.Round = toBigIntPtr(v)
		case "View":
			am.View = toBigIntPtr(v)
		case "BlockHash":
			am.BlockHash = toString(v)
		case "PrevHash":
			am.PrevHash = toString(v)
		case "Timestamp":
			am.Timestamp = toTime(v)
		case "Proposer":
			am.Proposer = toString(v)
		case "Validator":
			am.Validator = toString(v)
		case "Signature":
			am.Signature = toString(v)
		case "CommitSeals":
			am.CommitSeals = toStringSlice(v)
		case "ViewChanges": //원소가 객체인 JSON array
			if arr, ok := v.([]interface{}); ok {
				am.ViewChanges = make([]abstraction.ViewChangeEntry, 0, len(arr))
				for _, iv := range arr {
					if obj, ok := iv.(map[string]interface{}); ok {
						am.ViewChanges = append(am.ViewChanges, abstraction.ViewChangeEntry{
							View:      toBigIntPtr(obj["view"]),
							Height:    toBigIntPtr(obj["height"]),
							Validator: toString(obj["validator"]),
							Signature: toString(obj["signature"]),
						})
					}
				}
			}
		case "type":
		default:
			b, _ := json.Marshal(v)
			am.Extras[kRaw] = b //표준 필드가 아닐 시 Extras
		}
	}
	return am, nil //parsing 결과 반환
} //JSON 바이트를 AbstractMessage로 변환

func (jsonCodec) Serialize(am *abstraction.AbstractMessage, _ SerializeOptions) ([]byte, error) {
	out := map[string]interface{}{
		"type": string(am.Type),
	}
	if am.Height != nil { //필드가 존재할 시
		out["height"] = am.Height.String() // big.Int는 문자열로 출력
	}
	if am.Round != nil {
		out["round"] = am.Round.String()
	}
	if am.View != nil {
		out["view"] = am.View.String()
	}
	if am.BlockHash != "" {
		out["block_hash"] = am.BlockHash
	}
	if am.PrevHash != "" {
		out["prev_hash"] = am.PrevHash
	}
	if !am.Timestamp.IsZero() {
		out["timestamp"] = am.Timestamp.UTC().Format(time.RFC3339) // RFC3339로 표준화
	}
	if am.Proposer != "" {
		out["proposer"] = am.Proposer
	}
	if am.Validator != "" {
		out["validator"] = am.Validator
	}
	if am.Signature != "" {
		out["signature"] = am.Signature
	}
	if len(am.CommitSeals) > 0 {
		out["commit_seals"] = am.CommitSeals
	}
	if len(am.ViewChanges) > 0 {
		vc := make([]map[string]interface{}, 0, len(am.ViewChanges))
		for _, e := range am.ViewChanges {
			item := map[string]interface{}{
				"view":      strOrNil(e.View),   //nil일 시 null, 아닐 시 문자열
				"height":    strOrNil(e.Height), //nil일 시 null, 아닐 시 문자열
				"validator": e.Validator,
				"signature": e.Signature,
			}
			vc = append(vc, item)
		}
		out["view_changes"] = vc
	}
	// Extras 병합
	for k, v := range am.Extras {
		if _, exists := out[k]; exists {
			continue
		}
		var any interface{}
		if err := json.Unmarshal(v, &any); err == nil {
			out[k] = any
		} else {
			out[k] = string(v)
		}
	} //동일 key 가진 필드 중 표준 필드 우선
	return json.Marshal(out) //JSON 바이트 반환
} //AbstractMessage를 JSON 바이트로 변환

func toString(v interface{}) string {
	if v == nil { //nil일 시 빈 문자열
		return ""
	}
	switch t := v.(type) {
	case string:
		return t //문자열
	case []byte:
		return string(t) // 바이트 → 문자열
	default:
		b, _ := json.Marshal(t)
		return string(b) //JSON으로 직렬화한 바이트를 문자열로 반환
	}
} //interface{} 값을 문자열로 변환

func toBigIntPtr(v interface{}) *big.Int {
	if v == nil {
		return nil
	}
	switch t := v.(type) {
	case float64: //일반 JSON 숫자는 float64
		return big.NewInt(int64(t)) //정수 부분만 사용
	case json.Number: //UseNumber 사용 시
		if i, err := t.Int64(); err == nil {
			return big.NewInt(i)
		}
		if bi, ok := new(big.Int).SetString(t.String(), 10); ok { // 큰 정수 문자열 파싱
			return bi
		}
	case string: //"1000" 등 문자열일 시
		if bi, ok := new(big.Int).SetString(t, 10); ok {
			return bi
		}
	}
	return nil //변환 실패 시 nil
} //JSON 수 표현을 *big.Int로 변환

func toStringSlice(v interface{}) []string {
	if v == nil {
		return nil
	}
	switch t := v.(type) {
	case []interface{}: //["a","b",...] 형태
		out := make([]string, 0, len(t))
		for _, e := range t {
			out = append(out, toString(e)) //각 원소를 문자열로 변환
		}
		return out
	case []string:
		return t
	case string:
		return []string{t}
	default:
		return []string{toString(t)}
	}
} //interface{}를 []string으로 변환

func toTime(v interface{}) time.Time {
	switch t := v.(type) {
	case string:
		if tm, err := time.Parse(time.RFC3339, t); err == nil { //RFC3339
			return tm
		}
		if bi, ok := new(big.Int).SetString(t, 10); ok { //epoch seconds 문자열
			return time.Unix(bi.Int64(), 0).UTC()
		}
	case float64: //JSON 숫자(epoch seconds)
		return time.Unix(int64(t), 0).UTC()
	case json.Number: //UseNumber 사용 시
		if i, err := t.Int64(); err == nil {
			return time.Unix(i, 0).UTC()
		}
	}
	return time.Time{} // 변환 실패 시 zero time
} //interface{} 값을 time.Time으로 변환

func strOrNil(x *big.Int) interface{} {
	if x == nil {
		return nil
	}
	return x.String()
} //*big.Int를 nil 또는 문자열로 변환
