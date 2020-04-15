#!/bin/bash
go generate ./...
echo > go.sum
go mod tidy
