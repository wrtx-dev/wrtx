all: wrtx

wrtx: cmds/* package/*
	@go build -C cmds/wrtx
	@mv cmds/wrtx/wrtx ./
