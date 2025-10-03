# codec
generic, JSON, Protobuf, BCS, RLP, MessagePack format supported

1. install dependencies: go mod tidy
2. generate protobuf descriptor set: protoc --proto_path=proto --descriptor_set_out=proto/abstraction.protoset --include_imports --include_source_info proto/abstraction.proto
3. run the encoding/decoding testapp: go run ./cmd/testapp
