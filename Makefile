all: wrtx

version=$(shell cat VERSION)

PACKAGES=$(shell find package -type f -name '*.go')

wrtx: $(wildcard cmds/wrtx/*.go) $(PACKAGES)
	@go build -C cmds/wrtx -ldflags "-X main.WrtxVersion=${version}"
	@mv cmds/wrtx/wrtx ./

clean:
	@rm -rvf wrtx
	@rm -rvf build

modupdate:
	go mod tidy

deb: all
	@bash script/deb_package.sh ${version}
