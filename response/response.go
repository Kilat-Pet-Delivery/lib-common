package response

import (
	"errors"
	"net/http"

	"github.com/Kilat-Pet-Delivery/lib-common/domain"
	"github.com/gin-gonic/gin"
)

// Success sends a 200 OK response with data.
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}

// Created sends a 201 Created response with data.
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    data,
	})
}

// NoContent sends a 204 No Content response.
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// Paginated sends a paginated response.
func Paginated(c *gin.Context, data interface{}, total int64, page, limit int) {
	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
		"pagination": gin.H{
			"total":       total,
			"page":        page,
			"limit":       limit,
			"total_pages": totalPages,
		},
	})
}

// Error sends an appropriate error response based on the error type.
func Error(c *gin.Context, err error) {
	var domainErr *domain.DomainError
	if errors.As(err, &domainErr) {
		c.JSON(domainErr.Code, gin.H{
			"success": false,
			"error":   domainErr.Message,
			"detail":  domainErr.Detail,
		})
		return
	}

	if errors.Is(err, domain.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "not found",
		})
		return
	}

	if errors.Is(err, domain.ErrUnauthorized) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "unauthorized",
		})
		return
	}

	if errors.Is(err, domain.ErrForbidden) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "forbidden",
		})
		return
	}

	if errors.Is(err, domain.ErrValidation) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Default: internal server error
	c.JSON(http.StatusInternalServerError, gin.H{
		"success": false,
		"error":   "internal server error",
	})
}

// BadRequest sends a 400 response with a message.
func BadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, gin.H{
		"success": false,
		"error":   message,
	})
}
