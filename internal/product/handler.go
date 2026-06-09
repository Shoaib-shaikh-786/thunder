package product

import (
	"net/http"
	"strconv"

	"backend/internal/pkg/auth"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	products := r.Group("/products")
	{
		// READ — all authenticated users
		products.GET("", h.List)
		products.GET("/:id", h.GetByID)

		// WRITE — admin only
		products.POST("", h.Create)
		products.PATCH("/:id", h.Update)
		products.DELETE("/:id", h.Delete)
	}
}

// GET /products?category=beverages&search=sugar&page=1&page_size=20
func (h *Handler) List(c *gin.Context) {
	claims := mustGetClaims(c)
	if claims == nil {
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	filter := ListProductsFilter{
		TenantID: claims.TenantID, // always scoped to the user's tenant
		Category: c.Query("category"),
		Search:   c.Query("search"),
		Page:     page,
		PageSize: pageSize,
	}

	resp, err := h.service.List(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GET /products/:id
func (h *Handler) GetByID(c *gin.Context) {
	claims := mustGetClaims(c)
	if claims == nil {
		return
	}

	p, err := h.service.GetByID(c.Request.Context(), claims.TenantID, c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, p)
}

// POST /products — admin only
func (h *Handler) Create(c *gin.Context) {
	claims := mustGetClaims(c)
	if claims == nil {
		return
	}
	if !isAdmin(c, claims) {
		return
	}

	var req CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	p, err := h.service.Create(c.Request.Context(), claims.TenantID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, p)
}

// PATCH /products/:id — admin only
func (h *Handler) Update(c *gin.Context) {
	claims := mustGetClaims(c)
	if claims == nil {
		return
	}
	if !isAdmin(c, claims) {
		return
	}

	var req UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.Update(c.Request.Context(), claims.TenantID, c.Param("id"), req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "product updated"})
}

// DELETE /products/:id — admin only
func (h *Handler) Delete(c *gin.Context) {
	claims := mustGetClaims(c)
	if claims == nil {
		return
	}
	if !isAdmin(c, claims) {
		return
	}

	if err := h.service.Delete(c.Request.Context(), claims.TenantID, c.Param("id")); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "product deleted"})
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func mustGetClaims(c *gin.Context) *auth.Claims {
	val, exists := c.Get("user_claim")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return nil
	}
	claims, ok := val.(*auth.Claims)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "malformed claims"})
		return nil
	}
	return claims
}

func isAdmin(c *gin.Context, claims *auth.Claims) bool {
	if claims.Type != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "only admins can perform this action"})
		return false
	}
	return true
}
