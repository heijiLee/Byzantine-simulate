# Byzantine Message Bridge Demo Makefile

.PHONY: demo build clean help

# 기본 타겟
all: demo

# 데모 실행
demo:
	@echo "🎮 CometBFT Byzantine Message Bridge 데모 실행 중..."
	@go run cmd/demo/*.go

# 빌드 (바이너리 생성)
build:
	@echo "🔨 바이너리 빌드 중..."
	@go build -o bin/demo cmd/demo/*.go
	@echo "✅ 빌드 완료: bin/demo"

# 빌드된 바이너리로 데모 실행
run: build
	@echo "🚀 빌드된 바이너리로 데모 실행 중..."
	@./bin/demo

# 정리
clean:
	@echo "🧹 정리 중..."
	@rm -rf bin/
	@echo "✅ 정리 완료"

# 테스트 실행
test:
	@echo "🧪 테스트 실행 중..."
	@go test ./...

# 의존성 설치
deps:
	@echo "📦 의존성 설치 중..."
	@go mod download
	@go mod tidy

# 코드 포맷팅
fmt:
	@echo "🎨 코드 포맷팅 중..."
	@go fmt ./...

# 린트 검사
lint:
	@echo "🔍 린트 검사 중..."
	@go vet ./...

# 도움말
help:
	@echo "📋 사용 가능한 명령어:"
	@echo "  make demo     - 데모 실행 (go run cmd/demo/*.go)"
	@echo "  make build    - 바이너리 빌드"
	@echo "  make run      - 빌드된 바이너리로 데모 실행"
	@echo "  make clean    - 빌드 파일 정리"
	@echo "  make test     - 테스트 실행"
	@echo "  make deps     - 의존성 설치"
	@echo "  make fmt      - 코드 포맷팅"
	@echo "  make lint     - 린트 검사"
	@echo "  make help     - 이 도움말 표시"
