# Usa la imagen base de Golang
FROM golang:1.23

# Configura el directorio de trabajo
WORKDIR /app

# Copia el código fuente al directorio de trabajo
COPY . .

RUN go mod download

# Compila el código fuente
RUN go build -o run-test


CMD ["./run-test"]
