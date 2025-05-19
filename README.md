# httprouterext

A library for authorizing using nio.

# Rebuild protobuf code

    protoc -I $(pwd)/proto --go_out=$(pwd)/proto $(pwd)/proto/iam.proto --go-grpc_out=$(pwd)/proto
