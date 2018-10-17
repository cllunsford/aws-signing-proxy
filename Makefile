name=aws-signing-proxy
registry=cllunsford
gitrepo=github.com/cllunsford
tag=latest
go_ver=1.10

default:
	@echo ""
	@echo "make build:"
	@echo "	compiles the aws-signing-proxy app and builds the docker image"
	@echo "make gobuild:"
	@echo "	compiles the aws-signing-proxy app (binary located in ./_bin)"
	@echo "make dockbuild:"
	@echo "	builds the docker image"
	@echo "make clean:"
	@echo "	removes all temporary files and build artifacts"


build: gobuild dockbuild

dockbuild:
	[ -e ca-certificates.crt ] || wget https://curl.haxx.se/ca/cacert.pem -O ca-certificates.crt
	docker build -t ${registry}/${name}:${tag} .

gobuild:
	# copy src
	mkdir -p _src/${gitrepo}/${name}
	cp -r main.go _src/${gitrepo}/${name}
	# compile
	docker run \
	-v `pwd`/_pkg:/go/pkg \
	-v `pwd`/_bin:/go/bin \
	-v `pwd`/_src:/go/src \
	-e CGO_ENABLED=0 \
	-e GOOS=linux  \
	golang:${go_ver} \
	bash -c "go get ./src/${gitrepo}/${name}/...; chown -R $$(id -u):$$(id -g) ./"
	ln -f _bin/aws-signing-proxy

clean:
	rm -rf ./_* ca-certificates.crt aws-signing-proxy
