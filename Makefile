.PHONY: test

GO ?= go

test:
	@echo "+ $@"
	${GO} test $(shell go list ./... | grep -vE 'vendor')