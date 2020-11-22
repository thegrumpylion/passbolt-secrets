BIN ?= pbscontroller
IMAGE_TAG ?= passbolt-secrets-controller

${BIN}: cmd/main.go pkg/controller/controller.go pkg/apis/passboltsecrets/v1alpha1/zz_generated.deepcopy.go
	CGO_ENABLED=0 go build -tags netgo -o $@ ./cmd

pkg/apis/passboltsecrets/v1alpha1/zz_generated.deepcopy.go: pkg/apis/passboltsecrets/v1alpha1/types.go
	bash $(call modpath,k8s.io/code-generator)/generate-groups.sh all \
		github.com/thegrumpylion/passbolt-secrets/pkg/client \
		github.com/thegrumpylion/passbolt-secrets/pkg/apis \
		passboltsecrets:v1alpha1 \
		--output-base . \
		--go-header-file gen/boilerplate.go.txt
	cp -a github.com/thegrumpylion/passbolt-secrets/* .
	rm -rf github.com

image: ${BIN}
	docker build -t ${IMAGE_TAG} .

image_build:
	docker build -f Dockerfile.build -t ${IMAGE_TAG} .

deploy:
	kubectl delete pod -n passbolt-secrets passbolt-secrets-scontroller
	kubectl create -f artifacts/pbscontroller-pod.yaml

tools:
	go install -tags tools k8s.io/code-generator/cmd/...

clean:
	rm ${BIN}

clean_gen:
	rm -rf pkg/client pkg/apis/passboltsecrets/v1alpha1/zz_generated.deepcopy.go

define modpath
${GOPATH}/pkg/mod/$(1)@$(shell grep $(1) go.mod | head -n 1 | awk '{print $$2}')
endef