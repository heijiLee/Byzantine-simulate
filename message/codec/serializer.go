package codec

import (
	"codec/abstraction"
	"fmt"
	"strings"
	"time"
)

func (genericCodec) Serialize(am *abstraction.AbstractMessage, _ SerializeOptions) ([]byte, error) {
	s, err := SerializeGeneric(am)
	if err != nil {
		return nil, err
	}
	return []byte(s), nil //문자열을 바이트로 변환하여 반환
} //AbstractMessage를 generic 문자열 포맷으로 serializing

func SerializeGeneric(am *abstraction.AbstractMessage) (string, error) {
	phase := string(am.Type)
	var parts []string    //"k=v" 항목
	if am.Height != nil { //값이 존재할 시
		parts = append(parts, fmt.Sprintf("height=%s", am.Height)) //big.Int를 10진수 출력
	}
	if am.Round != nil {
		parts = append(parts, fmt.Sprintf("round=%s", am.Round))
	}
	if am.View != nil {
		parts = append(parts, fmt.Sprintf("view=%s", am.View))
	}
	if am.BlockHash != "" {
		parts = append(parts, fmt.Sprintf("block_hash=%s", am.BlockHash))
	}
	if am.PrevHash != "" {
		parts = append(parts, fmt.Sprintf("prev_hash=%s", am.PrevHash))
	}
	if !am.Timestamp.IsZero() { //Zero 타임이 아닐 시 포함
		parts = append(parts, fmt.Sprintf("timestamp=%s", am.Timestamp.UTC().Format(time.RFC3339))) //RFC3339로 출력
	}
	if am.Proposer != "" {
		parts = append(parts, fmt.Sprintf("proposer=%s", am.Proposer))
	}
	if am.Validator != "" {
		parts = append(parts, fmt.Sprintf("validator=%s", am.Validator))
	}
	if am.Signature != "" {
		parts = append(parts, fmt.Sprintf("signature=%s", am.Signature))
	}
	if len(am.CommitSeals) > 0 { //배열은 ','로 연결
		parts = append(parts, fmt.Sprintf("commit_seals=%s", strings.Join(am.CommitSeals, ",")))
	}
	for k, v := range am.Extras { //Extras는 표준화되지 않은 key-value 쌍
		parts = append(parts, fmt.Sprintf("%s=%s", k, string(v))) //[]byte 값을 문자열로 변환
	}
	return fmt.Sprintf("%s(%s)", phase, strings.Join(parts, ",")), nil
} //Phase(k=v,...) 형태의 문자열 생성
