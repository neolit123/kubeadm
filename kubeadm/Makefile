.PHONY: test

build:
	./hack/build.sh

test:
	go test ./app/...

test-integration:
	KUBE_ROOT=$(pwd) sudo -E env "PATH=${PATH}" go test ./test/...
