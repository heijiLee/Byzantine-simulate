# CometBFT Demo

CometBFT Byzantine Message Bridge 데모 프로그램입니다.

## 실행 방법

```bash
# 프로젝트 루트에서 실행
go run cmd/demo/main.go
```

## 기능

1. **메시지 시뮬레이션**: CometBFT 합의 메시지를 실시간으로 생성하고 변환
2. **Vote 변환 테스트**: Vote.json 파일의 예제들을 사용한 변환 테스트
3. **설정 테스트**: 기본 타입들이 정상적으로 로드되는지 확인

## 파일 구조

- `main.go`: 통합된 CometBFT 데모 프로그램

## 예제 파일

`examples/cometbft/Vote.json` 파일이 있어야 Vote 변환 테스트가 정상 작동합니다.

## 출력 예시

```
🎮 CometBFT Byzantine Message Bridge 데모
=======================================

📋 사용 가능한 데모:
   1. 메시지 시뮬레이션
   2. Vote 변환 테스트
   3. WAL 파일 분석
   4. 로컬넷 설정
   5. 설정 테스트

🚀 CometBFT 메시지 시뮬레이션 시작...

🚀 CometBFT 실시간 메시지 시뮬레이션 시작
=====================================
⏱️  실행 시간: 10s

📨 메시지 #1: CometBFT proposal 메시지 생성
   🔄 Canonical: height=1000, type=proposal
   📤 변환 완료: proposal

📨 메시지 #2: CometBFT prevote 메시지 생성
   🔄 Canonical: height=1001, type=prevote
   📤 변환 완료: prevote

...

✅ 시뮬레이션 완료! 총 5개 메시지 처리

🧪 Vote 변환 테스트 실행...
🧪 Vote 변환 테스트
==================
✅ Vote.json 파일 읽기 완료

📦 테스트 1: Prevote for Block
----------------------------------------
   🔄 RawCometBFT → Canonical 변환 중...
   🔄 Canonical → RawCometBFT 변환 중...
   🔍 원본과 변환된 메시지 비교 중...
   📊 변환 요약:
      Type: prevote
      Height: 1000
      Round: 1
      BlockHash: 0x1234567890abcdef...
      Validator: validator1
      Extensions: 0개
✅ 변환 성공!

...

📊 전체 결과: 6/6 성공 (100.0%)
🎉 모든 Vote 변환 테스트 통과!

🔧 설정 테스트 실행...
🔧 Byzantine Message Bridge 설정 테스트
=====================================
✅ ChainTypeCometBFT: cometbft
✅ ChainTypeHyperledger: hyperledger
✅ ChainTypeKaia: kaia
✅ MsgTypeProposal: proposal
✅ MsgTypeVote: vote
✅ MsgTypeBlock: block

🎉 모든 기본 타입이 정상적으로 로드되었습니다!
```
