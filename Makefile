# check if required binaries exists
EXECUTABLES = git go dep protoc
_ := $(foreach exec,$(EXECUTABLES),\
        $(if $(shell which $(exec)),some string,$(error "No $(exec) in PATH)))

# namespace-configuration should be first in the list
INFRA_MODULES := infra/namespace-configuration $(shell find infra -mindepth 1 -maxdepth 1 -type d | grep -v 'namespace-configuration')
JS_MODULES := apps/web-ui
GO_MODULES := $(shell find services -mindepth 1 -maxdepth 1 -type d)

BUILD_JS_TARGETS := $(foreach m,$(JS_MODULES),build-js/$(m))
BUILD_GO_TARGETS := $(foreach m,$(GO_MODULES),build-go/$(m))
DEPLOY_TARGETS := $(foreach m,$(INFRA_MODULES) $(JS_MODULES) $(GO_MODULES),deploy/$(m))

build: $(BUILD_JS_TASKS) $(BUILD_GO_TASKS)

setup-go:
	dep ensure
	go get -u github.com/golang/protobuf/protoc-gen-go
	protoc --go_out=paths=source_relative:. etc/schema/*.proto

build-go/%: setup-go
	@echo "Building go [$*]"
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
	 go build \
	 -o $*/build/app ./$*

build-js/%:
	@echo "Building js [$*]"
	cd $* && npm install && npm run build

build: $(BUILD_JS_TARGETS) $(BUILD_GO_TARGETS)

deploy: $(DEPLOY_TARGETS)

deploy/%:
	@echo "Deploying [$*]"
	$(call pull_images_for_cache,$*)
	cd $* && skaffold run
	$(call tag_n_push,$*)

##
## Get `cacheFrom` values from skafold and pull corresponding images
## Input params:
## $(1) - path to skaffold.yaml
##
define pull_images_for_cache
	$(eval IMAGE_NAMES := `yq r $(1)/skaffold.yaml -j | jq -re '.build.artifacts[]?.docker.cacheFrom | .[]?'`)
	for img in $(IMAGE_NAMES); do\
		docker pull $$img || true;\
    done
endef

##
## Get the newest image id by name from skaffold.yaml, tag it with "latest" and push it
## Input params:
## $(1) - path to skaffold.yaml
##
define tag_n_push
	$(eval IMAGE_NAMES := `yq r $(1)/skaffold.yaml -j | jq -re '.build.artifacts[]?.imageName'`)
	for img in $(IMAGE_NAMES); do\
		docker tag $$(docker images -q $$img | head -n 1) $$img:latest;\
		docker push $$img:latest;\
	done
endef
