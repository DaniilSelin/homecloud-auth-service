BINARY_NAME=authService
GO=go

.PHONY: gen clean

PROTO_DIR=protos
GO_OUT_DIR=./protos

gen:
	@echo "Generating Go code from proto files..."
	@protoc -I=$(PROTO_DIR) \
		--go_out=$(GO_OUT_DIR) \
		--go_opt=paths=source_relative \
		--go-grpc_out=$(GO_OUT_DIR) \
		--go-grpc_opt=paths=source_relative \
		$(PROTO_DIR)/*.proto
	@echo "Done!"

clean:
	@echo "Cleaning generated files..."
	@rm -f $(GO_OUT_DIR)/*.pb.go
	@echo "Done!"

install-tools:
	@echo "Installing protoc-gen-go and protoc-gen-go-grpc..."
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@echo "Done!"

