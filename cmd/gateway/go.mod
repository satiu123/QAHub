module qahub/gateway

go 1.25.1

require (
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.27.2
	google.golang.org/grpc v1.75.1
	google.golang.org/protobuf v1.36.9
	qahub/api v0.0.0
)

require (
	golang.org/x/net v0.44.0 // indirect
	golang.org/x/sys v0.36.0 // indirect
	golang.org/x/text v0.29.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250818200422-3122310a409c // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250818200422-3122310a409c // indirect
)

replace qahub/api => ../../api
