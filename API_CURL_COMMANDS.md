# ============================================================================
# THUNDER API - COMPLETE CURL COMMANDS REFERENCE
# Base URL: http://localhost:8080/api/v1
# ============================================================================

# ============================================================================
# 1. TENANT APIs
# ============================================================================

# Verify Tenant Exists
curl -X GET "http://localhost:8080/api/v1/tenants/verify?slug=factory-a" \
  -H "Content-Type: application/json"

# ============================================================================
# 2. AUTH APIs (No token required initially)
# ============================================================================

# Check Auth Path
curl -X POST "http://localhost:8080/api/v1/auth/check" \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "9876543210",
    "tenant_id": "factory-a"
  }'

# Verify Auth (Password/PIN)
curl -X POST "http://localhost:8080/api/v1/auth/verify" \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "9876543210",
    "password": "mypassword123",
    "tenant_id": "factory-a"
  }'

# Login with PIN (Most Common)
curl -X POST "http://localhost:8080/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "9876543210",
    "pin": "1234",
    "tenant_id": "factory-a"
  }'

# Get Role Info
curl -X POST "http://localhost:8080/api/v1/auth/role" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "user_id": "usr_001"
  }'

# Logout
curl -X POST "http://localhost:8080/api/v1/auth/logout" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# ============================================================================
# 3. USER APIs (Requires Bearer Token)
# ============================================================================

# Create Internal User (Admin only)
curl -X POST "http://localhost:8080/api/v1/users" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "name": "John Doe",
    "phone": "9876543211",
    "role": "field_agent",
    "pin": "5678"
  }'

# Update User (Admin only)
curl -X PATCH "http://localhost:8080/api/v1/users/usr_002" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "name": "John Smith",
    "phone": "9876543212",
    "pin": "9999"
  }'

# Delete User (Admin only)
curl -X DELETE "http://localhost:8080/api/v1/users/usr_002" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# ============================================================================
# 4. PRODUCT APIs (Requires Bearer Token)
# ============================================================================

# List Products (All Users)
curl -X GET "http://localhost:8080/api/v1/products?category=beverages&search=sugar&page=1&page_size=20" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# Get Product by ID (All Users)
curl -X GET "http://localhost:8080/api/v1/products/prod_001" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# Create Product (Admin only)
curl -X POST "http://localhost:8080/api/v1/products" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "name": "Orange Juice - 1L",
    "sku": "SKU-002",
    "quantity": 200,
    "category": "beverages",
    "unit_id": 1,
    "price": 45.50,
    "description": "Fresh orange juice",
    "images": [
      {
        "url": "https://example.com/orange-juice.jpg",
        "alt_text": "Orange Juice Bottle"
      }
    ],
    "physical_attributes": {
      "weight": "1000g",
      "dimensions": "10x10x20cm",
      "color": "Orange"
    }
  }'

# Update Product (Admin only)
curl -X PATCH "http://localhost:8080/api/v1/products/prod_002" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "price": 48.99,
    "quantity": 180,
    "is_available": true
  }'

# Delete Product (Admin only)
curl -X DELETE "http://localhost:8080/api/v1/products/prod_002" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# ============================================================================
# 5. ORDER APIs (Requires Bearer Token)
# ============================================================================

# List Orders (Role-based filtering)
curl -X GET "http://localhost:8080/api/v1/orders?status=pending&page=1&page_size=20" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# Get Order by ID
curl -X GET "http://localhost:8080/api/v1/orders/ord_001" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# Create Order (Buyer or Field Agent only)
curl -X POST "http://localhost:8080/api/v1/orders" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "items": [
      {
        "product_id": "prod_001",
        "quantity": 5
      },
      {
        "product_id": "prod_002",
        "quantity": 3
      }
    ],
    "note": "Please deliver by 5 PM"
  }'

# Add Note to Order (Buyer or Field Agent)
curl -X POST "http://localhost:8080/api/v1/orders/ord_001/notes" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "message": "Please include extra packing material"
  }'

# Accept Order (Admin only)
curl -X PATCH "http://localhost:8080/api/v1/orders/ord_001/accept" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "etd": "2024-01-20T14:00:00Z"
  }'

# Reject Order (Admin only)
curl -X PATCH "http://localhost:8080/api/v1/orders/ord_001/reject" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{}'

# Process Order (Staff only)
curl -X PATCH "http://localhost:8080/api/v1/orders/ord_001/process" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{}'

# Ship Order (Staff only)
curl -X PATCH "http://localhost:8080/api/v1/orders/ord_001/ship" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{}'

# Complete Order (Admin only)
curl -X PATCH "http://localhost:8080/api/v1/orders/ord_001/complete" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{}'

# Cancel Order (Buyer or Field Agent only)
curl -X PATCH "http://localhost:8080/api/v1/orders/ord_001/cancel" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{}'

# ============================================================================
# COMPLETE WORKFLOW EXAMPLE
# ============================================================================

# Step 1: Check tenant
curl -X GET "http://localhost:8080/api/v1/tenants/verify?slug=factory-a"

# Step 2: Login
TOKEN=$(curl -s -X POST "http://localhost:8080/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"phone": "9876543210", "pin": "1234", "tenant_id": "factory-a"}' | jq -r '.token')

# Step 3: Create Product
curl -X POST "http://localhost:8080/api/v1/products" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "Coffee Beans - 1kg",
    "sku": "SKU-COFFEE-1KG",
    "quantity": 100,
    "category": "beverages",
    "price": 299.99,
    "description": "Premium arabica coffee beans"
  }'

# Step 4: Place Order
curl -X POST "http://localhost:8080/api/v1/orders" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "items": [{"product_id": "prod_001", "quantity": 2}],
    "note": "Urgent delivery needed"
  }'

# Step 5: List Orders
curl -X GET "http://localhost:8080/api/v1/orders?status=pending" \
  -H "Authorization: Bearer $TOKEN"

# Step 6: Logout
curl -X POST "http://localhost:8080/api/v1/auth/logout" \
  -H "Authorization: Bearer $TOKEN"