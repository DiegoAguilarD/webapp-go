# 🚀 Inicio Rápido - Sistema Multi-tenant

## ¿Qué es esto?

Esta aplicación ahora utiliza **una base de datos MySQL independiente por cada usuario**, proporcionando:
- ✅ Aislamiento total de datos entre usuarios
- ✅ Seguridad mejorada
- ✅ Escalabilidad para múltiples clientes

## Estructura de Bases de Datos

```
Usuario "admin"    → Base de Datos: webapp_admin
Usuario "usuario1" → Base de Datos: webapp_usuario1
Usuario "demo"     → Base de Datos: webapp_demo
```

## Inicio Rápido

### Opción 1: Con Docker (Más Fácil) ⭐

```bash
# Inicia todo (MySQL + Aplicación)
docker-compose up -d

# Ver logs
docker-compose logs -f

# Acceder
http://localhost:8080
```

### Opción 2: Sin Docker

1. **Instalar MySQL:**
   ```bash
   # Windows: Descargar de mysql.com
   # Linux: sudo apt install mysql-server
   # Mac: brew install mysql
   ```

2. **Configurar credenciales en main.go (línea 30):**
   ```go
   RootPass: "TU_PASSWORD_MYSQL", // Cambiar aquí
   ```

3. **Ejecutar:**
   ```bash
   go mod download
   go run main.go
   ```

La aplicación creará automáticamente las 3 bases de datos.

## Usuarios Predefinidos

| Usuario | Password | Base de Datos |
|---------|----------|---------------|
| admin | admin123 | webapp_admin |
| usuario1 | pass123 | webapp_usuario1 |
| demo | demo123 | webapp_demo |

## Agregar un Nuevo Usuario

### Método Rápido:

1. Editar `add_new_user.sql` (línea 10):
   ```sql
   SET @nuevo_usuario = 'NOMBRE_NUEVO';
   ```

2. Ejecutar:
   ```bash
   mysql -u root -p < add_new_user.sql
   ```

3. Agregar en `main.go` (línea 37):
   ```go
   "nombre_nuevo": {
       "password": "password123",
       "dbname":   "webapp_nombre_nuevo",
   },
   ```

4. Reiniciar la aplicación

## Archivos Importantes

| Archivo | Propósito |
|---------|-----------|
| `main.go` | Código principal con configuración de usuarios |
| `init_databases.sql` | Script para crear todas las bases de datos |
| `add_new_user.sql` | Script para agregar un usuario nuevo |
| `docker-compose.yml` | Configuración Docker con MySQL |
| `config.example.env` | Ejemplo de configuración |

## Comandos Útiles

### Con Docker:
```bash
# Detener todo
docker-compose down

# Ver bases de datos
docker-compose exec mysql mysql -uroot -ppassword -e "SHOW DATABASES LIKE 'webapp_%';"

# Backup de una BD
docker-compose exec mysql mysqldump -uroot -ppassword webapp_admin > backup.sql

# Ver logs de MySQL
docker-compose logs mysql
```

### Sin Docker:
```bash
# Ver bases de datos
mysql -u root -p -e "SHOW DATABASES LIKE 'webapp_%';"

# Backup
mysqldump -u root -p webapp_admin > backup_admin.sql

# Restaurar
mysql -u root -p webapp_admin < backup_admin.sql
```

## Despliegue en Producción

Ver documentación completa en [DESPLIEGUE.md](DESPLIEGUE.md)

### Resumen:
- **Red local**: Cambiar `":8080"` por `"0.0.0.0:8080"` en main.go
- **Servidor Linux**: Compilar y copiar ejecutable + carpetas
- **Docker**: `docker-compose up -d`
- **Nube (AWS)**: Instalar Docker en EC2 y ejecutar

## Solución de Problemas

### Error: "Can't connect to MySQL server"
- Verificar que MySQL esté corriendo: `mysql -u root -p`
- Revisar password en main.go (línea 30)

### Error: "Access denied for user"
- Cambiar password en main.go para que coincida con MySQL

### Error: "Database does not exist"
- La aplicación crea las BDs automáticamente
- O ejecutar manualmente: `mysql -u root -p < init_databases.sql`

### Ver logs de la aplicación:
```bash
# Sin Docker
go run main.go

# Con Docker
docker-compose logs webapp
```

## Arquitectura

```
┌─────────────────────────────────────┐
│        Aplicación Web (Go)          │
│  - Autenticación                    │
│  - Gestión de sesiones              │
│  - Router por BD según usuario      │
└──────────────┬──────────────────────┘
               │
    ┌──────────┼──────────┐
    │          │          │
┌───▽───┐  ┌───▽───┐  ┌──▽────┐
│ MySQL │  │ MySQL │  │ MySQL │
│  BD1  │  │  BD2  │  │  BD3  │
│ admin │  │usuario│  │ demo  │
└───────┘  └───────┘  └───────┘
```

## Seguridad

⚠️ **Antes de producción:**
1. Cambiar passwords en `main.go`
2. Cambiar `SESSION_SECRET` (línea 27)
3. Usar HTTPS (ver DESPLIEGUE.md)
4. Restringir acceso de red a MySQL
5. Configurar firewall

## Soporte

- Documentación completa: [DESPLIEGUE.md](DESPLIEGUE.md)
- README: [README.md](README.md)
- Scripts SQL: `init_databases.sql`, `add_new_user.sql`
