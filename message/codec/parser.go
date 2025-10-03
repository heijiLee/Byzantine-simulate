package codec

import (
	"codec/abstraction"
	"fmt"
	"math/big"
	"strings"
	"time"
)

type genericCodec struct{} //Proposal(height=..., ...) 형태의 문자열을 parsing/serializing

func (genericCodec) Parse(data []byte, _ ParseOptions) (*abstraction.AbstractMessage, error) {
	raw := strings.TrimSpace(string(data)) //입력 바이트를 문자열로 바꾸고 양끝 공백 제거
	if raw == "" {                         //빈 문자열일 시
		return nil, fmt.Errorf("empty raw") //에러 반환
	}
	idx := strings.Index(raw, "(") //메시지명과 내용 구분하는 '(' 위치 탐색
	if idx < 0 {                   // '(' 가 없을 시 포맷 오류
		return nil, fmt.Errorf("invalid message: no '(', raw=%s", raw) //에러 반환
	}
	msgName := strings.TrimSpace(raw[:idx]) //'(' 이전 구간을 메시지명으로 추출하고 공백 제거
	body := TrimBrackets(raw)               //괄호 안 내용만 잘라냄
	am := &abstraction.AbstractMessage{
		Extras:             map[string][]byte{},     //표준화되지 않은 필드는 binary로 보존
		RawPayload:         []byte(raw),             //입력 원문
		OriginalFormat:     string(FormatGeneric),   //현재 format: generic
		OriginalMsgName:    msgName,                 //원본 메시지명
		OriginalFieldNames: make(map[string]string), //원본 필드명 -> 표준 필드명 매핑
	} //AbstractMessage 초기화
	if t, ok := PhaseSynonyms[msgName]; ok { // 유의어 존재할 시
		am.Type = abstraction.MsgType(t) // 표준 타입명으로 설정
	} else { // 유의어 없을 시
		am.Type = abstraction.MsgType(msgName) // 원문 그대로 사용
		am.OriginalMsgName = msgName
	}
	kv := SplitKeyValuePairs(body) //key=value 쌍의 맵으로 parsing
	for k, v := range kv {
		if fld, ok := FieldSynonyms[k]; ok { //원본 필드명 -> 표준 필드명 정규화
			am.OriginalFieldNames[fld] = k //원본 필드명 기록
			switch fld {
			case "Height":
				if x, ok := new(big.Int).SetString(v, 10); ok { //10진수 문자열 -> big.Int 변환
					am.Height = x
				}
			case "Round":
				if x, ok := new(big.Int).SetString(v, 10); ok {
					am.Round = x
				}
			case "View":
				if x, ok := new(big.Int).SetString(v, 10); ok {
					am.View = x
				}
			case "BlockHash":
				am.BlockHash = v
			case "PrevHash":
				am.PrevHash = v
			case "Timestamp":
				if t2, err := time.Parse(time.RFC3339, v); err == nil { //RFC3339 시도
					am.Timestamp = t2
				} else if sec, ok := new(big.Int).SetString(v, 10); ok { //epoch seconds 시도
					am.Timestamp = time.Unix(sec.Int64(), 0).UTC() //초 단위 epoch -> UTC Time
				}
			case "Proposer":
				am.Proposer = v
			case "Validator":
				am.Validator = v
			case "Signature":
				am.Signature = v
			case "CommitSeals": //','로 구분된 문자열 리스트
				am.CommitSeals = strings.Split(v, ",")
			case "ViewChanges": //view:height:validator:signature 형식의 리스트
				am.ViewChanges = parseViewChanges(v)
			default:
				am.Extras[k] = []byte(v) //정의되지 않은 필드명
			}
		} else {
			am.Extras[k] = []byte(v) //유의어 존재하지 않는 필드
		}
	}
	return am, nil //parsing한 메시지 반환
} //generic 포맷의 바이트를 AbstractMessage로 변환

func parseViewChanges(raw string) []abstraction.ViewChangeEntry {
	var entries []abstraction.ViewChangeEntry
	items := strings.Split(raw, ",") //','로 분리
	for _, item := range items {
		parts := strings.Split(strings.TrimSpace(item), ":") //view:height:validator:signature 분해
		if len(parts) < 4 {                                  //필드 개수 부족할 시
			continue
		}
		view, _ := new(big.Int).SetString(parts[0], 10)   //view 10진수 -> big.Int (실패 시 nil)
		height, _ := new(big.Int).SetString(parts[1], 10) //height 10진수 -> big.Int
		validator := parts[2]                             //validator 문자열
		signature := parts[3]                             //signature 문자열
		entries = append(entries, abstraction.ViewChangeEntry{
			View:      view,
			Height:    height,
			Validator: validator,
			Signature: signature,
		}) //parsing된 값을 ViewChangeEntry에 추가
	}
	return entries
} //view:height:validator:signature 문자열을 필드 4개로 구성된 []ViewChangeEntry로 변환
