package health

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Handler provides health check endpoints.
type Handler struct {
	db          *gorm.DB
	serviceName string
}

// NewHandler creates a new health handler.
func NewHandler(db *gorm.DB, serviceName string) *Handler {
	return &Handler{db: db, serviceName: serviceName}
}

// RegisterRoutes adds health check routes to the router.
func (h *Handler) RegisterRoutes(r *gin.Engine) {
	r.GET("/health", h.Health)
	r.GET("/readiness", h.Readiness)
}

// Health returns a simple liveness check.
func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": h.serviceName,
	})
}

// Readiness checks if all dependencies are ready.
func (h *Handler) Readiness(c *gin.Context) {
	checks := make(map[string]string)

	// Check database
	sqlDB, err := h.db.DB()
	if err != nil {
		checks["database"] = "error: " + err.Error()
	} else if err := sqlDB.Ping(); err != nil {
		checks["database"] = "error: " + err.Error()
	} else {
		checks["database"] = "ready"
	}

	// Determine overall status
	allReady := true
	for _, status := range checks {
		if status != "ready" {
			allReady = false
			break
		}
	}

	status := http.StatusOK
	overallStatus := "ready"
	if !allReady {
		status = http.StatusServiceUnavailable
		overallStatus = "not_ready"
	}

	c.JSON(status, gin.H{
		"status":  overallStatus,
		"service": h.serviceName,
		"checks":  checks,
	})
}
