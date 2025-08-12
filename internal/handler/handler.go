package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/senyabanana/shop-service/internal/service"
)

type Handler struct {
	services *service.Service
	log      *logrus.Logger
}

func NewHandler(services *service.Service, log *logrus.Logger) *Handler {
	return &Handler{
		services: services,
		log:      log,
	}
}
func (h *Handler) InitRoutes() *gin.Engine {
	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	api := router.Group("/api")
	{
		api.POST("/auth", h.authenticate)

		protected := api.Group("/", h.userIdentity)
		{
			protected.GET("/info", h.getInfo)
			protected.POST("/sendCoin", h.sendCoin)
			protected.GET("/buy/:item", h.buyItem)
		}
	}

	return router
}
