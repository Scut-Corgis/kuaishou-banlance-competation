rm ./*.go
protoc --proto_path=. --go-grpc_out=. --go_out=. B.proto
cp ./*.go ../A/pkg/api/apiB
cp ./*.go ../B/pkg/api/apiB
