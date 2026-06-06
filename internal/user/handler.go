package user

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes mounts all auth + user routes onto the router.
func (h *Handler) RegisterRoutes(r *gin.Engine) {
	auth := r.Group("/auth")
	{
		auth.POST("/login", h.Login)
		auth.POST("/logout", h.Logout)
		auth.POST("/dealer/join", h.DealerJoin) // QR invite registration
	}

	// Wholesaler-only: generate dealer invite QR
	r.POST("/dealers/invite", h.GenerateInvite)
}

// POST /auth/login
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.service.Login(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// POST /auth/logout
func (h *Handler) Logout(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	token := strings.TrimPrefix(authHeader, "Bearer ")

	if err := h.service.Logout(c.Request.Context(), token); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to logout"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

// POST /dealers/invite  — wholesaler only
func (h *Handler) GenerateInvite(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}

	if claims.Type != UserTypeWholesaler {
		c.JSON(http.StatusForbidden, gin.H{"error": "only wholesalers can generate invites"})
		return
	}

	resp, err := h.service.GenerateInvite(c.Request.Context(), claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// POST /auth/dealer/join  — public, uses invite token from QR
func (h *Handler) DealerJoin(c *gin.Context) {
	var req DealerJoinRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.service.DealerJoin(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// getClaims is a helper to extract validated claims from gin context.
func getClaims(c *gin.Context) (*Claims, bool) {
	val, exists := c.Get("user_claim")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return nil, false
	}
	claims, ok := val.(*Claims)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "malformed claims"})
		return nil, false
	}
	return claims, true
}
