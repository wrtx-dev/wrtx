all: wrtx wrtxproxy

version=$(shell cat VERSION)

wrtxproxy:
	@GOARCH=386 go build -C cmds/wrtxproxy -ldflags '-extldflags "-static"'
	@mv cmds/wrtxproxy/wrtxproxy ./
wrtx: $(wildcard *.go)
	@go build -C cmds/wrtx -ldflags "-X main.WrtxVersion=${version}"
	@mv cmds/wrtx/wrtx ./

clean:
	@rm -rvf wrtx wrtxproxy
	@rm -rvf build

modupdate:
	go mod tidy

deb: all
	@bash script/deb_package.sh ${version}
