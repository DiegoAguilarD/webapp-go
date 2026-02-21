# 📦 Guía Completa de Despliegue

Esta guía te explica paso a paso cómo desplegar tu aplicación web en Go para que otros usuarios puedan acceder a ella desde diferentes máquinas y servidores.

**🗄️ Arquitectura Multi-tenant:** Esta aplicación utiliza una **base de datos MySQL independiente por cada usuario**, proporcionando aislamiento total de datos entre usuarios.

---

## 📋 Índice

0. [**Configuración de Base de Datos (OBLIGATORIO)**](#0-configuración-de-base-de-datos-obligatorio)
1. [Despliegue en la Red Local (LAN)](#1-despliegue-en-la-red-local-lan)
2. [Despliegue en un Servidor Linux](#2-despliegue-en-un-servidor-linux)
3. [Despliegue en un Servidor Windows](#3-despliegue-en-un-servidor-windows)
4. [Despliegue con Docker](#4-despliegue-con-docker)
5. [Despliegue en la Nube](#5-despliegue-en-la-nube)
6. [Agregar Nuevos Usuarios](#6-agregar-nuevos-usuarios)
7. [Solución de Problemas](#7-solución-de-problemas)

---

## 0. Configuración de Base de Datos (OBLIGATORIO)

⚠️ **IMPORTANTE**: La aplicación requiere MySQL para funcionar. Cada usuario tiene su propia base de datos independiente.

### Paso 1: Instalar MySQL

#### En Windows:
1. Descargar MySQL Community Server desde [mysql.com/downloads](https://dev.mysql.com/downloads/mysql/)
2. Ejecutar el instalador y configurar password de root
3. Agregar MySQL a las variables de entorno PATH

#### En Linux:
```bash
# Ubuntu/Debian
sudo apt update
sudo apt install mysql-server
sudo mysql_secure_installation

# CentOS/RHEL
sudo yum install mysql-server
sudo systemctl start mysqld
sudo mysql_secure_installation
```

#### En Mac:
```bash
brew install mysql
brew services start mysql
mysql_secure_installation
```

### Paso 2: Inicializar las Bases de Datos

Hay dos opciones para inicializar las bases de datos:

#### Opción A: Inicialización Automática (Recomendado)

La aplicación crea automáticamente las bases de datos al iniciar. Solo asegúrate de:

1. **Configurar las credenciales de MySQL en [main.go](main.go) (líneas 25-29):**
```go
dbConfig = DBConfig{
    Host:     "localhost",
    Port:     "3306",
    RootUser: "root",
    RootPass: "TU_PASSWORD_MYSQL", // ⚠️ Cambiar esto
}
```

2. **Ejecutar la aplicación:**
```bash
go run main.go
```

La aplicación creará automáticamente:
- `webapp_admin` - Base de datos para usuario "admin"
- `webapp_usuario1` - Base de datos para usuario "usuario1"
- `webapp_demo` - Base de datos para usuario "demo"

#### Opción B: Inicialización Manual

Si prefieres inicializar manualmente o si la automática falla:

```bash
# Conectar a MySQL como root
mysql -u root -p

# Ejecutar el script de inicialización
source init_databases.sql
```

O desde PowerShell:
```powershell
Get-Content init_databases.sql | mysql -u root -p
```

### Paso 3: Instalar Dependencias de Go

```bash
go mod download
```

### Paso 4: Verificar la Conexión

Ejecuta la aplicación y verifica en los logs:
```bash
go run main.go
```

Deberías ver:
```
Inicializando bases de datos...
BD webapp_admin inicializada correctamente
BD webapp_usuario1 inicializada correctamente
BD webapp_demo inicializada correctamente
Servidor iniciado en http://localhost:8080
```

---

## 1. Despliegue en la Red Local (LAN)

Permite que otras computadoras en tu misma red accedan a la aplicación.

### Paso 1: Modificar el servidor para aceptar conexiones externas

Abre el archivo [main.go](main.go) y modifica la línea 33:

**Cambiar de:**
```go
log.Fatal(http.ListenAndServe(":8080", nil))
```

**A:**
```go
log.Fatal(http.ListenAndServe("0.0.0.0:8080", nil))
```

### Paso 2: Abrir el puerto en el firewall de Windows

Abre PowerShell como **Administrador** y ejecuta:

```powershell
New-NetFirewallRule -DisplayName "WebApp Go" -Direction Inbound -LocalPort 8080 -Protocol TCP -Action Allow
```

### Paso 3: Obtener tu IP local

En PowerShell:
```powershell
ipconfig
```

Busca tu **Dirección IPv4** (ejemplo: `192.168.1.100`)

### Paso 4: Ejecutar la aplicación

```powershell
go run main.go
```

### Paso 5: Acceso desde otras máquinas

Desde cualquier computadora en la misma red, abre un navegador:
```
http://TU_IP:8080
```

Ejemplo: `http://192.168.1.100:8080`

---

## 2. Despliegue en un Servidor Linux

### Opción A: Compilar y copiar el ejecutable

#### En tu máquina Windows:

1. **Compila para Linux:**
```powershell
$env:GOOS="linux"; $env:GOARCH="amd64"; go build -o webapp-linux main.go
```

2. **Copia los archivos al servidor:**
```powershell
# Usando SCP (necesitas tener instalado OpenSSH o usar WinSCP)
scp webapp-linux usuario@IP_SERVIDOR:/home/usuario/webapp/
scp -r templates/ usuario@IP_SERVIDOR:/home/usuario/webapp/
scp -r static/ usuario@IP_SERVIDOR:/home/usuario/webapp/
```

#### En el servidor Linux:

3. **Dale permisos de ejecución:**
```bash
cd /home/usuario/webapp
chmod +x webapp-linux
```

4. **Ejecuta la aplicación:**
```bash
./webapp-linux
```

### Opción B: Clonar el código en el servidor

#### En el servidor Linux:

1. **Instala Go:**
```bash
# Ubuntu/Debian
sudo apt update
sudo apt install golang-go

# CentOS/RHEL
sudo yum install golang
```

2. **Copia los archivos** (usando SCP, Git, o SFTP)

3. **Instala las dependencias:**
```bash
cd /ruta/a/webapp
go mod download
```

4. **Ejecuta:**
```bash
go run main.go
```

### Mantener la aplicación corriendo (systemd)

Crea un servicio para que la app se ejecute automáticamente:

1. **Crea el archivo de servicio:**
```bash
sudo nano /etc/systemd/system/webapp.service
```

2. **Contenido del archivo:**
```ini
[Unit]
Description=WebApp Go
After=network.target

[Service]
Type=simple
User=tu_usuario
WorkingDirectory=/home/usuario/webapp
ExecStart=/home/usuario/webapp/webapp-linux
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
```

3. **Activa y ejecuta el servicio:**
```bash
sudo systemctl daemon-reload
sudo systemctl enable webapp
sudo systemctl start webapp
sudo systemctl status webapp
```

4. **Abre el puerto en el firewall:**
```bash
# UFW (Ubuntu)
sudo ufw allow 8080

# firewalld (CentOS/RHEL)
sudo firewall-cmd --permanent --add-port=8080/tcp
sudo firewall-cmd --reload
```

---

## 3. Despliegue en un Servidor Windows

### Método 1: Ejecutar directamente

1. **Copia toda la carpeta** del proyecto al servidor Windows

2. **Instala Go** en el servidor (si no está instalado)

3. **Ejecuta:**
```powershell
cd C:\ruta\a\webapp
go run main.go
```

### Método 2: Compilar un ejecutable

1. **En tu máquina, compila:**
```powershell
go build -o webapp.exe main.go
```

2. **Copia estos archivos al servidor:**
   - `webapp.exe`
   - Carpeta `templates/`
   - Carpeta `static/`

3. **En el servidor, ejecuta:**
```powershell
.\webapp.exe
```

### Ejecutar como servicio de Windows

1. **Descarga NSSM** (Non-Sucking Service Manager): https://nssm.cc/download

2. **Instala el servicio:**
```powershell
nssm install WebAppGo "C:\ruta\a\webapp\webapp.exe"
nssm set WebAppGo AppDirectory "C:\ruta\a\webapp"
nssm start WebAppGo
```

3. **Abre el puerto en el firewall:**
```powershell
New-NetFirewallRule -DisplayName "WebApp Go" -Direction Inbound -LocalPort 8080 -Protocol TCP -Action Allow
```

---

## 4. Despliegue con Docker

Docker permite empaquetar la aplicación con MySQL incluido.

### Opción A: Docker Compose (Recomendado)

Incluye MySQL automáticamente.

#### Paso 1: Instalar Docker y Docker Compose

- **Windows/Mac:** [docker.com/products/docker-desktop](https://www.docker.com/products/docker-desktop)
- **Linux:**
  ```bash
  sudo apt install docker.io docker-compose
  ```

#### Paso 2: Configurar (Archivo ya incluido)

El archivo [docker-compose.yml](docker-compose.yml) ya está configurado y define:
- `mysql`: Contenedor MySQL 8.0
- `webapp`: Contenedor de la aplicación Go

#### Paso 3: Iniciar todo

```bash
# Construir e iniciar ambos contenedores
docker-compose up -d

# Ver los logs
docker-compose logs -f
```

La aplicación estará disponible en `http://localhost:8080`

#### Paso 4: Comandos útiles

```bash
# Ver estado
docker-compose ps

# Detener todo
docker-compose down

# Reiniciar
docker-compose restart

# Ver logs de un servicio específico
docker-compose logs webapp
docker-compose logs mysql

# Ejecutar comandos en MySQL
docker-compose exec mysql mysql -uroot -ppassword

# Ver bases de datos
docker-compose exec mysql mysql -uroot -ppassword -e "SHOW DATABASES LIKE 'webapp_%';"
```

#### Paso 5: Cambiar password de MySQL

Editar [docker-compose.yml](docker-compose.yml):
```yaml
environment:
  MYSQL_ROOT_PASSWORD: tu_nuevo_password
```

Y también actualizar en la sección `webapp`:
```yaml
environment:
  - DB_ROOT_PASSWORD=tu_nuevo_password
```

### Opción B: Docker solo (sin MySQL incluido)

Si ya tienes MySQL corriendo localmente:

#### Paso 1: Construir la imagen

```bash
docker build -t webapp-go:latest .
```

#### Paso 2: Ejecutar el contenedor

```bash
# Conectar al MySQL del host
docker run -d -p 8080:8080 \
  -e DB_HOST=host.docker.internal \
  -e DB_ROOT_PASSWORD=tu_password \
  --name webapp webapp-go:latest
```

### Compartir la imagen Docker

#### Opción 1: Guardar en archivo

```bash
# Guardar
docker save webapp-go:latest -o webapp-go.tar

# Copiar a otro servidor (por SCP, USB, etc.)

# En el otro servidor, cargar
docker load -i webapp-go.tar
```

#### Opción 2: Usar Docker Hub

```bash
# Login
docker login

# Etiquetar
docker tag webapp-go:latest tu_usuario/webapp-go:latest

# Subir
docker push tu_usuario/webapp-go:latest

# En otro servidor
docker pull tu_usuario/webapp-go:latest
```

### Desplegar con Docker Compose en servidor remoto

1. **Copiar archivos al servidor:**
```bash
scp docker-compose.yml usuario@servidor:/ruta/
scp Dockerfile usuario@servidor:/ruta/
scp -r templates/ static/ usuario@servidor:/ruta/
scp go.mod main.go init_databases.sql usuario@servidor:/ruta/
```

2. **En el servidor:**
```bash
cd /ruta/
docker-compose up -d
```

---

## 5. Despliegue en la Nube

### AWS EC2

1. **Crea una instancia EC2** (Ubuntu Server recomendado)

2. **Configura el Security Group** para permitir tráfico en el puerto 8080 y 3306 (MySQL)

3. **Conéctate por SSH:**
```bash
ssh -i tu-llave.pem ubuntu@IP_PUBLICA
```

4. **Instala Docker y Docker Compose:**
```bash
sudo apt update
sudo apt install docker.io docker-compose
sudo systemctl start docker
sudo usermod -aG docker ubuntu
```

5. **Copia los archivos** (como se mostró arriba)

6. **Ejecuta:**
```bash
docker-compose up -d
```

7. **Accede desde:**
```
http://IP_PUBLICA_EC2:8080
```

### Google Cloud Platform (GCP)

1. **Crea una VM en Compute Engine**

2. **Configura las reglas de firewall** para el puerto 8080

3. **SSH a la VM y despliega** siguiendo los pasos de Linux

### DigitalOcean

1. **Crea un Droplet** (Ubuntu recomendado)

2. **SSH al droplet:**
```bash
ssh root@IP_DROPLET
```

3. **Despliega** siguiendo los pasos de Linux

### Heroku (con Docker)

1. **Instala Heroku CLI**

2. **Login:**
```bash
heroku login
heroku container:login
```

3. **Crea la app:**
```bash
heroku create nombre-app
```

4. **Despliega:**
```bash
heroku container:push web
heroku container:release web
heroku open
```

---

## 6. Agregar Nuevos Usuarios

Cada usuario nuevo requiere su propia base de datos en el sistema multi-tenant.

### Opción A: Usando el Script SQL (Recomendado)

1. **Editar el script [add_new_user.sql](add_new_user.sql):**
   
   Cambiar la línea 10:
   ```sql
   SET @nuevo_usuario = 'NUEVO_USUARIO';  -- Cambiar 'NUEVO_USUARIO' por el nombre real
   ```

2. **Ejecutar el script en MySQL:**
   ```bash
   mysql -u root -p < add_new_user.sql
   ```

3. **Agregar el usuario en [main.go](main.go) (líneas 34-50):**
   ```go
   userConfig = map[string]map[string]string{
       "admin": {
           "password": "admin123",
           "dbname":   "webapp_admin",
       },
       "usuario1": {
           "password": "pass123",
           "dbname":   "webapp_usuario1",
       },
       "demo": {
           "password": "demo123",
           "dbname":   "webapp_demo",
       },
       // ⬇️ AGREGAR EL NUEVO USUARIO AQUÍ
       "nuevo_usuario": {
           "password": "password_seguro",
           "dbname":   "webapp_nuevo_usuario",
       },
   }
   ```

4. **Reiniciar la aplicación**

### Opción B: Creación Manual

1. **Conectar a MySQL:**
   ```bash
   mysql -u root -p
   ```

2. **Crear la base de datos:**
   ```sql
   CREATE DATABASE webapp_nuevo_usuario;
   USE webapp_nuevo_usuario;
   ```

3. **Crear las tablas:**
   ```sql
   CREATE TABLE user_data (
       id INT AUTO_INCREMENT PRIMARY KEY,
       key_name VARCHAR(100) UNIQUE NOT NULL,
       value VARCHAR(255),
       created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
       updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
   );

   CREATE TABLE activity_log (
       id INT AUTO_INCREMENT PRIMARY KEY,
       action VARCHAR(100) NOT NULL,
       description TEXT,
       created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
   );

   CREATE TABLE session_data (
       id INT AUTO_INCREMENT PRIMARY KEY,
       session_key VARCHAR(100) UNIQUE NOT NULL,
       session_value TEXT,
       created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
   );
   ```

4. **Añadir en [main.go](main.go)** como se mostró arriba.

### Arquitectura Multi-tenant

```
┌─────────────────────┐
│   Aplicación Go     │
└──────────┬──────────┘
           │
    ┌──────┴──────┐
    │             │
┌───▽───┐   ┌────▽────┐   ┌────▽────┐
│ BD 1  │   │  BD 2   │   │  BD 3   │
│admin  │   │usuario1 │   │  demo   │
└───────┘   └─────────┘   └─────────┘
```

**Ventajas:**
- ✅ Aislamiento total de datos por usuario
- ✅ Seguridad mejorada
- ✅ Cada usuario puede tener diferentes configuraciones
- ✅ Backups individuales por usuario
- ✅ Facilita cumplimiento regulatorio

**Consideraciones:**
- Requiere más recursos de BD
- Gestión de múltiples esquemas
- Cada BD debe mantenerse

### Cambiar el puerto de la aplicación

En [main.go](main.go), última línea:
```go
log.Fatal(http.ListenAndServe(":3000", nil))  // Cambia 8080 por el puerto que prefieras
```

### Usar HTTPS (SSL/TLS)

Para conexiones seguras:

1. **Obtén certificados SSL** (gratis con Let's Encrypt)

2. **Modifica main.go:**
```go
log.Fatal(http.ListenAndServeTLS(":443", "cert.pem", "key.pem", nil))
```

---

## 7. Solución de Problemas

### La aplicación no inicia

**Error:** `bind: address already in use`
- **Solución:** El puerto 8080 está ocupado. Cambia el puerto en main.go o mata el proceso:
```powershell
# Windows
netstat -ano | findstr :8080
taskkill /PID <PID> /F

# Linux
sudo lsof -i :8080
sudo kill -9 <PID>
```

### No puedo acceder desde otra máquina

1. **Verifica que usas `0.0.0.0:8080`** en lugar de `localhost:8080`
2. **Revisa el firewall** está permitiendo el puerto 8080
3. **Prueba con ping:**
```powershell
ping IP_SERVIDOR
```
4. **Asegúrate de usar la IP correcta** (no uses 127.0.0.1)

### El navegador no carga la página

1. **Verifica que el servidor está corriendo:**
```powershell
# Ver el proceso
Get-Process | Where-Object {$_.ProcessName -like "*go*"}
```

2. **Revisa los logs** en la terminal donde corre la app

3. **Prueba localmente primero:**
```
http://localhost:8080
```

### Problemas con templates o CSS

**Error:** `template not found`
- **Causa:** Las carpetas `templates/` y `static/` no están en el mismo directorio que el ejecutable
- **Solución:** Asegúrate de copiar estas carpetas junto al ejecutable

### Sesiones no funcionan

- Cambia la clave secreta en [main.go](main.go), línea 15:
```go
store = sessions.NewCookieStore([]byte("tu-nueva-clave-super-secreta-unica"))
```

---

## 🔒 Recomendaciones de Seguridad

Para usar en producción:

1. ✅ **Cambia la clave secreta** de las sesiones
2. ✅ **Usa HTTPS** (SSL/TLS)
3. ✅ **Implementa hash de contraseñas** (bcrypt)
4. ✅ **Usa una base de datos** real en lugar del mapa en memoria
5. ✅ **Implementa rate limiting** para prevenir ataques de fuerza bruta
6. ✅ **Configura CORS** si necesitas acceso desde otros dominios
7. ✅ **Mantén Go actualizado** con parches de seguridad
8. ✅ **Usa un proxy reverso** como Nginx o Caddy

---

## 📞 Resumen Rápido

### Para LAN (red local):
1. Cambia `:8080` a `0.0.0.0:8080` en main.go
2. Abre el firewall: `New-NetFirewallRule -DisplayName "WebApp" -Direction Inbound -LocalPort 8080 -Protocol TCP -Action Allow`
3. Obtén tu IP: `ipconfig`
4. Comparte: `http://TU_IP:8080`

### Para servidor remoto:
1. Compila: `go build -o webapp-linux main.go`
2. Copia archivos al servidor (webapp-linux, templates/, static/)
3. Ejecuta: `./webapp-linux`
4. Accede: `http://IP_SERVIDOR:8080`

### Para Docker:
1. `docker build -t webapp-go .`
2. `docker run -d -p 8080:8080 webapp-go`

---

¡Tu aplicación está lista para ser usada por múltiples usuarios! 🚀
