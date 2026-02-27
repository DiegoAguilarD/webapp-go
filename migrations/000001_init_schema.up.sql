-- Migración inicial: Crear tablas del esquema multi-tenant
-- Cada base de datos de cliente (db_[slug]) ejecuta esta migración.

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

INSERT INTO activity_log (action, description) VALUES
    ('SYSTEM', 'Base de datos inicializada mediante migración');
