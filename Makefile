all: wrtx

version=$(shell cat VERSION)
src_file=$(vpath %.go ./)
wrtx: ${src_file}
	@echo ${src_file}
	@go build -C cmds/wrtx -ldflags "-X main.WrtxVersion=${version}"
	@mv cmds/wrtx/wrtx ./

clean:
	@rm -rvf wrtx
	@test -d build && rm -rvf build

modupdate:
	go mod tidy

deb: all
	@bash script/deb_package.sh ${version}
