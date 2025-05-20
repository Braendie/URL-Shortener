MAIN_FILE=cmd/url-shortener/main.go

BINARY_NAME=url-shortener

CONFIG_PATH="config/local.yaml"

run:
	go run $(MAIN_FILE) -config=$(CONFIG_PATH)

build:
	go build -o $(BINARY_NAME) $(MAIN_FILE)

clean:
	rm -f $(BINARY_NAME)

test:
	go test ./...

help:
	@echo "Использование:"
	@echo "  make run       - Запуск приложения"
	@echo "  make build     - Сборка бинарного файла"
	@echo "  make clean     - Удаление собранных файлов"
	@echo "  make test      - Запуск тестов"