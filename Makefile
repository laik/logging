CRD_OPTIONS ?= "crd:trivialVersions=true"

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	curl -OL https://github.com/yametech/controller-tools/archive/v0.4.1.tar.gz && tar -zxvf v0.4.1.tar.gz && cd controller-tools-0.4.1 ;\
	cd ./cmd/controller-gen && go install && cd ../helpgen && go install && cd ../type-scaffold && go install ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

# Just install controller-gen tools set
install-tools: controller-gen
	@echo "install controller-gen done"

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

# Generate manifests e.g. CRD, RBAC etc.
manifests: generate controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=deploy/crds

install: manifests
	kubectl apply -f deploy/crds/

uninstall: manifests
	kubectl delete -f deploy/crds/

build:
	docker build -t yametech/logging-api-server:v1.0.0 .
	docker push yametech/logging-api-server:v1.0.0