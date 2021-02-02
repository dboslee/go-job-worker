PROTODIR ?= pkg/grpc/proto
OS ?= $(shell go env GOOS)
ARCH ?= $(shell go env GOARCH)
SERVER_IP ?= 0.0.0.0
CLIENT_IP ?= 0.0.0.0

.PHONY: build build-server build-client clean test clean-output proto certs ca-cert server-cert client-cert certs-dir

build: build-server build-client

build-server: vendor
	GOOS=$(OS) GOARCH=$(ARCH) go build -mod vendor -o server cmd/server/server.go

build-client: vendor
	GOOS=$(OS) GOARCH=$(ARCH) go build -mod vendor -o client cmd/client/client.go

clean: clean-output
	@rm -f server
	@rm -f client
	@rm -rf vendor
	@rm -rf certs

clean-output:
	@rm -rf /tmp/job-worker-output-*

test: vendor
	go test ./...

vendor:
	@go mod vendor

proto:
	@cd $(PROTODIR); \
	protoc --go_out=plugins=grpc:. *.proto

certs: ca-cert server-cert client-cert

ca-cert: certs-dir
	cd certs;\
	openssl req -x509 -newkey rsa:2048 -nodes -keyout ca.key -out ca.pem -subj "/C=US/ST=RI/L=Providence/CN=CA";\
	openssl x509 -in ca.pem -noout -text
	
server-cert: certs-dir server.cnf
	cd certs;\
	openssl req -newkey rsa:2048 -nodes -keyout server.key -out server.crs -subj "/CN=server";\
	openssl x509 -req -in server.crs -CA ca.pem -CAkey ca.key -CAcreateserial -out server.pem -extfile server.cnf

client-cert: certs-dir client.cnf
	cd certs;\
	openssl req -newkey rsa:2048 -nodes -keyout client1.key -out client1.crs -subj "/CN=client1";\
	openssl x509 -req -in client1.crs -CA ca.pem -CAkey ca.key -CAcreateserial -out client1.pem -extfile client.cnf

server.cnf: certs-dir
	echo subjectAltName = IP:$(SERVER_IP)> certs/server.cnf

client.cnf: certs-dir
	echo subjectAltName = IP:$(CLIENT_IP) > certs/client.cnf

certs-dir:
	@mkdir -p certs
