package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/sessions"
)

// Configuración de base de datos
type DBConfig struct {
	Host     string
	Port     string
	RootUser string
	RootPass string
}

// Pool de conexiones por usuario
var (
	// Store de sesiones
	store = sessions.NewCookieStore([]byte("clave-secreta-super-segura-12345"))

	// Configuración de MySQL (se puede sobrescribir con variables de entorno)
	dbConfig = DBConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "3306"),
		RootUser: getEnv("DB_ROOT_USER", "root"),
		RootPass: getEnv("DB_ROOT_PASSWORD", "admin"), // Cambiar en producción
	}

	// Mapa de usuarios y sus credenciales
	// Formato: username -> {password, database_name, token}
	userConfig = map[string]map[string]string{
		"admin": {
			"password": "admin123",
			"dbname":   "webapp_admin",
			"token":    "admin_token_demo",
		},
		"usuario1": {
			"password": "pass123",
			"dbname":   "webapp_usuario1",
			"token":    "usuario1_token_demo",
		},
		"demo": {
			"password": "demo123",
			"dbname":   "webapp_demo",
			"token":    "demo_token_demo",
		},
	}

	// Mapa inverso: token -> username
	tokenToUser = make(map[string]string)

	// Pool de conexiones a bases de datos por usuario
	dbConnections = make(map[string]*sql.DB)
	dbMutex       sync.RWMutex

	templates *template.Template
)

// Helper para obtener variables de entorno con valor por defecto
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func init() {
	// Cargar todos los templates
	templates = template.Must(template.ParseGlob("templates/*.html"))

	// Inicializar mapa inverso de tokens
	for username, config := range userConfig {
		if token, ok := config["token"]; ok {
			tokenToUser[token] = username
		}
	}
}

// Función para generar un token único
func generateUniqueToken() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		log.Printf("Error al generar token: %v", err)
		return fmt.Sprintf("token_%d", time.Now().Unix())
	}
	return hex.EncodeToString(bytes)
}

// Función para obtener o crear la conexión a la BD del usuario
func getUserDB(username string) (*sql.DB, error) {
	dbMutex.RLock()
	db, exists := dbConnections[username]
	dbMutex.RUnlock()

	if exists && db != nil {
		return db, nil
	}

	// Crear nueva conexión
	dbMutex.Lock()
	defer dbMutex.Unlock()

	// Verificar nuevamente por si otro goroutine ya creó la conexión
	if db, exists := dbConnections[username]; exists && db != nil {
		return db, nil
	}

	config, ok := userConfig[username]
	if !ok {
		return nil, fmt.Errorf("usuario no configurado: %s", username)
	}

	dbname := config["dbname"]
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		dbConfig.RootUser, dbConfig.RootPass, dbConfig.Host, dbConfig.Port, dbname)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("error al conectar a BD %s: %v", dbname, err)
	}

	// Configurar pool de conexiones
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	// Verificar conexión
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error al hacer ping a BD %s: %v", dbname, err)
	}

	dbConnections[username] = db
	log.Printf("Conexión establecida a BD: %s", dbname)
	return db, nil
}

// Función para inicializar las bases de datos de todos los usuarios
func initializeDatabases() error {
	// Conectar como root para crear las bases de datos
	rootDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/",
		dbConfig.RootUser, dbConfig.RootPass, dbConfig.Host, dbConfig.Port)

	rootDB, err := sql.Open("mysql", rootDSN)
	if err != nil {
		return fmt.Errorf("error al conectar como root: %v", err)
	}
	defer rootDB.Close()

	// Crear base de datos para cada usuario
	for username, config := range userConfig {
		dbname := config["dbname"]

		// Crear base de datos si no existe
		_, err := rootDB.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", dbname))
		if err != nil {
			log.Printf("Error al crear BD %s: %v", dbname, err)
			continue
		}

		// Conectar a la BD del usuario
		userDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
			dbConfig.RootUser, dbConfig.RootPass, dbConfig.Host, dbConfig.Port, dbname)

		userDB, err := sql.Open("mysql", userDSN)
		if err != nil {
			log.Printf("Error al conectar a BD %s: %v", dbname, err)
			continue
		}

		// Crear tabla user_data
		_, err = userDB.Exec(`
			CREATE TABLE IF NOT EXISTS user_data (
				id INT AUTO_INCREMENT PRIMARY KEY,
				key_name VARCHAR(100) UNIQUE NOT NULL,
				value VARCHAR(255),
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
			)
		`)
		if err != nil {
			log.Printf("Error al crear tabla user_data en BD %s: %v", dbname, err)
			userDB.Close()
			continue
		}

		// Crear tabla activity_log
		_, err = userDB.Exec(`
			CREATE TABLE IF NOT EXISTS activity_log (
				id INT AUTO_INCREMENT PRIMARY KEY,
				action VARCHAR(100) NOT NULL,
				description TEXT,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)
		`)
		if err != nil {
			log.Printf("Error al crear tabla activity_log en BD %s: %v", dbname, err)
			userDB.Close()
			continue
		}

		// Crear tabla session_data
		_, err = userDB.Exec(`
			CREATE TABLE IF NOT EXISTS session_data (
				id INT AUTO_INCREMENT PRIMARY KEY,
				session_key VARCHAR(100) UNIQUE NOT NULL,
				session_value TEXT,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)
		`)
		if err != nil {
			log.Printf("Error al crear tabla session_data en BD %s: %v", dbname, err)
			userDB.Close()
			continue
		}

		// Insertar datos iniciales
		_, err = userDB.Exec(`
			INSERT IGNORE INTO user_data (key_name, value) VALUES 
			('username', ?),
			('last_login', NOW()),
			('login_count', '0')
		`, username)

		userDB.Close()
		log.Printf("BD %s inicializada correctamente", dbname)
	}

	return nil
}

// Función para crear un nuevo usuario con su base de datos
func createNewUser(username, password string) error {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	// Verificar si el usuario ya existe
	if _, exists := userConfig[username]; exists {
		return fmt.Errorf("el usuario ya existe")
	}

	// Validar nombre de usuario (solo alfanuméricos y guion bajo)
	for _, char := range username {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') || char == '_') {
			return fmt.Errorf("nombre de usuario inválido: solo letras, números y guion bajo")
		}
	}

	if len(username) < 3 || len(username) > 20 {
		return fmt.Errorf("el nombre de usuario debe tener entre 3 y 20 caracteres")
	}

	if len(password) < 6 {
		return fmt.Errorf("la contraseña debe tener al menos 6 caracteres")
	}

	dbname := fmt.Sprintf("webapp_%s", username)

	// Conectar como root para crear la base de datos
	rootDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/",
		dbConfig.RootUser, dbConfig.RootPass, dbConfig.Host, dbConfig.Port)

	rootDB, err := sql.Open("mysql", rootDSN)
	if err != nil {
		return fmt.Errorf("error al conectar como root: %v", err)
	}
	defer rootDB.Close()

	// Crear base de datos
	_, err = rootDB.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", dbname))
	if err != nil {
		return fmt.Errorf("error al crear base de datos: %v", err)
	}

	// Conectar a la nueva BD
	userDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
		dbConfig.RootUser, dbConfig.RootPass, dbConfig.Host, dbConfig.Port, dbname)

	userDB, err := sql.Open("mysql", userDSN)
	if err != nil {
		return fmt.Errorf("error al conectar a la nueva BD: %v", err)
	}
	defer userDB.Close()

	// Crear tabla user_data
	_, err = userDB.Exec(`
		CREATE TABLE IF NOT EXISTS user_data (
			id INT AUTO_INCREMENT PRIMARY KEY,
			key_name VARCHAR(100) UNIQUE NOT NULL,
			value VARCHAR(255),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("error al crear tabla user_data: %v", err)
	}

	// Crear tabla activity_log
	_, err = userDB.Exec(`
		CREATE TABLE IF NOT EXISTS activity_log (
			id INT AUTO_INCREMENT PRIMARY KEY,
			action VARCHAR(100) NOT NULL,
			description TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("error al crear tabla activity_log: %v", err)
	}

	// Crear tabla session_data
	_, err = userDB.Exec(`
		CREATE TABLE IF NOT EXISTS session_data (
			id INT AUTO_INCREMENT PRIMARY KEY,
			session_key VARCHAR(100) UNIQUE NOT NULL,
			session_value TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("error al crear tabla session_data: %v", err)
	}

	// Insertar datos iniciales
	_, err = userDB.Exec(`
		INSERT INTO user_data (key_name, value) VALUES 
		('username', ?),
		('last_login', NOW()),
		('login_count', '0')
	`, username)
	if err != nil {
		return fmt.Errorf("error al insertar datos iniciales: %v", err)
	}

	// Registrar en activity_log
	_, err = userDB.Exec(`
		INSERT INTO activity_log (action, description) VALUES (?, ?)
	`, "REGISTER", fmt.Sprintf("Usuario %s registrado exitosamente", username))
	if err != nil {
		log.Printf("Advertencia: error al registrar actividad: %v", err)
	}

	// Generar token único para el usuario
	token := generateUniqueToken()

	// Guardar token en la base de datos
	_, err = userDB.Exec(`
		INSERT INTO user_data (key_name, value) VALUES ('access_token', ?)
	`, token)
	if err != nil {
		log.Printf("Advertencia: error al guardar token: %v", err)
	}

	// Agregar al mapa de usuarios
	userConfig[username] = map[string]string{
		"password": password,
		"dbname":   dbname,
		"token":    token,
	}

	// Agregar al mapa inverso
	tokenToUser[token] = username

	log.Printf("Nuevo usuario creado: %s con BD: %s y token: %s", username, dbname, token)
	return nil
}

func main() {
	// Inicializar bases de datos
	log.Println("Inicializando bases de datos...")
	if err := initializeDatabases(); err != nil {
		log.Printf("Advertencia: Error al inicializar BDs: %v", err)
		log.Println("Continuando sin bases de datos...")
	}

	// Configurar rutas
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/dashboard", dashboardHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/app/", appAccessHandler) // Nueva ruta para acceso por token
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Iniciar servidor
	log.Println("Servidor iniciado en http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Handler de la página principal
func homeHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")

	// Si ya está autenticado, redirigir al dashboard
	if auth, ok := session.Values["authenticated"].(bool); ok && auth {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	templates.ExecuteTemplate(w, "login.html", nil)
}

// Handler del login
func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	// Verificar credenciales
	userCfg, ok := userConfig[username]
	if !ok || userCfg["password"] != password {
		// Login fallido
		data := map[string]interface{}{
			"Error": "Usuario o contraseña incorrectos",
		}
		templates.ExecuteTemplate(w, "login.html", data)
		return
	}

	// Login exitoso - registrar en la BD del usuario
	db, err := getUserDB(username)
	if err != nil {
		log.Printf("Error al obtener BD para %s: %v", username, err)
	} else {
		// Actualizar última conexión y contador de logins
		_, err = db.Exec(`
			INSERT INTO user_data (key_name, value) VALUES ('last_login', NOW())
			ON DUPLICATE KEY UPDATE value = NOW()
		`)
		if err != nil {
			log.Printf("Error al actualizar last_login: %v", err)
		}

		_, err = db.Exec(`
			INSERT INTO user_data (key_name, value) VALUES ('login_count', '1')
			ON DUPLICATE KEY UPDATE value = CAST(value AS UNSIGNED) + 1
		`)
		if err != nil {
			log.Printf("Error al actualizar login_count: %v", err)
		}

		// Registrar actividad
		_, err = db.Exec(`
			INSERT INTO activity_log (action, description) VALUES (?, ?)
		`, "LOGIN", fmt.Sprintf("Usuario %s inició sesión", username))
		if err != nil {
			log.Printf("Error al registrar actividad: %v", err)
		}
	}

	// Crear sesión
	session, _ := store.Get(r, "session-name")
	session.Values["authenticated"] = true
	session.Values["username"] = username
	session.Save(r, w)

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

// Handler del dashboard (requiere autenticación)
func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")

	// Verificar si está autenticado
	auth, ok := session.Values["authenticated"].(bool)
	if !ok || !auth {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	username := session.Values["username"].(string)

	// Obtener datos de la BD del usuario
	db, err := getUserDB(username)

	stats := map[string]interface{}{
		"Usuarios":   len(userConfig),
		"Sesiones":   1,
		"Visitantes": 0,
		"LoginCount": 0,
	}

	activities := []string{
		"Login exitoso",
		"Acceso al dashboard",
		"Sistema operando",
	}

	if err != nil {
		log.Printf("Error al conectar a BD para dashboard: %v", err)
	} else {
		// Obtener contador de logins
		var loginCount string
		err = db.QueryRow("SELECT value FROM user_data WHERE key_name = 'login_count'").Scan(&loginCount)
		if err == nil {
			stats["LoginCount"] = loginCount
		}

		// Obtener últimas actividades
		rows, err := db.Query(`
			SELECT action, description, created_at 
			FROM activity_log 
			ORDER BY created_at DESC 
			LIMIT 5
		`)
		if err == nil {
			defer rows.Close()
			activities = []string{}
			for rows.Next() {
				var action, description string
				var createdAt time.Time
				if err := rows.Scan(&action, &description, &createdAt); err == nil {
					activities = append(activities, fmt.Sprintf("[%s] %s",
						createdAt.Format("15:04:05"), description))
				}
			}
		}

		// Contar registros en activity_log
		var visitantes int
		err = db.QueryRow("SELECT COUNT(*) FROM activity_log").Scan(&visitantes)
		if err == nil {
			stats["Visitantes"] = visitantes
		}
	}

	// Datos para mostrar en el dashboard
	data := map[string]interface{}{
		"Username":   username,
		"Time":       time.Now().Format("02/01/2006 15:04:05"),
		"Stats":      stats,
		"Activities": activities,
		"Database":   userConfig[username]["dbname"],
	}

	templates.ExecuteTemplate(w, "dashboard.html", data)
}

// Handler del logout
func logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")

	// Registrar logout en BD si está autenticado
	if username, ok := session.Values["username"].(string); ok {
		db, err := getUserDB(username)
		if err == nil {
			_, err = db.Exec(`
				INSERT INTO activity_log (action, description) VALUES (?, ?)
			`, "LOGOUT", fmt.Sprintf("Usuario %s cerró sesión", username))
			if err != nil {
				log.Printf("Error al registrar logout: %v", err)
			}
		}
	}

	session.Values["authenticated"] = false
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Handler de la página de registro
func registerHandler(w http.ResponseWriter, r *http.Request) {
	// Si es GET, mostrar el formulario
	if r.Method == "GET" {
		templates.ExecuteTemplate(w, "register.html", nil)
		return
	}

	// Si es POST, procesar el registro
	if r.Method != "POST" {
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirm_password")

	// Validar que los campos no estén vacíos
	if username == "" || password == "" || confirmPassword == "" {
		data := map[string]interface{}{
			"Error": "Todos los campos son obligatorios",
		}
		templates.ExecuteTemplate(w, "register.html", data)
		return
	}

	// Validar que las contraseñas coincidan
	if password != confirmPassword {
		data := map[string]interface{}{
			"Error": "Las contraseñas no coinciden",
		}
		templates.ExecuteTemplate(w, "register.html", data)
		return
	}

	// Crear el nuevo usuario
	err := createNewUser(username, password)
	if err != nil {
		data := map[string]interface{}{
			"Error": fmt.Sprintf("Error al crear el usuario: %s", err.Error()),
		}
		templates.ExecuteTemplate(w, "register.html", data)
		return
	}

	// Registro exitoso - obtener el token generado
	token := userConfig[username]["token"]

	// Construir el enlace de acceso único
	accessURL := fmt.Sprintf("http://%s/app/%s", r.Host, token)

	// Mostrar página de éxito con el enlace único
	data := map[string]interface{}{
		"Success":   true,
		"Username":  username,
		"AccessURL": accessURL,
		"Token":     token,
	}

	log.Printf("Usuario %s registrado exitosamente. Enlace de acceso: %s", username, accessURL)
	templates.ExecuteTemplate(w, "success_register.html", data)
}

// Handler para acceso por token único
func appAccessHandler(w http.ResponseWriter, r *http.Request) {
	// Extraer el token de la URL (/app/{token})
	path := strings.TrimPrefix(r.URL.Path, "/app/")
	token := strings.TrimSpace(path)

	if token == "" {
		http.Error(w, "Token no proporcionado", http.StatusBadRequest)
		return
	}

	// Buscar el usuario asociado al token
	dbMutex.RLock()
	username, exists := tokenToUser[token]
	dbMutex.RUnlock()

	if !exists {
		data := map[string]interface{}{
			"Error": "Token inválido o expirado. Por favor, contacta al administrador.",
		}
		templates.ExecuteTemplate(w, "login.html", data)
		return
	}

	// Verificar que el usuario existe en userConfig
	if _, ok := userConfig[username]; !ok {
		data := map[string]interface{}{
			"Error": "Usuario no encontrado. Por favor, contacta al administrador.",
		}
		templates.ExecuteTemplate(w, "login.html", data)
		return
	}

	// Token válido - crear sesión automáticamente
	session, _ := store.Get(r, "session-name")
	session.Values["authenticated"] = true
	session.Values["username"] = username
	session.Save(r, w)

	// Registrar actividad de acceso por token
	db, err := getUserDB(username)
	if err == nil {
		_, err = db.Exec(`
			INSERT INTO activity_log (action, description) VALUES (?, ?)
		`, "TOKEN_ACCESS", fmt.Sprintf("Usuario %s accedió mediante enlace único", username))
		if err != nil {
			log.Printf("Error al registrar actividad de token: %v", err)
		}
	}

	log.Printf("Usuario %s autenticado exitosamente mediante token", username)
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}
