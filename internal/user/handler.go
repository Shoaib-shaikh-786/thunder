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
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	auth := r.Group("/auth")
	{
		auth.POST("/login", h.Login)
		auth.POST("/role", h.GetRole)
		auth.POST("/logout", h.Logout)
		auth.POST("/dealer/join", h.DealerJoin) // QR invite registration
	}

	// Wholesaler-only: generate dealer invite QR
	r.POST("/dealers/invite", h.GenerateInvite)

	// Wholesaler-only: manage salesmen and staff
	r.POST("/salesmen", h.CreateSalesman)
	r.DELETE("/salesmen/:id", h.DeleteSalesman)
	r.POST("/staff", h.CreateStaff)
	r.DELETE("/staff/:id", h.DeleteStaff)

	// Dealer update: wholesaler OR the dealer themselves
	r.PATCH("/dealers/:id", h.UpdateDealer)
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

	if claims.Type != string(UserTypeWholesaler) {
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

// POST /salesmen — wholesaler only
func (h *Handler) CreateSalesman(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	if claims.Type != string(UserTypeWholesaler) {
		c.JSON(http.StatusForbidden, gin.H{"error": "wholesalers only"})
		return
	}

	var req CreateSalesmanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.service.CreateSalesman(c.Request.Context(), claims.UserID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// DELETE /salesmen/:id — wholesaler only
func (h *Handler) DeleteSalesman(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	if claims.Type != string(UserTypeWholesaler) {
		c.JSON(http.StatusForbidden, gin.H{"error": "wholesalers only"})
		return
	}

	if err := h.service.DeleteUser(c.Request.Context(), claims.UserID, c.Param("id"), UserTypeSalesman); err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "not found or not yours" {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// POST /staff — wholesaler only
func (h *Handler) CreateStaff(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	if claims.Type != string(UserTypeWholesaler) {
		c.JSON(http.StatusForbidden, gin.H{"error": "wholesalers only"})
		return
	}

	var req CreateStaffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.service.CreateStaff(c.Request.Context(), claims.UserID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// DELETE /staff/:id — wholesaler only
func (h *Handler) DeleteStaff(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}
	if claims.Type != string(UserTypeWholesaler) {
		c.JSON(http.StatusForbidden, gin.H{"error": "wholesalers only"})
		return
	}

	if err := h.service.DeleteUser(c.Request.Context(), claims.UserID, c.Param("id"), UserTypeStaff); err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "not found or not yours" {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// PATCH /dealers/:id — wholesaler (any of their dealers) OR dealer (themselves only)
func (h *Handler) UpdateDealer(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		return
	}

	targetID := c.Param("id")

	// Dealers can only update themselves
	if claims.Type == string(UserTypeDealer) && claims.UserID != targetID {
		c.JSON(http.StatusForbidden, gin.H{"error": "dealers can only update their own profile"})
		return
	}

	// Salesmen/staff cannot update dealers
	if claims.Type != string(UserTypeWholesaler) && claims.Type != string(UserTypeDealer) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	var req UpdateDealerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Phone == nil && req.PIN == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "nothing to update"})
		return
	}

	if err := h.service.UpdateDealer(c.Request.Context(), claims, targetID, req); err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "not found or not yours" {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// POST /auth/role  — public route, no token required
// Request:  { "phone": "8287263475" }
// Response: { "role": "dealer", "wholesaler_name": "Sharma Traders" }
func (h *Handler) GetRole(c *gin.Context) {
	var req RoleLookupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.service.GetRoleByPhone(c.Request.Context(), req)
	if err != nil {
		// Return 404 so frontend knows to show "phone not registered"
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
