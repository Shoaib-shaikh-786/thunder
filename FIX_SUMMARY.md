# Quick Fix Summary

## Problem
Login API returned: `ERROR: column "address" does not exist`

## Root Cause
The Go code and database schema were out of sync:
- Go structs used `string` IDs, database used `BIGSERIAL`
- Go used `PinHash` field, database used `password_hash` column
- Go used `Type` field, database used `role` column
- Address field wasn't being properly handled
- Missing `email` and `updated_at` fields

## Solution Applied

### ✅ Files Updated

1. **schema.sql** - Clean schema with DROP statements and all required columns
2. **demoData.sql** - Fixed demo data with proper bcrypt hashes and address handling
3. **internal/user/domain.go** - Fixed struct fields (ID: int64, Role instead of Type, PasswordHash)
4. **internal/user/repository.go** - Updated queries to match schema and domain
5. **internal/user/service.go** - Simplified service logic for new domain
6. **internal/user/handler.go** - Cleaned up handlers and fixed ID parsing
7. **internal/pkg/auth/service.go** - Fixed Claims struct to use int64 UserID

### ✅ Key Changes

| Component | Before | After |
|-----------|--------|-------|
| User ID Type | string (UUID) | int64 (BIGSERIAL) |
| Field Name | `PinHash` | `PasswordHash` |
| Role Field | `Type` | `Role` |
| User Struct | Missing email, updated_at | All fields present |
| Database | IF NOT EXISTS clauses | Clean DROP + CREATE |

## Next Steps

1. **Update your database** with the new schema:
   ```bash
   psql $DATABASE_URL < schema.sql
   psql $DATABASE_URL < demoData.sql
   ```

2. **Test the login endpoint**:
   ```bash
   curl -X POST "http://localhost:8080/api/v1/auth/login" \
     -H "Content-Type: application/json" \
     -d '{"phone": "9876543210", "pin": "1234", "tenant_id": "factory-a"}'
   ```

3. **Expected result**: Get a valid token with user_id as an integer

## Demo Credentials

- **Admin**: 9876543210 / 1234
- **Field Agent**: 9876543211 / 5678  
- **Buyer**: 9876543212 / 9999
- **Staff**: 9876543213 / 4321

All use `tenant_id: factory-a`

---

✅ **All code is now properly synced with the database schema!**
