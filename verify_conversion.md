# CometBFT ↔ Canonical 변환 검증 방법

## 1. 코드 분석을 통한 검증

### RawCometBFT → Canonical 변환 과정

**입력 (Vote.json 형식):**
```json
{
  "type": 1,
  "height": "1000", 
  "round": "0",
  "block_id": {
    "hash": "7B1C3F5E8D9A2E4F6C8B0A1D3E5F7A9B2C4D6E8F0A1B3C5D7E9F1A3B5C7D9E0F"
  },
  "validator_address": "95CEC8D3BCD896B97A9195BCC9FC3F5A7C65E092",
  "signature": "3045022100E1F23456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABC0220DE67890ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF012"
}
```

**변환 과정:**
1. `ToCanonical()` 함수 호출
2. JSON payload 파싱 → `CometBFTConsensusMessage` 구조체
3. Vote 타입 감지 (type: 1 → Prevote)
4. 문자열 → big.Int 변환 (height, round)
5. CanonicalMessage 생성

**출력 (Canonical 형식):**
```json
{
  "chain_id": "cosmos-hub-4",
  "height": 1000,
  "round": 0,
  "type": "prevote",
  "block_hash": "7B1C3F5E8D9A2E4F6C8B0A1D3E5F7A9B2C4D6E8F0A1B3C5D7E9F1A3B5C7D9E0F",
  "validator": "95CEC8D3BCD896B97A9195BCC9FC3F5A7C65E092",
  "signature": "3045022100E1F23456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABC0220DE67890ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF012",
  "extensions": {
    "vote_type": "PrevoteType",
    "validator_index": 0
  }
}
```

### Canonical → RawCometBFT 변환 과정

**입력:** 위의 Canonical 메시지

**변환 과정:**
1. `FromCanonical()` 함수 호출
2. Canonical Type 감지 (prevote → MsgTypePrevote)
3. big.Int → 문자열 변환 (height, round)
4. Vote 타입 설정 (type: 1)
5. CometBFTConsensusMessage 생성
6. JSON 직렬화

**출력:** 원본과 동일한 RawCometBFT 메시지

## 2. 실제 테스트 방법

### 방법 1: 단위 테스트 실행
```bash
cd /Users/heiji/Develop/HeijiLee/Byzantine-simulate
go test ./cometbft/adapter -v
```

### 방법 2: 간단한 테스트 스크립트 실행
```bash
go run test_conversion_simple.go
```

### 방법 3: 수동 검증
```bash
go run test_conversion_manual.go
```

## 3. 검증 체크리스트

### ✅ RawCometBFT → Canonical 변환
- [x] Vote 타입 올바르게 매핑 (1 → prevote, 2 → precommit)
- [x] Height/Round 문자열 → big.Int 변환
- [x] BlockHash 올바르게 복사
- [x] Validator 주소 올바르게 복사
- [x] Signature 올바르게 복사
- [x] Extensions에 추가 정보 저장

### ✅ Canonical → RawCometBFT 변환
- [x] Canonical Type → Vote 타입 매핑
- [x] big.Int → 문자열 변환
- [x] Vote 타입 설정 (1 또는 2)
- [x] 모든 필드 올바르게 복원
- [x] JSON 형식으로 직렬화

### ✅ 데이터 무결성
- [x] 원본 데이터 완전 보존
- [x] 타입 정보 올바르게 유지
- [x] Extension 데이터 처리
- [x] 메타데이터 보존

## 4. 예상 결과

변환이 올바르게 작동한다면:
- RawCometBFT → Canonical → RawCometBFT 변환 후 원본과 100% 동일
- 모든 필드가 올바르게 매핑됨
- 타입 정보가 정확히 보존됨
- Extension 데이터가 완전히 처리됨

## 5. 문제 해결

만약 테스트가 실패한다면:
1. Go 모듈 의존성 확인: `go mod tidy`
2. 컴파일 오류 확인: `go build ./cometbft/adapter`
3. 타입 변환 오류 확인: 문자열 ↔ big.Int 변환
4. JSON 직렬화 오류 확인: 구조체 태그

## 6. 성공 기준

- ✅ 모든 단위 테스트 통과
- ✅ 원본 데이터 100% 복원
- ✅ 타입 정보 정확한 매핑
- ✅ Extension 데이터 완전 처리
- ✅ 성능 오버헤드 최소화
