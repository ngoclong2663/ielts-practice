run:
	go run ./cmd
build:
	CGO_ENABLED=0 go build -trimpath -o bin/api ./cmd
test:
	go test ./...
test-unit:
	go test -short ./...
lint:
	gofmt -l . && go vet ./...
gorm-gen:
	gorm gen -i ./app/models -o ./app/generated
newmg:
	migrate create -ext sql -dir migrations $(name)
up mgup:
	@bash scripts/migrate-up.sh $(version)
upto:
	@bash scripts/migrate-upto.sh $(version)
down mgdown:
	@bash scripts/migrate-down.sh $(version)
downto:
	@bash scripts/migrate-downto.sh $(version)

.PHONY: run build test test-unit lint gorm-gen newmg up mgup upto down mgdown downto
