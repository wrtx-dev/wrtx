all: wrtx

version=$(shell cat VERSION)

PACKAGES=$(shell find package -type f)
INTERNAL_PACKAGES=$(shell find internal -type f -name '*.go')
CMDS=$(shell find cmds -type f -name '*.go')

.PHONY: all clean modupdate deb

wrtx: $(CMDS) $(PACKAGES) $(INTERNAL_PACKAGES)
	@go build -C cmds/wrtx -ldflags "-X main.WrtxVersion=${version}"
	@mv cmds/wrtx/wrtx ./

clean:
	@rm -rvf wrtx
	@rm -rvf build

modupdate:
	go mod tidy

deb: all
	@bash script/deb_package.sh ${version}
