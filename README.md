# 🚀 Aplicación Web en Go - Multi-tenant

Aplicación web funcional con sistema de login y **arquitectura multi-tenant**: cada usuario tiene su propia base de datos MySQL independiente, proporcionando aislamiento total de datos.

## 📋 Características

- ✅ Sistema de autenticación con login/logout
- ✅ Manejo de sesiones seguras
- ✅ **Base de datos independiente por usuario (Multi-tenant)**
- ✅ Dashboard con información dinámica en tiempo real
- ✅ Registro de actividades por usuario
- ✅ Diseño responsivo y moderno
- ✅ Fácil de desplegar

## 🏗️ Arquitectura Multi-tenant

```
Usuario 1 → Base de Datos 1 (webapp_admin)
Usuario 2 → Base de Datos 2 (webapp_usuario1)
Usuario 3 → Base de Datos 3 (webapp_demo)
```

Cada usuario tiene:
- Su propia base de datos MySQL
- Tablas de datos de usuario
- Log de actividades individual
- Datos de sesión independientes

## 🛠️ Requisitos

- Go 1.21 o superior
- MySQL 5.7+ o MariaDB 10.3+

## 🚀 Instalación y Ejecución

### 1. Instalar MySQL

#### Windows:
- Descargar desde [mysql.com/downloads](https://dev.mysql.com/downloads/mysql/)
- Ejecutar instalador y configurar password de root

#### Linux:
```bash
sudo apt update && sudo apt install mysql-server
```

#### Mac:
```bash
brew install mysql
brew services start mysql
```

### 2. Configurar la Base de Datos

**Opción A: Inicialización Automática (Recomendado)**

1. Editar credenciales en [main.go](main.go) (líneas 25-29):
```go
dbConfig = DBConfig{
    Host:     "localhost",
    Port:     "3306",
    RootUser: "root",
    RootPass: "TU_PASSWORD", // ⚠️ Cambiar esto
}
```

2. La aplicación creará automáticamente las bases de datos al ejecutarse.

**Opción B: Inicialización Manual**

```bash
# Ejecutar el script SQL
mysql -u root -p < init_databases.sql
```

### 3. Instalar Dependencias de Go

```bash
go mod download
```

### 4. Ejecutar la Aplicación

```bash
go run main.go
```

Deberías ver:
```
Inicializando bases de datos...
BD webapp_admin inicializada correctamente
BD webapp_usuario1 inicializada correctamente
BD webapp_demo inicializada correctamente
Conexión establecida a BD: webapp_admin
Servidor iniciado en http://localhost:8080
```

### 5. Abrir en el Navegador

```
http://localhost:8080
```

### Compilar para producción:

```bash
go build -o webapp
```

Esto genera un ejecutable llamado `webapp` (o `webapp.exe` en Windows) que puedes ejecutar directamente:

```bash
./webapp
```

## 👤 Usuarios de Prueba

Cada usuario tiene su propia base de datos:

| Usuario | Contraseña | Base de Datos |
|---------|-----------|---------------|
| admin | admin123 | webapp_admin |
| usuario1 | pass123 | webapp_usuario1 |
| demo | demo123 | webapp_demo |

**Cada base de datos contiene:**
- `user_data`: Información del usuario (última conexión, contador de logins, etc.)
- `activity_log`: Historial de acciones del usuario
- `session_data`: Datos de sesión

## 📦 Despliegue en Otros Servidores

⚠️ **IMPORTANTE**: El servidor de destino debe tener MySQL instalado y configurado.

### Opción 1: Copiar el ejecutable compilado

1. Compila la aplicación:
```bash
# Para Linux
GOOS=linux GOARCH=amd64 go build -o webapp-linux

# Para Windows
GOOS=windows GOARCH=amd64 go build -o webapp-windows.exe

# Para MacOS
GOOS=darwin GOARCH=amd64 go build -o webapp-mac
```

2. Copia el ejecutable y las carpetas necesarias al servidor:
```bash
scp webapp-linux user@servidor:/ruta/destino/
scp -r templates/ user@servidor:/ruta/destino/
scp -r static/ user@servidor:/ruta/destino/
scp init_databases.sql user@servidor:/ruta/destino/
```

3. En el servidor:
```bash
# Instalar MySQL si no está instalado
sudo apt install mysql-server

# Inicializar las bases de datos
mysql -u root -p < init_databases.sql

# Ejecutar la aplicación
./webapp-linux
```

### Opción 2: Clonar y ejecutar

1. En el servidor:
```bash
git clone <tu-repositorio>
cd <directorio>

# Instalar MySQL
sudo apt install mysql-server

# Inicializar bases de datos
mysql -u root -p < init_databases.sql

# Instalar dependencias de Go y ejecutar
go mod download
go run main.go
```

### Opción 3: Docker (incluye MySQL)

El Dockerfile incluye MySQL y la aplicación:

```bash
# Construir la imagen
docker build -t webapp-go .

# Ejecutar el contenedor
docker run -p 8080:8080 webapp-go
```

## 🆕 Agregar Nuevos Usuarios

1. Editar [add_new_user.sql](add_new_user.sql) y cambiar el nombre de usuario
2. Ejecutar: `mysql -u root -p < add_new_user.sql`
3. Agregar el usuario en [main.go](main.go) en el mapa `userConfig`
4. Reiniciar la aplicación

Ver [DESPLIEGUE.md](DESPLIEGUE.md) para instrucciones detalladas.

## 🗄️ Gestión de Base de Datos

### Backup de una base de datos específica:
```bash
mysqldump -u root -p webapp_admin > backup_admin.sql
```

### Restaurar:
```bash
mysql -u root -p webapp_admin < backup_admin.sql
```

### Ver todas las bases de datos:
```bash
mysql -u root -p -e "SHOW DATABASES LIKE 'webapp_%';"
```

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o webapp

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/webapp .
COPY templates/ templates/
COPY static/ static/
EXPOSE 8080
CMD ["./webapp"]
```

Construye y ejecuta:

```bash
docker build -t webapp-go .
docker run -p 8080:8080 webapp-go
```

## 🔧 Configuración

### Cambiar el puerto:

En [main.go](main.go#L33), modifica:
```go
log.Fatal(http.ListenAndServe(":8080", nil))
```

Por ejemplo, para usar el puerto 3000:
```go
log.Fatal(http.ListenAndServe(":3000", nil))
```

### Agregar más usuarios:

En [main.go](main.go#L17-L21), modifica el mapa `users`:
```go
users = map[string]string{
    "admin":    "admin123",
    "nuevo_usuario": "nueva_contraseña",
}
```

### Cambiar la clave secreta de sesiones:

En [main.go](main.go#L15), modifica:
```go
store = sessions.NewCookieStore([]byte("tu-clave-super-secreta"))
```

## 🌐 Acceso desde la red

Para permitir acceso desde otras máquinas en la red:

1. Cambia `localhost` por `0.0.0.0`:
```go
log.Fatal(http.ListenAndServe("0.0.0.0:8080", nil))
```

2. Si tienes firewall, abre el puerto:
```bash
# Linux (ufw)
sudo ufw allow 8080

# Windows (PowerShell como admin)
New-NetFirewallRule -DisplayName "WebApp Go" -Direction Inbound -LocalPort 8080 -Protocol TCP -Action Allow
```

3. Accede desde otra máquina usando la IP del servidor:
```
http://IP_DEL_SERVIDOR:8080
```

## 📁 Estructura del Proyecto

```
parcial/
├── main.go              # Servidor principal y handlers
├── go.mod               # Dependencias
├── templates/           # Plantillas HTML
│   ├── login.html
│   └── dashboard.html
├── static/              # Archivos estáticos
│   └── style.css
└── README.md
```

## 🔒 Seguridad

⚠️ **Nota importante:** Esta es una aplicación de demostración. Para producción:

- Usa una base de datos real para usuarios
- Implementa hash de contraseñas (bcrypt)
- Usa HTTPS/TLS
- Cambia la clave secreta de las sesiones
- Implementa rate limiting
- Agrega validación de entrada más robusta

## 📝 Licencia

Proyecto de demostración para fines educativos.
