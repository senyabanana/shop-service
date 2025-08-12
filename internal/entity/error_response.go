package entity

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type ErrorResponse struct {
	Message string `json:"errors"`
}

func NewErrorResponse(c *gin.Context, log *logrus.Logger, statusCode int, message string) {
	switch statusCode {
	case http.StatusBadRequest:
		log.Warnf("Bad Request (400): %s", message)
	case http.StatusUnauthorized:
		log.Warnf("Unauthorized access (401): %s", message)
	case http.StatusInternalServerError:
		log.Errorf("Internal server error (500): %s", message)
	default:
		log.Errorf("Unhandled error (%d): %s", statusCode, message)
	}

	c.IndentedJSON(statusCode, ErrorResponse{Message: message})
}
