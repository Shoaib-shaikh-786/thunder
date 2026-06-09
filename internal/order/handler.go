package order

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
	orders := r.Group("/orders")
	{
		// Read — all roles
		orders.GET("", h.List)
		orders.GET("/:id", h.GetByID)

		// Create — buyer and field_agent
		orders.POST("", h.Create)

		// Notes — buyer or field_agent only
		orders.POST("/:id/notes", h.AddNote)

		// Status transitions — role enforced inside handler
		orders.PATCH("/:id/accept", h.Accept)     // admin
		orders.PATCH("/:id/reject", h.Reject)     // admin
		orders.PATCH("/:id/complete", h.Complete) // admin
		orders.PATCH("/:id/process", h.Process)   // staff
		orders.PATCH("/:id/ship", h.Ship)         // staff
		orders.PATCH("/:id/cancel", h.Cancel)     // buyer or field_agent
	}
}

// POST /orders
func (h *Handler) Create(c *gin.Context) {
	claims := mustGetClaims(c)
	if claims == nil {
		return
	}

	// Only buyer and field_agent can place orders
	if claims.Type != "buyer" && claims.Type != "field_agent" {
		c.JSON(http.StatusForbidden, gin.H{"error": "only buyers and field agents can place orders"})
		return
	}

	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := strconv.FormatInt(claims.UserID, 10)
	o, err := h.service.CreateOrder(
		c.Request.Context(),
		claims.TenantID,
		userID,
		claims.Type,
		req,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, o)
}

// GET /orders?status=pending&page=1&page_size=20
func (h *Handler) List(c *gin.Context) {
	claims := mustGetClaims(c)
	if claims == nil {
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	filter := ListOrdersFilter{
		TenantID:   claims.TenantID,
		Status:     OrderStatus(c.Query("status")),
		Page:       page,
		PageSize:   pageSize,
	}

	userID := strconv.FormatInt(claims.UserID, 10)

	// Buyers see their own orders; field agents see their own placed orders.
	switch claims.Type {
	case "buyer":
		filter.BuyerID = userID
	case "field_agent":
		filter.PlacedByID = userID
	}
	// admin and staff see all orders

	resp, err := h.service.List(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GET /orders/:id
func (h *Handler) GetByID(c *gin.Context) {
	claims := mustGetClaims(c)
	if claims == nil {
		return
	}

	o, err := h.service.GetByID(c.Request.Context(), claims.TenantID, c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, o)
}

// PATCH /orders/:id/accept — admin only
func (h *Handler) Accept(c *gin.Context) {
	claims := mustGetClaims(c)
	if claims == nil {
		return
	}
	if !requireRole(c, claims, "admin") {
		return
	}

	var req AcceptOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.AcceptOrder(c.Request.Context(), claims.TenantID, c.Param("id"), req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "order accepted"})
}

// PATCH /orders/:id/reject — admin only
func (h *Handler) Reject(c *gin.Context) {
	claims := mustGetClaims(c)
	if claims == nil {
		return
	}
	if !requireRole(c, claims, "admin") {
		return
	}

	if err := h.service.RejectOrder(c.Request.Context(), claims.TenantID, c.Param("id")); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "order rejected"})
}

// PATCH /orders/:id/complete — admin only
func (h *Handler) Complete(c *gin.Context) {
	claims := mustGetClaims(c)
	if claims == nil {
		return
	}
	if !requireRole(c, claims, "admin") {
		return
	}

	if err := h.service.CompleteOrder(c.Request.Context(), claims.TenantID, c.Param("id")); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "order completed"})
}

// PATCH /orders/:id/process — staff only
func (h *Handler) Process(c *gin.Context) {
	claims := mustGetClaims(c)
	if claims == nil {
		return
	}
	if !requireRole(c, claims, "staff") {
		return
	}

	if err := h.service.ProcessOrder(c.Request.Context(), claims.TenantID, c.Param("id")); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "order processing"})
}

// PATCH /orders/:id/ship — staff only
func (h *Handler) Ship(c *gin.Context) {
	claims := mustGetClaims(c)
	if claims == nil {
		return
	}
	if !requireRole(c, claims, "staff") {
		return
	}

	if err := h.service.ShipOrder(c.Request.Context(), claims.TenantID, c.Param("id")); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "order shipped"})
}

// PATCH /orders/:id/cancel — buyer or field_agent only
func (h *Handler) Cancel(c *gin.Context) {
	claims := mustGetClaims(c)
	if claims == nil {
		return
	}
	if claims.Type != "buyer" && claims.Type != "field_agent" {
		c.JSON(http.StatusForbidden, gin.H{"error": "only buyers or field agents can cancel orders"})
		return
	}

 	userID := strconv.FormatInt(claims.UserID, 10)
	if err := h.service.CancelOrder(c.Request.Context(), claims.TenantID, userID, c.Param("id")); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "order cancelled"})
}

// POST /orders/:id/notes — buyer or field_agent only
func (h *Handler) AddNote(c *gin.Context) {
	claims := mustGetClaims(c)
	if claims == nil {
		return
	}
	if claims.Type != "buyer" && claims.Type != "field_agent" {
		c.JSON(http.StatusForbidden, gin.H{"error": "only buyers or field agents can add notes"})
		return
	}

	var req AddNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := strconv.FormatInt(claims.UserID, 10)
	if err := h.service.AddNote(c.Request.Context(), claims.TenantID, userID, claims.Type, c.Param("id"), req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "note added"})
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

func requireRole(c *gin.Context, claims *auth.Claims, role string) bool {
	if claims.Type != role {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied for role: " + claims.Type})
		return false
	}
	return true
}
