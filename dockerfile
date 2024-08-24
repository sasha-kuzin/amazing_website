# Используем официальный образ Go
FROM golang:1.22.5

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем go.mod и go.sum для кэширования зависимостей
COPY go.mod ./

# Загружаем зависимости
RUN go mod download

# Копируем остальной код проекта
COPY . .

# Собираем бинарный файл в папку /app/bin
RUN go build -o bin/amazing_website ./cmd/amazing_website

# Указываем команду запуска
CMD ["./bin/amazing_website"]

# Открываем порт 8080 для доступа
EXPOSE 8080

#docker build -t web-server . && docker run -d -p 8080:8080 web-server