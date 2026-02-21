-- Script para agregar un nuevo usuario con su propia base de datos
-- Ejecutar como root de MySQL

-- INSTRUCCIONES:
-- 1. Reemplazar 'NUEVO_USUARIO' con el nombre de usuario deseado
-- 2. Ejecutar este script en MySQL
-- 3. Agregar el usuario a main.go en el mapa userConfig

-- Variables (modificar estos valores)
SET @nuevo_usuario = 'NUEVO_USUARIO';  -- Cambiar aquí
SET @nueva_bd = CONCAT('webapp_', @nuevo_usuario);

-- Preparar el statement dinámico para crear la BD
SET @sql_create_db = CONCAT('CREATE DATABASE IF NOT EXISTS ', @nueva_bd);
PREPARE stmt FROM @sql_create_db;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Usar la nueva base de datos
SET @sql_use_db = CONCAT('USE ', @nueva_bd);
PREPARE stmt FROM @sql_use_db;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Crear tablas
CREATE TABLE IF NOT EXISTS user_data (
    id INT AUTO_INCREMENT PRIMARY KEY,
    key_name VARCHAR(100) UNIQUE NOT NULL,
    value VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_key_name (key_name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS activity_log (
    id INT AUTO_INCREMENT PRIMARY KEY,
    action VARCHAR(100) NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_created_at (created_at),
    INDEX idx_action (action)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS session_data (
    id INT AUTO_INCREMENT PRIMARY KEY,
    session_key VARCHAR(100) UNIQUE NOT NULL,
    session_value TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_session_key (session_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Insertar datos iniciales
INSERT IGNORE INTO user_data (key_name, value) 
SELECT 'username', @nuevo_usuario
UNION ALL
SELECT 'last_login', NOW()
UNION ALL
SELECT 'login_count', '0';

INSERT INTO activity_log (action, description) VALUES 
    ('SYSTEM', CONCAT('Base de datos creada para usuario: ', @nuevo_usuario));

-- Mostrar confirmación
SELECT CONCAT('Base de datos "', @nueva_bd, '" creada exitosamente para usuario "', @nuevo_usuario, '"') AS resultado;

-- Recordatorio
SELECT '
IMPORTANTE: Agregar el siguiente código en main.go dentro del mapa userConfig:

"NUEVO_USUARIO": {
    "password": "TU_PASSWORD_AQUI",
    "dbname":   "webapp_NUEVO_USUARIO",
},

' AS recordatorio;
