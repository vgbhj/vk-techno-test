# gRPC PubSub Service

Реализация микросервиса PubSub на Go с in‑memory брокером и gRPC.

## Описание

* Клиенты **Subscribe**: открывают стрим по ключу и получают события в порядке FIFO ([grpc.io](https://grpc.io/docs/languages/go/?utm_source=chatgpt.com)).
* Клиенты **Publish**: отправляют сообщение с `key` и `data`, которое доставляется всем подписчикам этого ключа.
* Логирование через **zap** для структурированных логов ([pkg.go.dev](https://pkg.go.dev/go.uber.org/zap?utm_source=chatgpt.com)).
* Корректный **graceful shutdown**: сигналы SIGINT/SIGTERM завершают приём новых запросов и дожидаются или отменяют текущие стримы.
* **Dependency Injection**: компоненты (конфиг, логгер, broker) передаются через конструкторы.

## Структура

```
grpc-pubsub/
├── cmd/server/main.go      # точка входа
├── config.yaml             # настройки сервиса
├── internal/config         # загрузка YAML
├── internal/log            # настройка zap
├── internal/server         # реализация gRPC
├── proto/pubsub.proto      # API definition
├── subpub                  # in-memory брокер
└── Dockerfile              # контейнеризация
```

## Установка и запуск

1. Собрать бинарь:

   ```bash
   go build -o bin/server ./cmd/server
   ```
2. Запустить:

   ```bash
   ./bin/server
   ```

## Docker

1. Построить образ:

   ```bash
   docker build -t grpc-pubsub .
   ```
2. Запустить контейнер:

   ```bash
   docker run -d -p 50051:50051 grpc-pubsub
   ```