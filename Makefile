IMAGE_NAME := training-demo
IMAGE_TAG := 0.0.1
CONTAINER_NAME := test

SERVICE_PORT := 9090
LISTEN_PORT := 9090
COMPOSE_HOST := 127.0.0.1

PROTO_PATH = gen
PROTO_VAL_VER = v1.2.1
GRPC_GW_VER = v2.26.1
PROTOC_GEN_GO_VER = v1.35.1

docker-build:
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .

docker-run:
	docker run -d -p $(SERVICE_PORT):$(LISTEN_PORT) $(IMAGE_NAME):$(IMAGE_TAG)

docker-compose:
	IMAGE_NAME=$(IMAGE_NAME) IMAGE_TAG=$(IMAGE_TAG) SERVICE_PORT=$(SERVICE_PORT)\
 	LISTEN_PORT=$(LISTEN_PORT) COMPOSE_HOST=$(COMPOSE_HOST) docker compose up -d

proto-gen:
	@rm -rf ${PROTO_PATH}
	@mkdir ${PROTO_PATH}
	@protoc api/*.proto \
    			-I external \
    			-I api \
    			-I ${GOPATH}/pkg/mod/github.com/grpc-ecosystem/grpc-gateway/v2@${GRPC_GW_VER} \
    			-I${GOPATH}/pkg/mod/github.com/envoyproxy/protoc-gen-validate@${PROTO_VAL_VER} \
    			--go_out=${PROTO_PATH}  \
    			--go_opt=paths=source_relative \
    			--go-grpc_out=${PROTO_PATH}  \
    			--go-grpc_opt=paths=source_relative \
    			--validate_out=lang=go:${PROTO_PATH}/ \
    			--validate_opt=paths=source_relative \
    			--grpc-gateway_out=allow_delete_body=true:${PROTO_PATH} \
    			--grpc-gateway_opt paths=source_relative \
    			--openapiv2_out=logtostderr=true,allow_delete_body=true,json_names_for_fields=false,allow_merge=true,merge_file_name=training:${PROTO_PATH}