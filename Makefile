# namespace-configuration should be first in the list
INFRA_MODULES := infra/namespace-configuration $(shell find -mindepth 1 -maxdepth 1 -type d infra | grep -v 'namespace-configuration')
SERVICE_MODULES := $(shell find -mindepth 1 -maxdepth 1 -type d services)
MODULES := $(INFRA_MODULES) $(SERVICE_MODULES)

BUILD_TASKS := $(foreach m,$(MODULES),build/$(m))
DEPLOY_TASKS := $(foreach m,$(MODULES),deploy/$(m))

build: $(BUILD_TASKS)

build/%:
	@echo "Building [$*]"
	$(call pull_images_for_cache,$*)
	cd $* && skaffold build
	$(call tag_n_push,$*)

deploy: $(DEPLOY_TASKS)

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
