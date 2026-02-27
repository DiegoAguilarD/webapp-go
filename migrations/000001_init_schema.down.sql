-- Rollback de la migración inicial: Eliminar tablas del esquema

DROP TABLE IF EXISTS session_data;
DROP TABLE IF EXISTS activity_log;
DROP TABLE IF EXISTS user_data;
