# Используем официальный образ Go
FROM golang:1.22.5

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем go.mod и go.sum для кэширования зависимостей
COPY go.mod ./

# Загружаем зависимостиp
RUN go mod download

# Копируем остальной код проекта
COPY . .

# Выполняем сборку с помощью скрипта build.go
RUN go run build.go

# Указываем команду запуска для приложения
CMD ["./bin/amazing_website"]

# Открываем порт 8080 для доступа
EXPOSE 8080

#docker build -t web-server .
#docker run -d -p 8080:8080 web-server