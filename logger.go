package pkgcommon

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gofrs/uuid/v5"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type LogData struct {
	CarID  string
	UserID string
	Data   any
}

// loggerOptions contains all options for logging
type loggerOptions struct {
	LogFor      string
	Error       error
	Data        datatypes.JSONType[LogData]
	ErrorDetail string
	DB          *gorm.DB
}

// NewLogger creates a new logger instance with required options
func NewLogger(db *gorm.DB, logFor string, err error, errorDetail string) *loggerOptions {
	return &loggerOptions{
		LogFor:      logFor,
		ErrorDetail: errorDetail,
		Error:       err,
		DB:          db,
	}
}

// WithData adds logData information to the logger
func (l *loggerOptions) WithData(logData LogData) *loggerOptions {
	l.Data = datatypes.NewJSONType(logData)
	return l
}

// Log executes the logging process
func (l *loggerOptions) Log() error {
	if l.DB == nil {
		panic("Log Error - DB is not initialized")
	}

	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in Logger: %v", r)
		}
	}()

	logEntry := map[string]interface{}{
		"id":           uuid.Must(uuid.NewV7()).String(),
		"error":        fmt.Sprintf("Error: %v", l.Error),
		"detail":       l.ErrorDetail,
		"app_env":      os.Getenv("APP_ENV"),
		"service_name": os.Getenv("APP_NAME"),
		"created_at":   time.Now().UTC(),
	}

	// Handle production alerts
	if os.Getenv("APP_ENV") == "prod" && l.Data.Data().Data != nil {
		alert := ServiceAlert{
			Service:        os.Getenv("APP_NAME"),
			Detail:         logEntry["detail"].(string),
			FallbackDetail: logEntry["error"].(string),
			Title:          l.LogFor,
			TopicName:      "services-alerts",
			Data:           l.Data,
			CarID:          l.Data.Data().CarID,
			UserID:         l.Data.Data().UserID,
		}

		if err := ServiceAlertNotification(alert); err != nil {
			log.Printf("Failed to send service alert: %v", err)
		}
	}

	// Log to database
	if err := l.DB.Table("error_logs").Create(&logEntry).Error; err != nil {
		return fmt.Errorf("failed to create error log: %w", err)
	}

	return nil
}
