install:
	cd internal/code_gen/cmd/micro && go install
	cd internal/code_gen/cmd/protoc-gen-micro && go install