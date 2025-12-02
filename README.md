libs

go get github.com/go-playground/validator/v10 Валидатор

go get github.com/golang-jwt/jwt/v5 токены

go get golang.org/x/crypto@latest

go get -u github.com/knadh/koanf/v2

go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest

migrate -path ./migrations -database "sqlite3://file:./data/auth.db?_foreign_keys=1" up
migrate create -ext sql -dir ./migrations create_users
# создаст 000001_create_users.up.sql и 000001_create_users.down.sql
