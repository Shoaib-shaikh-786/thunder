package tenant

import (
    "net/http"

    "github.com/gin-gonic/gin"
)

type Handler struct {
    service *Service
}

func NewHandler(service *Service) *Handler {
    return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
    tenant := r.Group("/tenants")
    {
        tenant.GET("/:id", h.GetTenant)
    }
}

func (h *Handler) GetTenant(c *gin.Context) {
    c.JSON(http.StatusNotImplemented, gin.H{"error": "tenant routes not implemented"})
}
