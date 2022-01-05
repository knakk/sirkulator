translate: generate
	@go generate ./internal/translations/translations.go

generate: http/html/*.ego
	@go run github.com/benbjohnson/ego/cmd/ego ./http/html

run: generate
	@go run cmd/sirkulatord/sirkulatord.go --assets=$(CURDIR)/http/assets
