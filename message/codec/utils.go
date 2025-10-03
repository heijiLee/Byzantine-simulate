package codec

import (
	"bytes"
	"encoding/json"
	"strings"
)

func jsonFromInterface(v interface{}) ([]byte, error) {
	return json.Marshal(v)
} //go-bcs/rlp/msgpack 등 포맷 JSON 바이트로 직렬화(bcs/rlp/msgpack 등의 포맷)

func unmarshalJSON(data []byte, v interface{}) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	return dec.Decode(v)
} //JSON 바이트를 포인터로 역직렬화

func TrimBrackets(s string) string {
	if i := strings.Index(s, "("); i >= 0 { //첫 '(' 위치
		if j := strings.LastIndex(s, ")"); j > i { //마지막 ')' 위치
			return s[i+1 : j] //괄호 안 내용 반환
		}
	}
	return s //괄호가 없거나 형식이 다를 시 원문 반환
} //Name(k=v, ...) 형식의 문자열에서 괄호 내부 반환

func SplitKeyValuePairs(s string) map[string]string {
	m := make(map[string]string)
	if strings.TrimSpace(s) == "" { // 빈 문자열일 경우
		return m
	}
	for _, part := range strings.Split(s, ",") { //','로 분리
		if kv := strings.SplitN(part, "=", 2); len(kv) == 2 { // '=' 기준으로 분리
			key := strings.TrimSpace(kv[0]) //key 앞뒤 공백 제거
			val := strings.TrimSpace(kv[1]) //value 앞뒤 공백 제거
			m[key] = val
		}
	}
	return m
} //k=v, k=v, ... 형식 parsing
