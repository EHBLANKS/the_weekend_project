package proto

//go:generate protoc -I. -I../../api/proto --go_out=paths=source_relative:. --go-grpc_out=paths=source_relative,require_unimplemented_servers=false:. sports/sports.proto
