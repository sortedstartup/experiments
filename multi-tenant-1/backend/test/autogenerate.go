package main

// TODO: make a plugin for vscode
// This generates the go code from .proto files
// For simple cases like this it avoids the need to have a Makefile

//go:generate protoc --go_opt=module=sortedstartup/multi-tenant/test/proto --go-grpc_opt=module=sortedstartup/multi-tenant/test/proto --go_out=./proto/ --go-grpc_out=./proto/ --proto_path=../../proto multi-tenant.proto

// This generates JS code from .proto files
//go:generate protoc --ts_opt=no_namespace --ts_opt=unary_rpc_promise=true --ts_opt=target=web --ts_out=../../frontend/sorted-test/proto/ --proto_path=../../proto multi-tenant.proto

// The following lines require a Unix shell and will not work on Windows by default.
// Please run the equivalent PowerShell commands manually if needed.
// //go:generate sh -c "sed -i  's|@grpc/grpc-js|grpc-web|g' ../../frontend/sorted-test/proto/test.ts"
// //go:generate sh -c "sed -i '1i\\// @ts-nocheck' ../../frontend/sorted-test/proto/test.ts"

// // go:generate sqlc -f db/scripts/sqlc.yaml generate
