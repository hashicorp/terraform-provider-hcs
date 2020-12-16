default: testacc
GOOS=$(shell go env GOOS)
GOARCH=$(shell go env GOARCH)
INSTALL_PATH=~/.local/share/terraform/plugins/localhost/providers/hcs/0.0.1/linux_$(GOARCH)
BUILD_ALL_PATH=${PWD}/bin

ifeq ($(GOOS), darwin)
	INSTALL_PATH=~/Library/Application\ Support/io.terraform/plugins/localhost/providers/hcs/0.0.1/darwin_$(GOARCH)
endif
ifeq ($(GOOS), "windows")
	INSTALL_PATH=%APPDATA%/HashiCorp/Terraform/plugins/localhost/providers/hcs/0.0.1/windows_$(GOARCH)
endif

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

# Get the latest hcs-ama-api-spec swagger.json
get-latest-hcs-ama-api-spec-swagger:
	curl https://raw.githubusercontent.com/hashicorp/cloud-consul-ama-api-spec/master/hcs/swagger.json --output ./internal/client/hcs-ama-api-spec/swagger/hcs-swagger.json

# Generate Go models from the latest hcs-ama-api-spec swagger.json
generate-hcs-ama-api-spec-models: get-latest-hcs-ama-api-spec-swagger
	swagger generate model \
		--spec=./internal/client/hcs-ama-api-spec/swagger/hcs-swagger.json \
		--target=./internal/client/hcs-ama-api-spec \
		--skip-validation

dev:
	mkdir -p $(INSTALL_PATH)
	go build -o $(INSTALL_PATH)/terraform-provider-hcs main.go

all:
	mkdir -p $(BUILD_ALL_PATH)
	GOOS=darwin go build -o $(BUILD_ALL_PATH)/terraform-provider-hcs_darwin-amd64 main.go
	GOOS=windows go build -o $(BUILD_ALL_PATH)/terraform-provider-hcs_windows-amd64 main.go
	GOOS=linux go build -o $(BUILD_ALL_PATH)/terraform-provider-hcs_linux-amd64 main.go
