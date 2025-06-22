BINARY_NAME=authService
GO=go

.PHONY: gen clean gen-all

PROTO_DIR=protos
GO_OUT_DIR=./protos
FILE_PROTO_DIR=fileClient/protos
FILE_GO_OUT_DIR=./fileClient/protos

gen:
	@echo "Generating Go code from auth proto files..."
	@protoc -I=$(PROTO_DIR) \
		--go_out=$(GO_OUT_DIR) \
		--go_opt=paths=source_relative \
		--go-grpc_out=$(GO_OUT_DIR) \
		--go-grpc_opt=paths=source_relative \
		$(PROTO_DIR)/*.proto
	@echo "Done!"

gen-file:
	@echo "Generating Go code from file service proto files..."
	@protoc -I=$(FILE_PROTO_DIR) \
		--go_out=$(FILE_GO_OUT_DIR) \
		--go_opt=paths=source_relative \
		--go-grpc_out=$(FILE_GO_OUT_DIR) \
		--go-grpc_opt=paths=source_relative \
		$(FILE_PROTO_DIR)/*.proto
	@echo "Done!"

gen-all: gen gen-file
	@echo "Generated all proto files!"

clean:
	@echo "Cleaning generated files..."
	@rm -f $(GO_OUT_DIR)/*.pb.go
	@rm -f $(FILE_GO_OUT_DIR)/*.pb.go
	@echo "Done!"

install-tools:
	@echo "Installing protoc-gen-go and protoc-gen-go-grpc..."
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@echo "Done!"

