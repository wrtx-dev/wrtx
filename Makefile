all: wrtx

version=$(shell cat VERSION)


wrtx: $(wildcard *.go)
	@go build -C cmds/wrtx -ldflags "-X main.WrtxVersion=${version}"
	@mv cmds/wrtx/wrtx ./

clean:
	@rm -rvf wrtx
	@rm -rvf build

modupdate:
	go mod tidy

deb: all
	@bash script/deb_package.sh ${version}
