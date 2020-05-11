.PHONY: pr

pr:
	go run cloudeng.io/go/cmd/gousage --overwrite ./...
	go run cloudeng.io/go/cmd/goannotate --config=copyright-annotation.yaml --annotation=cloudeng-copyright ./...
	go run cloudeng.io/go/cmd/gomarkdown --overwrite --circleci=cloudengio/go.gotools --goreportcard ./...
	echo > go.sum
