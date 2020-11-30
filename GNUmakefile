default: testacc

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
