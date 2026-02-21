FROM golang:1.21-alpine AS builder

# Establecer directorio de trabajo
WORKDIR /app

# Copiar archivos de dependencias
COPY go.mod go.sum* ./

# Descargar dependencias
RUN go mod download

# Copiar el código fuente
COPY . .

# Compilar la aplicación
RUN go build -o webapp main.go

# ===== Imagen final =====
FROM alpine:latest

# Instalar certificados SSL y cliente MySQL
RUN apk --no-cache add ca-certificates mysql-client

# Crear directorio de trabajo
WORKDIR /app

# Copiar el binario compilado
COPY --from=builder /app/webapp .

# Copiar templates y archivos estáticos
COPY templates/ templates/
COPY static/ static/
COPY init_databases.sql ./

# Exponer el puerto
EXPOSE 8080

# Script de inicio para esperar MySQL
RUN echo '#!/bin/sh' > /app/entrypoint.sh && \
    echo 'echo "Esperando MySQL..."' >> /app/entrypoint.sh && \
    echo 'sleep 10' >> /app/entrypoint.sh && \
    echo 'echo "Iniciando aplicación..."' >> /app/entrypoint.sh && \
    echo './webapp' >> /app/entrypoint.sh && \
    chmod +x /app/entrypoint.sh

# Ejecutar la aplicación
CMD ["/app/entrypoint.sh"]
