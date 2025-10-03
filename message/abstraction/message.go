package abstraction

import (
	"math/big"
	"time"
)

type MsgType string //합의 프로토콜에서 사용하는 메시지 종류

const (
	MsgTypeProposal   MsgType = "Proposal"
	MsgTypePrepare    MsgType = "Prepare"
	MsgTypeVote       MsgType = "Vote" //HotStuff 등에서 Prepare + Commit의 의미
	MsgTypeCommit     MsgType = "Commit"
	MsgTypeViewChange MsgType = "ViewChange"
	MsgTypeNewView    MsgType = "NewView"
) //여러 구현체의 메시지 타입명을 표준값으로 정규화

type AbstractMessage struct {
	Type        MsgType           `json:"type"`                 //메시지 타입
	Height      *big.Int          `json:"height,omitempty"`     //블록 높이
	Round       *big.Int          `json:"round,omitempty"`      //라운드/epoch
	View        *big.Int          `json:"view,omitempty"`       //뷰 번호
	Timestamp   time.Time         `json:"timestamp,omitempty"`  //메시지 생성 시각
	BlockHash   string            `json:"block_hash,omitempty"` //제안 블록의 해시
	PrevHash    string            `json:"prev_hash,omitempty"`  //이전 블록 해시
	Proposer    string            `json:"proposer,omitempty"`   //제안자 ID
	Validator   string            `json:"validator,omitempty"`  //검증자 노드 ID
	Signature   string            `json:"signature,omitempty"`  //메시지 서명
	CommitSeals []string          `json:"commit_seals,omitempty"`
	ViewChanges []ViewChangeEntry `json:"view_changes,omitempty"`
	Extras      map[string][]byte `json:"extras,omitempty"`      //표준화되지 않은 필드
	RawPayload  []byte            `json:"raw_payload,omitempty"` //원본 메시지 바이트

	//아래 필드는 JSON serialization 시 제외됨
	OriginalFormat     string            `json:"-"` //최초 파싱된 포맷
	OriginalMsgName    string            `json:"-"` //원본 메시지 타입명
	OriginalFieldNames map[string]string `json:"-"` //원본 필드명 → 정규화된 필드명 매핑
} //여러 구현체의 field명을 synonyms로 정규화

type ViewChangeEntry struct {
	View      *big.Int `json:"view"`      //뷰 번호
	Height    *big.Int `json:"height"`    //해당 시점의 블록 높이
	Validator string   `json:"validator"` //검증자 ID
	Signature string   `json:"signature"` //검증자 서명
} //ViewChange 관련
