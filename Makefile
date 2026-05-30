GO_MODULE := github.com/leonardolimaArt/linkforge/LinkForge.Redirect
PROTO_OUT := LinkForge.Redirect

.PHONY: proto-gen proto-check proto-clean

proto-gen:
	protoc \
	  --proto_path=proto \
	  --go_out=$(PROTO_OUT) \
	  --go_opt=module=$(GO_MODULE) \
	  --go-grpc_out=$(PROTO_OUT) \
	  --go-grpc_opt=module=$(GO_MODULE) \
	  proto/linkforge/v1/link_service.proto
	@echo "✓ Proto generated em $(PROTO_OUT)/gen/"
proto-check: proto-gen
	@git diff --exit-code -- '$(PROTO_OUT)/gen/**/*.pb.go' || \
	  (echo "ERROR: generated protos are undated. Run 'make proto-gen' and commit." && exit 1)

proto-clean:
	rm -rf $(PROTO_OUT)/gen/