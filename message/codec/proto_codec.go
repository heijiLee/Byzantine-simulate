// .proto, .pb.go, .desc(option) 필요
package codec

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"

	"codec/abstraction"
)

type ProtoDescriptorProvider interface {
	FindMessageByName(fullName protoreflect.FullName) (protoreflect.MessageDescriptor, error)
} //full name으로 메시지 descriptor 탐색

type globalRegistryProvider struct{} //global registry 조회

func (globalRegistryProvider) FindMessageByName(fn protoreflect.FullName) (protoreflect.MessageDescriptor, error) {
	mt, err := protoregistry.GlobalTypes.FindMessageByName(fn)
	if err != nil {
		return nil, err
	}
	return mt.Descriptor(), nil
} //global registry에서 메시지 타입 찾고 descriptor 반환

var DefaultDescriptorRegistry = NewDescriptorRegistry() //파일 registry 인스턴스

type compositeProvider struct {
	primary  ProtoDescriptorProvider //파일에서 로딩된 registry 등 우선 조회
	fallback ProtoDescriptorProvider //global registry 등 대체
} //primary -> fallback 순으로 제공

func (c compositeProvider) FindMessageByName(fn protoreflect.FullName) (protoreflect.MessageDescriptor, error) {
	if md, err := c.primary.FindMessageByName(fn); err == nil {
		return md, nil
	}
	return c.fallback.FindMessageByName(fn)
} //우선 primary에서 탐색 후 실패 시 fallback에서 탐색

type DescriptorRegistry struct {
	files *protoregistry.Files // 파일 단위 레지스트리(여러 .proto 집합)
} //파일 descriptor set

func NewDescriptorRegistry() *DescriptorRegistry {
	return &DescriptorRegistry{files: &protoregistry.Files{}}
} //파일 registry 생성

func (r *DescriptorRegistry) RegisterFile(fd protoreflect.FileDescriptor) error {
	return r.files.RegisterFile(fd)
} //파일 descriptor를 registry에 등록

func (r *DescriptorRegistry) RegisterFileDescriptorSet(fds *descriptorpb.FileDescriptorSet) error {
	files, err := protodesc.NewFiles(fds) // descriptorpb → protoreflect 변환
	if err != nil {
		return err
	}
	var regErr error
	files.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		if err := r.files.RegisterFile(fd); err != nil { //각 파일을 registry에 등록
			regErr = err
		}
		return true
	})
	_ = regErr
	return nil
} //.protoset/.desc 파일에 있는 FileDescriptorSet 등록

func (r *DescriptorRegistry) FindMessageByName(fullName protoreflect.FullName) (protoreflect.MessageDescriptor, error) {
	d, err := r.files.FindDescriptorByName(fullName) //이름으로 descriptor 조회
	if err == nil {
		if md, ok := d.(protoreflect.MessageDescriptor); ok {
			return md, nil
		}
		return nil, fmt.Errorf("descriptor %s found but not a message", fullName) //메시지가 아닌 경우
	}
	return nil, err //조회 실패
} //등록된 파일에서 full name으로 메시지 descriptor 탐색

func RegisterDescriptorSetFile(path string) error {
	blob, err := os.ReadFile(path) //파일 전체 읽기
	if err != nil {
		return err //파일 접근 실패
	}
	var fds descriptorpb.FileDescriptorSet              //파일 내용은 FileDescriptorSet 바이너리
	if err := proto.Unmarshal(blob, &fds); err != nil { //protobuf 바이너리 → fds
		return fmt.Errorf("descriptor set unmarshal: %w", err) //parsing 실패
	}
	return DefaultDescriptorRegistry.RegisterFileDescriptorSet(&fds) //registry에 등록
} //.protoset/.desc 파일 읽어 DefaultDescriptorRegistry에 등록

func init() {
	paths := os.Getenv("PROTO_DESC_FILES") //경로 목록
	if paths == "" {
		return //환경변수 미설정 시
	}
	for _, p := range strings.Split(paths, string(os.PathListSeparator)) { //경로 목록 순회
		p = strings.TrimSpace(p) //공백 제거
		if p == "" {
			continue
		}
		matches, _ := filepath.Glob(p)
		if len(matches) == 0 { //매칭 결과가 없을 시
			_ = RegisterDescriptorSetFile(p) //glob 없이 단일 경로 시도
			continue
		}
		for _, m := range matches { //매칭된 파일 각각 등록
			if err := RegisterDescriptorSetFile(m); err != nil {
				log.Printf("[proto] register desc failed: %s: %v\n", m, err) //실패
			} else {
				log.Printf("[proto] registered descriptor: %s\n", m) //성공
			}
		}
	}
} //환경변수 PROTO_DESC_FILES 읽어 자동으로 descriptor 등록

type protoCodec struct{} //protobuf 바이너리 <-> AbstractMessage 변환

func (protoCodec) providerFrom(opts ParseOptions) ProtoDescriptorProvider {
	if opts.DescriptorProvider != nil {
		return opts.DescriptorProvider
	}
	return compositeProvider{
		primary:  DefaultDescriptorRegistry,
		fallback: globalRegistryProvider{},
	} //local registry 우선, 실패 시 global registry
} //DescriptorProvider 우선 사용, 없을 시 DefaultDescriptorRegistry -> global registry 조회

func (pc protoCodec) Parse(data []byte, opts ParseOptions) (*abstraction.AbstractMessage, error) {
	provider := pc.providerFrom(opts)
	if opts.ProtoMessageFullName == "" {
		return nil, fmt.Errorf("protobuf parse requires ProtoMessageFullName")
	}
	md, err := provider.FindMessageByName(protoreflect.FullName(opts.ProtoMessageFullName)) //full name으로 메시지 descriptor 조회
	if err != nil {
		return nil, fmt.Errorf("descriptor not found for %s: %w", opts.ProtoMessageFullName, err)
	}
	msg := dynamicpb.NewMessage(md)
	if err := proto.Unmarshal(data, msg); err != nil {
		return nil, fmt.Errorf("protobuf unmarshal: %w", err)
	}
	js, err := protojson.MarshalOptions{
		UseEnumNumbers:  false, //enum을 문자열로 출력, true일 시 숫자 출력
		EmitUnpopulated: false, //빈 필드 무시
	}.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("protojson marshal: %w", err)
	}
	return (jsonCodec{}).Parse(js, ParseOptions{Format: FormatJSON, OverrideMsgType: opts.OverrideMsgType}) //jsonCodec으로 parsing하여 AbstractMessage로 정규화
} //protobuf 바이너리를 AbstractMessage로 변환

func (pc protoCodec) Serialize(am *abstraction.AbstractMessage, opts SerializeOptions) ([]byte, error) {
	provider := pc.providerFrom(ParseOptions{
		DescriptorProvider: opts.DescriptorProvider, //SerializeOptions에서 전달
	})
	if opts.ProtoMessageFullName == "" { //대상 protobuf 메시지 타입
		return nil, fmt.Errorf("protobuf serialize requires ProtoMessageFullName")
	}
	md, err := provider.FindMessageByName(protoreflect.FullName(opts.ProtoMessageFullName)) //대상 메시지 descriptor 조회
	if err != nil {
		return nil, fmt.Errorf("descriptor not found for %s: %w", opts.ProtoMessageFullName, err)
	}
	msg := dynamicpb.NewMessage(md)                                              //descriptor 기반 동적 메시지 생성
	js, err := (jsonCodec{}).Serialize(am, SerializeOptions{Format: FormatJSON}) //jsonCodec으로 AbstractMessage -> JSON 텍스트로 직렬화
	if err != nil {
		return nil, err
	}
	uopts := protojson.UnmarshalOptions{} //JSON 텍스트 -> protobuf 동적 메시지
	if opts.ProtoDiscardUnknown {
		uopts.DiscardUnknown = true //빈 필드 무시
	}
	if err := uopts.Unmarshal(js, msg); err != nil { //JSON → 메시지
		return nil, fmt.Errorf("protojson unmarshal to message(%s): %w", md.FullName(), err)
	}
	return proto.Marshal(msg)
} //AbstractMessage를 protobuf 바이너리로 변환
