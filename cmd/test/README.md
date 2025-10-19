# CometBFT ↔ Canonical 변환 테스트

이 디렉토리는 CometBFT Vote 메시지와 Canonical Message 간의 변환이 올바르게 작동하는지 테스트하는 코드를 포함합니다.

## 테스트 파일

### `vote_conversion_test.go`
- `TestVoteConversionFromJSON`: Vote.json 파일의 모든 예제를 테스트
- `TestSingleVoteConversion`: 단일 Vote 예제를 상세하게 테스트
- `TestVoteConversionBasic`: 기본 변환 기능을 간단히 테스트

## 테스트 실행 방법

### 1. 전체 테스트 실행
```bash
cd /Users/heiji/Develop/HeijiLee/Byzantine-simulate
go test ./cmd/test -v
```

### 2. 특정 테스트만 실행
```bash
# Vote.json 파일 테스트
go test ./cmd/test -run TestVoteConversionFromJSON -v

# 단일 Vote 테스트
go test ./cmd/test -run TestSingleVoteConversion -v

# 기본 변환 테스트
go test ./cmd/test -run TestVoteConversionBasic -v
```

### 3. 상세 출력과 함께 실행
```bash
go test ./cmd/test -v -count=1
```

## 테스트 내용

### TestVoteConversionFromJSON
- Vote.json의 6가지 Vote 예제 모두 테스트:
  - prevote_for_block
  - prevote_nil
  - precommit_basic
  - precommit_with_extension
  - precommit_nil
  - prevote_round_1

### TestSingleVoteConversion
- prevote_for_block 예제를 상세하게 테스트
- RawCometBFT → Canonical → RawCometBFT 순환 변환
- 각 단계별 결과 출력 및 비교

### TestVoteConversionBasic
- 기본적인 변환 기능 검증
- Height, Type, Validator 필드 정확성 확인

## 예상 결과

성공적인 테스트 실행 시:
```
=== RUN   TestVoteConversionFromJSON
🧪 Vote.json → Canonical → Vote.json 변환 테스트
===============================================
✅ Vote.json 파일 읽기 완료

📦 테스트 1: Prevote for Block
----------------------------------------
   🔄 RawCometBFT → Canonical 변환 중...
   🔄 Canonical → RawCometBFT 변환 중...
   🔍 원본과 변환된 메시지 비교 중...
   📊 변환 요약:
      Type: prevote
      Height: 1000
      Round: 0
      BlockHash: 7B1C3F5E8D9A2E4F6C8B0A1D3E5F7A9B2C4D6E8F0A1B3C5D7E9F1A3B5C7D9E0F
      Validator: 95CEC8D3BCD896B97A9195BCC9FC3F5A7C65E092
      Extensions: 2개
✅ 변환 성공!

📊 전체 결과: 6/6 성공 (100.0%)
--- PASS: TestVoteConversionFromJSON (0.01s)
```

## 문제 해결

### Go 모듈 문제 시:
```bash
go mod tidy
go mod download
```

### 파일 경로 문제 시:
- Vote.json 파일이 `examples/cometbft/Vote.json`에 있는지 확인
- 테스트 실행 위치가 프로젝트 루트인지 확인

### 컴파일 오류 시:
- import 경로 확인
- Go 버전 확인 (1.21 이상 권장)
