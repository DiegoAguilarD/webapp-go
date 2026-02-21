-- Script de inicialización para bases de datos Multi-tenant
-- Ejecutar como root de MySQL

-- ============================================
-- PASO 1: Crear las bases de datos
-- ============================================

CREATE DATABASE IF NOT EXISTS webapp_admin;
CREATE DATABASE IF NOT EXISTS webapp_usuario1;
CREATE DATABASE IF NOT EXISTS webapp_demo;

-- ============================================
-- PASO 2: Crear tablas en webapp_admin
-- ============================================

USE webapp_admin;

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

-- Insertar datos iniciales para admin
INSERT IGNORE INTO user_data (key_name, value) VALUES 
    ('username', 'admin'),
    ('last_login', NOW()),
    ('login_count', '0');

INSERT INTO activity_log (action, description) VALUES 
    ('SYSTEM', 'Base de datos inicializada');

-- ============================================
-- PASO 3: Crear tablas en webapp_usuario1
-- ============================================

USE webapp_usuario1;

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

-- Insertar datos iniciales para usuario1
INSERT IGNORE INTO user_data (key_name, value) VALUES 
    ('username', 'usuario1'),
    ('last_login', NOW()),
    ('login_count', '0');

INSERT INTO activity_log (action, description) VALUES 
    ('SYSTEM', 'Base de datos inicializada');

-- ============================================
-- PASO 4: Crear tablas en webapp_demo
-- ============================================

USE webapp_demo;

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

-- Insertar datos iniciales para demo
INSERT IGNORE INTO user_data (key_name, value) VALUES 
    ('username', 'demo'),
    ('last_login', NOW()),
    ('login_count', '0');

INSERT INTO activity_log (action, description) VALUES 
    ('SYSTEM', 'Base de datos inicializada');

-- ============================================
-- PASO 5: Verificación
-- ============================================

-- Mostrar las bases de datos creadas
SHOW DATABASES LIKE 'webapp_%';

-- Verificar tablas en cada BD
SELECT 'webapp_admin' as database_name, TABLE_NAME 
FROM INFORMATION_SCHEMA.TABLES 
WHERE TABLE_SCHEMA = 'webapp_admin'
UNION ALL
SELECT 'webapp_usuario1', TABLE_NAME 
FROM INFORMATION_SCHEMA.TABLES 
WHERE TABLE_SCHEMA = 'webapp_usuario1'
UNION ALL
SELECT 'webapp_demo', TABLE_NAME 
FROM INFORMATION_SCHEMA.TABLES 
WHERE TABLE_SCHEMA = 'webapp_demo';

-- Fin del script
SELECT 'Inicialización completada exitosamente!' as status;
