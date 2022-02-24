js:
	@curl -L "https://unpkg.com/htmx.org@1.6.1" > http/assets/htmx.min.js

translate: generate
	@go generate ./internal/translations/translations.go

generate: http/html/*.ego
	@go run github.com/benbjohnson/ego/cmd/ego ./http/html

cleanrun: generate
	@rm -rf ./data
	@go run cmd/sirkulatord/sirkulatord.go --assets=$(CURDIR)/http/assets

run: generate
	@go run cmd/sirkulatord/sirkulatord.go --assets=$(CURDIR)/http/assets

datarun: generate
	@go run cmd/sirkulatord/sirkulatord.go --assets=$(CURDIR)/http/assets --db=_data

datareset:
	@rm _data/main.db
	@rm _data/files.db
	@rm -rf _data/index
