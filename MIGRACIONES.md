# 🔄 Estrategia de Migraciones de Base de Datos

## Resumen

En una arquitectura multi-tenant donde cada cliente tiene su propia base de datos (`db_[slug]`), las migraciones deben ejecutarse de forma aislada en cada base de datos al momento del despliegue. Esta guía describe la estrategia recomendada usando [golang-migrate](https://github.com/golang-migrate/migrate).

---

## Estructura de Archivos

Las migraciones se encuentran en el directorio `migrations/` con el siguiente formato:

```
migrations/
├── 000001_init_schema.up.sql      # Crear tablas iniciales
├── 000001_init_schema.down.sql    # Rollback de tablas iniciales
├── 000002_add_feature.up.sql      # Siguiente migración (ejemplo)
└── 000002_add_feature.down.sql    # Rollback (ejemplo)
```

Cada par de archivos (`up`/`down`) representa una migración versionada:
- **`up.sql`**: Aplica el cambio (crear tablas, agregar columnas, etc.)
- **`down.sql`**: Revierte el cambio (rollback)

---

## Integración con Go

### Opción recomendada: golang-migrate

La librería [`golang-migrate`](https://github.com/golang-migrate/migrate) permite ejecutar migraciones desde la línea de comandos o de forma programática.

#### Instalación en el Dockerfile

La herramienta CLI de `migrate` se incluye en la imagen Docker para que Ansible pueda invocarla:

```dockerfile
# En la etapa final del Dockerfile
RUN apk --no-cache add curl \
    && curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz \
    | tar -xz -C /usr/local/bin/
COPY migrations/ /app/migrations/
```

#### Ejecución de migraciones

Ansible ejecuta las migraciones dentro del contenedor del cliente tras el despliegue:

```bash
# Aplicar todas las migraciones pendientes
/app/webapp -migrate

# O usando la CLI de golang-migrate directamente
migrate -path /app/migrations \
        -database "mysql://${DB_USER}:${DB_PASSWORD}@tcp(${DB_HOST}:${DB_PORT})/${DB_NAME}" \
        up
```

### Opción alternativa: Migraciones integradas en la aplicación

Si se prefiere no instalar herramientas adicionales, la aplicación puede ejecutar migraciones al arrancar con el flag `-migrate`:

```go
import "github.com/golang-migrate/migrate/v4"

func runMigrations(dbURL string) error {
    m, err := migrate.New("file:///app/migrations", dbURL)
    if err != nil {
        return err
    }
    if err := m.Up(); err != nil && err != migrate.ErrNoChange {
        return err
    }
    return nil
}
```

---

## Flujo de Despliegue

El siguiente diagrama muestra cómo las migraciones se integran en el pipeline:

```
Jenkins (Jenkinsfile)
  │
  ├─ Parámetros: CLIENT_NAME, SLUG, ADMIN_EMAIL, ADMIN_PASSWORD
  ├─ Buscar puerto libre (≥ 9000)
  │
  └─ Ansible (app_deploy.yml)
       │
       ├─ 1. Construir imagen Docker (multi-stage build)
       ├─ 2. Generar contraseña aleatoria para BD
       ├─ 3. Crear BD (db_[slug]) y usuario en MariaDB compartido
       ├─ 4. Desplegar contenedor del cliente (puerto asignado)
       └─ 5. Ejecutar migraciones ← aquí se aplican los .sql
```

---

## Buenas Prácticas

1. **Nunca editar migraciones ya aplicadas**: Crear una nueva migración para cambios adicionales.
2. **Migraciones idempotentes**: Usar `IF NOT EXISTS` / `IF EXISTS` en las sentencias SQL.
3. **Una migración por cambio**: Cada archivo debe representar un cambio atómico.
4. **Probar rollbacks**: Verificar que cada `down.sql` revierte correctamente su `up.sql`.
5. **Control de versiones**: Los archivos de migración se versionan junto al código fuente.

---

## Ejemplo: Agregar una nueva migración

Para agregar un nuevo campo `email` a la tabla `user_data`:

**`migrations/000002_add_email_to_user_data.up.sql`**
```sql
ALTER TABLE user_data ADD COLUMN email VARCHAR(255) AFTER value;
```

**`migrations/000002_add_email_to_user_data.down.sql`**
```sql
ALTER TABLE user_data DROP COLUMN email;
```

Tras agregar los archivos, el pipeline aplicará la migración automáticamente en el próximo despliegue de cada cliente.
