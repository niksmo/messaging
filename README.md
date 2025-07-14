# Messaging

Проект для отработки навыков по работе с Kafka Streams с использованием Goka.

Компоненты системы:

- Server - HTTP сервер для приема и отправки сообщений
- Block_user - утилита для блокировки/разблокировки пользователей
- Censor_word - утилита для добавления слов заменителей с целью цензуры контента сообщений
- Processor - набор процессоров для обработки сообщений пользователей

## Запуск проекта

1. Клонируйте репозиторий, установите зависимости:

```
git clone git@github.com:niksmo/messaging.git
cd messaging
go mod download
```

2. Запустите кластер Kafka используя `docker compose`:

```
docker compose up -d
```

3. Скомпилируйте Go-приложения:

```
mkdir bin && \
go build -o ./bin/server ./cmd/server/. & \
go build -o ./bin/processor ./cmd/processor/. & \
go build -o ./bin/block_user ./cmd/block_user/. & \
go build -o ./bin/censor_word ./cmd/censor_word/. & \
wait
```

4. Запустите `processor` и `server` в отдельный терминалах, первым должен быть запущен `processor`:

Первый терминал
```
./bin/processor
```
Подождите пока не увидите сообщение `processors are running`

Второй терминал
```
./bin/server
```
Сервер запустится на адресе `http://127.0.0.1:8000`

## Проверка функционала

### 1. Отправка сообщения

- Отправьте сообщение, например Джеку от Дэвида:

```
curl --data '{"To": "Jack", "Content": "Hi! My name is David."}' http://127.0.0.1:8000/David
```

- Прочитайте сообщения Джека:

```
curl http://127.0.0.1:8000/Jack
```

### 2. Блокировка пользователя

- Заблокируйте пользователя с помощью утилиты:

```
./bin/block_user -name Kevin -blocked
```

- Отправьте сообщение от Кевина Дэвиду:

```
curl --data '{"To": "David", "Content": "Hi! I am Kevin."}' http://127.0.0.1:8000/Kevin
```

- Прочитайте сообщения Дэвида, убедитесь что список сообщений пуст, добавьте параметр `-i` чтобы увидеть 204 HTTP-стaтус:

```
curl -i http://127.0.0.1:8000/David
```

### 3. Цензура контента сообщений

- Добавьте замену для слова `apple`, например на `orange` с помощью утилиты:

```
./bin/censor_word -word apple -change orange
```

- Отправьте сообщение со словом `apple` от Дэвида Джеку:

```
curl --data '{"To": "Jack", "Content": "I like green apple"}' http://127.0.0.1:8000/David
```

- Прочитайте сообщения Джэка, убедитесь что Дэвиду нравятся зеленые апельсины:

```
curl http://127.0.0.1:8000/Jack
```
