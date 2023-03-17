//go:generate oapi-codegen -package oapi -include-tags "morphs" -o ../oapi/morphs.gen.go ../openapi.yml
//go:generate stepci generate ../openapi.yml ../workflow.yml

package main

func main() {}
