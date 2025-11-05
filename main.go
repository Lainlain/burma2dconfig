package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

// AppVersion represents app version and update information
type AppVersion struct {
	LatestVersion      string   `json:"latest_version"`
	LatestVersionCode  int      `json:"latest_version_code"`
	MinimumVersionCode int      `json:"minimum_version_code"`
	ForceUpdate        bool     `json:"force_update"`
	UpdateTitle        string   `json:"update_title"`
	UpdateMessage      string   `json:"update_message"`
	DownloadURL        string   `json:"download_url"`
	WhatsNew           []string `json:"whats_new"`
	ReleaseDate        string   `json:"release_date"`
}

// InAppMessage represents an in-app announcement/message
type InAppMessage struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Title       string `json:"title"`
	Message     string `json:"message"`
	ImageURL    string `json:"image_url,omitempty"`
	ActionText  string `json:"action_text,omitempty"`
	ActionURL   string `json:"action_url,omitempty"`
	Priority    int    `json:"priority"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
	ShowOnce    bool   `json:"show_once"`
	Dismissible bool   `json:"dismissible"`
}

// AppConfig is the complete configuration response
type AppConfig struct {
	AppVersion    AppVersion     `json:"app_version"`
	InAppMessages []InAppMessage `json:"in_app_messages"`
}

// InitDB initializes the SQLite database
func InitDB(dbPath string) error {
	var err error
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}

	// Test connection
	if err = db.Ping(); err != nil {
		return err
	}

	// Create tables
	createTables := `
	CREATE TABLE IF NOT EXISTS app_version (
		id INTEGER PRIMARY KEY CHECK (id = 1),
		latest_version TEXT NOT NULL,
		latest_version_code INTEGER NOT NULL,
		minimum_version_code INTEGER NOT NULL,
		force_update BOOLEAN NOT NULL,
		update_title TEXT NOT NULL,
		update_message TEXT NOT NULL,
		download_url TEXT NOT NULL,
		whats_new TEXT NOT NULL,
		release_date TEXT NOT NULL,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS in_app_messages (
		id TEXT PRIMARY KEY,
		type TEXT NOT NULL,
		title TEXT NOT NULL,
		message TEXT NOT NULL,
		image_url TEXT,
		action_text TEXT,
		action_url TEXT,
		priority INTEGER NOT NULL,
		start_date TEXT NOT NULL,
		end_date TEXT NOT NULL,
		show_once BOOLEAN NOT NULL,
		dismissible BOOLEAN NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`

	if _, err = db.Exec(createTables); err != nil {
		return err
	}

	// Insert default data if empty
	if err = insertDefaultData(); err != nil {
		return err
	}

	log.Println("✅ Database initialized successfully")
	return nil
}

// insertDefaultData inserts default configuration if database is empty
func insertDefaultData() error {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM app_version").Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		// Insert default app version
		whatsNew := []string{
			"✨ Brand new App Config & In-App Message system",
			"💬 Real-time chat with auto-reconnect",
			"🎁 Interactive gift system with native ads",
			"📰 Paper/News section with beautiful UI",
			"📺 Interstitial ads on all button clicks",
			"🐛 Performance improvements and bug fixes",
		}
		whatsNewJSON, _ := json.Marshal(whatsNew)

		_, err = db.Exec(`
			INSERT INTO app_version (
				id, latest_version, latest_version_code, minimum_version_code,
				force_update, update_title, update_message, download_url,
				whats_new, release_date
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			1,
			"1.5.0",
			150,
			1, // Changed from 100 to 1 - allows version 1 to use the app
			false,
			"🎉 New Update Available!",
			"We've added exciting new features and improvements to enhance your experience!",
			"https://github.com/Lainlain/burma2daio/releases/latest",
			string(whatsNewJSON),
			time.Now().Format("2006-01-02"),
		)
		if err != nil {
			return err
		}

		// Insert default messages
		messages := []InAppMessage{
			{
				ID:          "welcome_2025",
				Type:        "info",
				Title:       "🎊 Welcome to Burma 2D 2025!",
				Message:     "Thank you for using our app! We're committed to providing you with the best lottery experience. Check out our new features and enjoy real-time updates!",
				ImageURL:    "",
				ActionText:  "",
				ActionURL:   "",
				Priority:    5,
				StartDate:   "2025-11-01",
				EndDate:     "2025-12-31",
				ShowOnce:    true,
				Dismissible: true,
			},
			{
				ID:          "promo_nov_2025",
				Type:        "promo",
				Title:       "🎁 Special November Offer!",
				Message:     "Get access to premium features this month! Exclusive analysis, predictions, and more. Don't miss out on this limited-time offer!",
				ImageURL:    "",
				ActionText:  "Learn More",
				ActionURL:   "http://localhost:4545",
				Priority:    10,
				StartDate:   "2025-11-01",
				EndDate:     "2025-11-30",
				ShowOnce:    false,
				Dismissible: true,
			},
			{
				ID:          "maintenance_alert",
				Type:        "warning",
				Title:       "⚠️ Scheduled Maintenance",
				Message:     "We'll be performing system maintenance on November 10th from 2:00 AM to 4:00 AM. The app may be temporarily unavailable during this time.",
				ImageURL:    "",
				ActionText:  "",
				ActionURL:   "",
				Priority:    8,
				StartDate:   "2025-11-04",
				EndDate:     "2025-11-10",
				ShowOnce:    false,
				Dismissible: true,
			},
		}

		for _, msg := range messages {
			_, err = db.Exec(`
				INSERT INTO in_app_messages (
					id, type, title, message, image_url, action_text, action_url,
					priority, start_date, end_date, show_once, dismissible
				) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				msg.ID, msg.Type, msg.Title, msg.Message, msg.ImageURL,
				msg.ActionText, msg.ActionURL, msg.Priority, msg.StartDate,
				msg.EndDate, msg.ShowOnce, msg.Dismissible,
			)
			if err != nil {
				return err
			}
		}

		log.Println("✅ Default data inserted")
	}

	return nil
}

// GetAppVersion retrieves the app version from database
func GetAppVersion() (*AppVersion, error) {
	var version AppVersion
	var whatsNewJSON string

	err := db.QueryRow(`
		SELECT latest_version, latest_version_code, minimum_version_code,
		       force_update, update_title, update_message, download_url,
		       whats_new, release_date
		FROM app_version WHERE id = 1
	`).Scan(
		&version.LatestVersion,
		&version.LatestVersionCode,
		&version.MinimumVersionCode,
		&version.ForceUpdate,
		&version.UpdateTitle,
		&version.UpdateMessage,
		&version.DownloadURL,
		&whatsNewJSON,
		&version.ReleaseDate,
	)

	if err != nil {
		return nil, err
	}

	// Parse whats_new JSON
	if err = json.Unmarshal([]byte(whatsNewJSON), &version.WhatsNew); err != nil {
		version.WhatsNew = []string{}
	}

	return &version, nil
}

// UpdateAppVersion updates the app version in database
func UpdateAppVersion(version *AppVersion) error {
	whatsNewJSON, err := json.Marshal(version.WhatsNew)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		UPDATE app_version SET
			latest_version = ?,
			latest_version_code = ?,
			minimum_version_code = ?,
			force_update = ?,
			update_title = ?,
			update_message = ?,
			download_url = ?,
			whats_new = ?,
			release_date = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = 1
	`,
		version.LatestVersion,
		version.LatestVersionCode,
		version.MinimumVersionCode,
		version.ForceUpdate,
		version.UpdateTitle,
		version.UpdateMessage,
		version.DownloadURL,
		string(whatsNewJSON),
		version.ReleaseDate,
	)

	return err
}

// GetInAppMessages retrieves all in-app messages from database
func GetInAppMessages() ([]InAppMessage, error) {
	rows, err := db.Query(`
		SELECT id, type, title, message, image_url, action_text, action_url,
		       priority, start_date, end_date, show_once, dismissible
		FROM in_app_messages
		ORDER BY priority DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []InAppMessage
	for rows.Next() {
		var msg InAppMessage
		var imageURL, actionText, actionURL sql.NullString

		err = rows.Scan(
			&msg.ID, &msg.Type, &msg.Title, &msg.Message,
			&imageURL, &actionText, &actionURL,
			&msg.Priority, &msg.StartDate, &msg.EndDate,
			&msg.ShowOnce, &msg.Dismissible,
		)
		if err != nil {
			return nil, err
		}

		msg.ImageURL = imageURL.String
		msg.ActionText = actionText.String
		msg.ActionURL = actionURL.String

		messages = append(messages, msg)
	}

	return messages, nil
}

// UpsertInAppMessage inserts or updates an in-app message
func UpsertInAppMessage(msg *InAppMessage) error {
	_, err := db.Exec(`
		INSERT INTO in_app_messages (
			id, type, title, message, image_url, action_text, action_url,
			priority, start_date, end_date, show_once, dismissible
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			type = excluded.type,
			title = excluded.title,
			message = excluded.message,
			image_url = excluded.image_url,
			action_text = excluded.action_text,
			action_url = excluded.action_url,
			priority = excluded.priority,
			start_date = excluded.start_date,
			end_date = excluded.end_date,
			show_once = excluded.show_once,
			dismissible = excluded.dismissible,
			updated_at = CURRENT_TIMESTAMP
	`,
		msg.ID, msg.Type, msg.Title, msg.Message, msg.ImageURL,
		msg.ActionText, msg.ActionURL, msg.Priority, msg.StartDate,
		msg.EndDate, msg.ShowOnce, msg.Dismissible,
	)

	return err
}

// DeleteInAppMessage deletes an in-app message by ID
func DeleteInAppMessage(id string) error {
	_, err := db.Exec("DELETE FROM in_app_messages WHERE id = ?", id)
	return err
}

// GetAppConfigHandler returns the current app configuration
func GetAppConfigHandler(c *gin.Context) {
	version, err := GetAppVersion()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get version: " + err.Error()})
		return
	}

	messages, err := GetInAppMessages()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get messages: " + err.Error()})
		return
	}

	config := AppConfig{
		AppVersion:    *version,
		InAppMessages: messages,
	}

	c.JSON(http.StatusOK, config)
	log.Println("✅ App config sent to client")
}

// UpdateAppConfigHandler allows updating the configuration
func UpdateAppConfigHandler(c *gin.Context) {
	var newConfig AppConfig
	if err := c.ShouldBindJSON(&newConfig); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update app version
	if err := UpdateAppVersion(&newConfig.AppVersion); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update version: " + err.Error()})
		return
	}

	// Delete all existing messages and insert new ones
	if _, err := db.Exec("DELETE FROM in_app_messages"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear messages: " + err.Error()})
		return
	}

	// Insert new messages
	for _, msg := range newConfig.InAppMessages {
		if err := UpsertInAppMessage(&msg); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert message: " + err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Configuration updated successfully",
		"config":  newConfig,
	})
	log.Println("✅ App config updated in database")
}

// GetAppVersionHandler returns only version information
func GetAppVersionHandler(c *gin.Context) {
	version, err := GetAppVersion()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, version)
	log.Println("✅ App version info sent")
}

// GetInAppMessagesHandler returns only active messages
func GetInAppMessagesHandler(c *gin.Context) {
	messages, err := GetInAppMessages()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, messages)
	log.Println("✅ In-app messages sent")
}

func main() {
	// Initialize database
	dbPath := "./appconfig.db"
	if err := InitDB(dbPath); err != nil {
		log.Fatal("❌ Failed to initialize database: ", err)
	}
	defer db.Close()

	// Set Gin to release mode
	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()

	// CORS configuration
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// API Routes
	api := r.Group("/api/burma2d")
	{
		// Main config endpoint (used by Android app)
		api.GET("/config", GetAppConfigHandler)

		// Individual endpoints
		api.GET("/version", GetAppVersionHandler)
		api.GET("/messages", GetInAppMessagesHandler)

		// Admin endpoint to update config
		api.POST("/config", UpdateAppConfigHandler)
	}

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":   "healthy",
			"time":     time.Now().Format(time.RFC3339),
			"database": "sqlite3",
		})
	})

	// Root endpoint
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"name":     "Burma 2D - App Config Server",
			"version":  "1.0.0",
			"status":   "running",
			"database": "SQLite3",
			"endpoints": []string{
				"GET  /api/burma2d/config   - Get full app configuration",
				"GET  /api/burma2d/version  - Get version info only",
				"GET  /api/burma2d/messages - Get in-app messages only",
				"POST /api/burma2d/config   - Update configuration (admin)",
				"GET  /health               - Health check",
			},
		})
	})

	port := "8080"
	log.Printf("🚀 Burma 2D App Config Server starting on port %s...\n", port)
	log.Printf("📡 Main endpoint: http://localhost:%s/api/burma2d/config\n", port)
	log.Printf("💾 Database: %s (SQLite3)\n", dbPath)
	log.Printf("🏥 Health check:  http://localhost:%s/health\n", port)
	log.Println("✅ Server ready to serve app configuration!")

	if err := r.Run(":" + port); err != nil {
		log.Fatal("❌ Failed to start server: ", err)
	}
}
