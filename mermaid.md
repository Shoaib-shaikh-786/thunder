LOGIN FLOW 

1. High-Level Design (HLD): System Context & Data Routing				
graph TD
    %% Core Entities & Entry Points
    User([User: Admin / Staff / Agent / Buyer]) -->|Visits sub.yourapp.com| FrontEnd[Responsive Web UI / Capacitor Wrapper]
    
    %% API Gateway & Middleware Layer
    FrontEnd -->|HTTPS API Requests| Gateway[Go Backend Server / API Router]
    
    subgraph Go Backend Architecture
        Gateway -->|Step 1: Pre-flight Handshake| TenantCheck{Is Slug Active?}
        TenantCheck -->|No| Err404[Return 404 / Block Frontend]
        TenantCheck -->|Yes| ServeBranding[Inject Branding Config to UI]
        
        Gateway -->|Step 2: Subsequent Core API Calls| AuthMiddleware{JWT Middleware Validation}
        AuthMiddleware -->|Token Verification Failed| Denied[401/403 Forbidden]
        AuthMiddleware -->|Token Decoded: Extracts Tenant Context| InjectContext[Set Global Query Scope / Tenant Context]
    end


    %% Data Isolation Layer
    subgraph Data Tier
        InjectContext -->|Enforces strict logical isolation| SharedDB[(Single Shared PostgreSQL Instance)]
        SharedDB --> RLS_Engine{Row-Level Security / SQL Scopes}
        RLS_Engine -->|WHERE tenant_id = 'abc'| IsolatedData[(Tenant Specific View)]
    end

    %% Formatting Elements
    style User fill:#f9f,stroke:#333,stroke-width:2px
    style SharedDB fill:#bbf,stroke:#333,stroke-width:2px
    style IsolatedData fill:#bfb,stroke:#333,stroke-width:2px





2. Low-Level Design (LLD): Future-Proof Two-Step Auth Sequence


sequenceDiagram
    autonumber
    actor User as Mobile / Web Client
    participant FE as Frontend (App/Web)
    participant BE as Go Backend API
    participant DB as Shared Database

    %% Handshake Phase
    User->>FE: Navigates to tenant-slug.yourapp.com
    FE->>BE: GET /api/v1/tenants/verify?slug=tenant-slug
    BE->>DB: SELECT * FROM tenants WHERE slug = 'tenant-slug' AND is_active = true
    DB-->>BE: Tenant Metadata (Name, Logo, Colors, TenantType)
    BE-->>FE: 200 OK + Tenant JSON Metadata
    FE->>User: Dynamically renders branded UI with Mobile Number Input Box

    %% Step 1: Identifier Check
    User->>FE: Enters Mobile Number (+919876543210) & hits Next
    FE->>BE: POST /api/v1/auth/check { phone: "...", tenant_id: "..." }
    BE->>DB: SELECT role FROM users WHERE phone = '...' AND tenant_id = '...'
    DB-->>BE: User Profile Found (e.g., Role: field_agent)
    Note over BE: Today: Returns "password"<br/>Future: Check role, trigger SMS gateway & return "otp"
    BE-->>FE: 200 OK { status: "allowed", auth_method: "password" }
    FE->>User: Hides phone entry, slides in Password Input Box smoothly

    %% Step 2: Verification Phase
    User->>FE: Types Password & clicks Login
    FE->>BE: POST /api/v1/auth/verify { phone, tenant_id, password }
    BE->>BE: Compare Bcrypt Password Hash
    BE-->>FE: 200 OK + Signed JWT Token (Payload: UserID, TenantID, Role)
    FE->>User: Persists Token inside LocalStorage & unlocks Dashboard Dashboard




3. Low-Level Design (LLD): Database Schema & Entity Relationships


erDiagram
    TENANTS {
        string id PK "Unique identifier / Subdomain Slug (e.g. 'factory-a')"
        string name "Official Corporate Name"
        string business_type "factory | corporate"
        jsonb branding_config "Holds JSON map of logos, primary/secondary hex colors"
        timestamp created_at
    }

    USERS {
        uint id PK
        string tenant_id FK "Composite Index: idx_tenant_phone"
        string name "User Full Name"
        string phone UK "Composite Index: idx_tenant_phone (Unique within tenant only)"
        string email "Optional text field"
        string password_hash "Encrypted secure string"
        string role "admin | buyer | field_agent | staff"
        boolean is_active "Soft kill flag"
        jsonb metadata "Dynamic storage (e.g., Assigned routes, vehicle IDs, MoQs)"
    }

    PRODUCTS {
        uint id PK
        string tenant_id FK "Indexed for query scoping"
        string name "Product Title"
        string sku "Stock Keeping Unit number"
        decimal standard_price "Base price value"
        jsonb custom_attributes "Dynamic data (e.g., Batch numbers, Expiries, MOQs)"
    }

    ORDERS {
        uint id PK
        string tenant_id FK
        uint buyer_id FK "Points to USERS where role = buyer"
        uint agent_id FK "Points to USERS where role = field_agent"
        string status "Engine State: pending | processing | shipped | completed"
        decimal total_amount
        timestamp created_at
    }

    INVOICES {
        uint id PK
        string tenant_id FK
        uint order_id FK "One-to-One assignment to related order string"
        string invoice_number "Generated sequential string sequence"
        decimal tax_amount "Multi-tax breakdown value"
        decimal grand_total
        timestamp generated_at
    }

    MARKET_KNOWLEDGE {
        uint id PK
        string tenant_id FK
        uint agent_id FK "Points to field agent logging data"
        string intelligence_type "competitor_pricing | stock_alert | general_feedback"
        jsonb logs "Flexible key-value payload (Geo-locations, image URLs, competitor metrics)"
        timestamp logged_at
    }

    %% Relationship Rules
    TENANTS ||--o{ USERS : "contains"
    TENANTS ||--o{ PRODUCTS : "owns"
    TENANTS ||--o{ ORDERS : "manages"
    TENANTS ||--o{ INVOICES : "issues"
    TENANTS ||--o{ MARKET_KNOWLEDGE : "collects"
    
    USERS ||--o{ ORDERS : "places / captures"
    ORDERS ||--|| INVOICES : "generates"
    USERS ||--o{ MARKET_KNOWLEDGE : "reports"





1. Master Request Lifecycle & Architecture (High-Level Design)


graph TD
    %% Entry Point
    User([User Mobile/Web Client]) -->|1. Opens abc.yourapp.com| FE[Frontend App / Wrapper]
    
    %% Handshake & Login Cycle
    FE -->|2. Pre-flight Check| BE_Handshake[GET /api/v1/tenants/verify]
    BE_Handshake -->|3. Validate Slug| DB[(Single Shared Database)]
    DB -->|4. Return Branding/Config| FE
    
    FE -->|5. POST /api/v1/auth/verify| BE_Login[Auth Controller]
    BE_Login -->|6. Verify Credentials| DB
    BE_Login -->|7. Issues Signed JWT Token| FE
    
    %% Frontend Token Storage Strategy
    subgraph Frontend Architecture
        FE -->|8a. Save Raw Token| LocalStorage[(LocalStorage / Secure Storage)]
        FE -->|8b. Decode Base64 Payload| JWTDecode[jwt-decode Engine]
        JWTDecode -->|8c. Hydrate Memory State| AppState[Global State: userPermissions Array]
        AppState -->|8d. Render/Hide UI Elements| DOM[DOM Layout / Sidebar Buttons]
    end

    %% Outgoing Secure API Flow
    LocalStorage -->|9. Inject Bearer Token into Header| AxiosInterceptor[HTTP Request Client]
    AxiosInterceptor -->|10. Secure API Call with Auth Header| GIN[Go Backend API Gateway]

    %% Backend Middleware Pipeline
    subgraph Go Backend Middleware Chain
        GIN -->|Checkpoint A| MW_JWT{1. JWT Auth Middleware}
        MW_JWT -->|Invalid Signature| Err401[401 Unauthorized]
        MW_JWT -->|Valid: Extracts Tenant & Perms to Context| MW_Tenant{2. Tenant URL Cross-Check}
        
        MW_Tenant -->|Token Tenant != URL Domain| Err403_A[403 Forbidden]
        MW_Tenant -->|Valid Match| MW_Perms{3. Permission Middleware}
        
        MW_Perms -->|Missing Required Key| Err403_B[403 Forbidden]
        MW_Perms -->|Authorized| Controller[Core Business Controller]
    end

    %% Scoped Query Execution
    Controller -->|11. Inject c.GetString-tenant_id| ScopedQuery[SQL Context Scope]
    ScopedQuery -->|12. SELECT WHERE tenant_id = x| DB

    %% Styling
    style User fill:#f9f,stroke:#333,stroke-width:2px
    style DB fill:#bbf,stroke:#333,stroke-width:2px
    style LocalStorage fill:#ffb,stroke:#333,stroke-width:2px




End-to-End Sequence Diagram (Low-Level Design)


sequenceDiagram
    autonumber
    actor User as Ground User (Agent/Buyer)
    participant FE as Frontend UI & Storage
    participant BE_MW as Go Middleware Chain
    participant BE_CTRL as Go Business Controller
    participant DB as Central PostgreSQL

    %% PHASE 1: Pre-Flight Check
    User->>FE: Enters 'abc.yourapp.com'
    FE->>BE_MW: GET /api/v1/tenants/verify?slug=abc
    BE_MW->>DB: SELECT * FROM tenants WHERE slug = 'abc'
    DB-->>BE_MW: Found: tenant_id=abc_99, color=#1E3A8A, type=factory
    BE_MW-->>FE: 200 OK + Branding JSON
    Note over FE: Frontend dynamic style update:<br/>Injects company logo & sets primary hex colors

    %% PHASE 2: Two-Step Authentication
    User->>FE: Inputs Mobile Number
    FE->>BE_MW: POST /api/v1/auth/check { phone: "987...", tenant_id: "abc_99" }
    BE_MW->>DB: SELECT role FROM users WHERE phone='987...' AND tenant_id='abc_99'
    DB-->>BE_MW: Match found: role = field_agent
    BE_MW-->>FE: 200 OK { auth_method: "password" }
    FE->>User: Renders Password entry screen
    User->>FE: Inputs Password & clicks Login
    FE->>BE_MW: POST /api/v1/auth/verify { phone, password, tenant_id }
    BE_MW->>BE_MW: Bcrypt hash validation comparison
    BE_MW-->>FE: 200 OK + { token: "eyJhbGci..." }

    %% PHASE 3: Frontend Hydration
    Note over FE: Frontend Execution Loop:<br/>1. localStorage.setItem("token", token)<br/>2. payload = decode(token)<br/>3. appState.permissions = payload.permissions
    FE->>User: Displays Workspace Dashboard (Sidebar items hidden/shown dynamically)

    %% PHASE 4: Executing a Secured Request
    User->>FE: Clicks "Submit New Order"
    Note over FE: Axios Interceptor fetches raw token from localStorage<br/>and attaches it as 'Authorization: Bearer eyJhbGci...'
    FE->>BE_MW: POST /api/v1/orders { buyer_id: 12, items: [...] }
    
    Note over BE_MW: Middleware Checkpoint 1 (JWT Verification):<br/>Decodes Token Signature, extracts tenant_id & permissions array.<br/>Stores values inside Request Context context via c.Set()
    Note over BE_MW: Middleware Checkpoint 2 (Tenant Cross-Check):<br/>Verifies Context tenant_id matches requested host domain
    Note over BE_MW: Middleware Checkpoint 3 (Permission Check):<br/>Confirms context.permissions contains "order:create" string
    
    BE_MW->>BE_CTRL: Passes verified Request Context control down line
    BE_CTRL->>DB: INSERT INTO orders (tenant_id, buyer_id) VALUES (context.tenant_id, 12)
    DB-->>BE_CTRL: Write Successful Confirmation
    BE_CTRL-->>FE: 201 Created { status: "Order Synced" }
    FE->>User: Shows success checkmark confirmation screen





3. Frontend Token Parsing & UI Guard Architecture (LLD)


graph TD
    subgraph Token Distribution Node
        LoginResponse[Login Response JWT String] -->|Saved Intact| LocalStorage[(LocalStorage Cache)]
        LoginResponse -->|Passed Through| Decoder[Base64 JWT Decoder Engine]
    end

    subgraph Memory Space (State Hydration)
        Decoder -->|Extract Claims| UnpackedJSON[JSON Claims Object]
        UnpackedJSON -->|Extract tenant_id| TenantState[Global Tenant Context]
        UnpackedJSON -->|Extract permissions| PermsState[Global Permissions Array Context]
    end

    subgraph Execution Layers (Visual Access Gatekeepers)
        %% Layer A: Router Guards
        PermsState -->|Evaluates Route Requests| RouterGuard{Frontend Router Guard}
        RouterGuard -->|Array contains 'invoice:read'| AcceptRoute[Mount /billing Page View]
        RouterGuard -->|Array missing key| DenyRoute[Redirect to /unauthorized Screen]

        %% Layer B: DOM Engine Conditional Blocks
        PermsState -->|Evaluates Component Elements| CoreUI[Active Page DOM Template]
        CoreUI -->|permissions.includes 'product:create'| RenderBtn[Render 'Add Product' Button]
        CoreUI -->|permissions.includes 'market:log'| RenderSidebar[Render 'Market Insights' Sidebar Link]
    end

    %% Raw Outbound Pipeline
    LocalStorage -->|Attached automatically via Axios request interceptor| HTTPHeader[HTTP Authorization Header: Bearer String]
    HTTPHeader -->|Dispatched to network wire| GoBackend((Go Server Boundary))

    %% Styling
    style LocalStorage fill:#ffb,stroke:#333,stroke-width:1px
    style PermsState fill:#bfb,stroke:#333,stroke-width:2px
    style GoBackend fill:#fbb,stroke:#333,stroke-width:2px




Master Platform Blueprint (Mermaid Code)




graph TD
   %% Global Class Styles Configuration
   classDef entry fill:#eceff1,stroke:#37474f,stroke-width:2px;
   classDef page fill:#e0f2fe,stroke:#0369a1,stroke-width:2px;
   classDef api fill:#fef08a,stroke:#a16207,stroke-width:2px;
   classDef middleware fill:#ffe4e6,stroke:#9f1239,stroke-width:2px;
   classDef db fill:#dcfce7,stroke:#166534,stroke-width:2px;


   %% ==========================================
   %% LAYER 1: INITIAL ENTRY & AUTHENTICATION HANDSHAKE
   %% ==========================================
   subgraph Layer_1_Gateway_And_Authentication_Flow ["1. Gateway & Authentication Flow"]
       A[User visits: domain.yourapp.com] -->|Extracts sub-domain string| B(Next.js Hook initialization)
       B -->|Trigger Handshake API Call| API_Verify["GET /api/v1/tenants/verify?slug=abc"]
      
       API_Verify -->|Read tenant validation| DB_Tenants[(tenants table)]
       DB_Tenants -->|Return Branding: Colors, Logos, Type| B
      
       B -->|Dynamic UI Skinning| C[Mobile Number Form Entry]
       C -->|Trigger Step 1 API Call| API_Check["POST /api/v1/auth/check {phone, tenant_id}"]
      
       API_Check -->|Verify Role & Configuration| DB_Users[(users table)]
       DB_Users -->|Return auth_method: password / otp| C
      
       C -->|Step 2 Credentials Entry| D[Password Input Field]
       D -->|Trigger Verification API Call| API_Verify_Auth["POST /api/v1/auth/verify {phone, password, tenant_id}"]
      
       API_Verify_Auth -->|Bcrypt Check| DB_Users
       API_Verify_Auth -->|Issue Crypto Key| E[Return Signed JWT Token]
   end


   %% ==========================================
   %% LAYER 2: NEXT.JS RUNTIME CONTEXT & STATE DECODER
   %% ==========================================
   subgraph Layer_2_Frontend_State_Context ["2. Frontend State Context"]
       E -->|Store Intact token| LS[(Browser LocalStorage / Capacitor SecureStorage)]
       E -->|Unpack public metadata| Dec[jwt-decode Engine]
       Dec -->|Hydrate Context Wrapper| SessionState[Global useAuth Context: 'session' Object]
      
       SessionState -->|Inject Context| Guard{shadcn Navigation Guard}
   end


   %% ==========================================
   %% LAYER 3: CORE APPLICATION PAGES & EXPLICIT API CALLS
   %% ==========================================
   subgraph Layer_3_Application_Workspace_Pages ["3. Application Workspace Pages & Role Actions"]
      
       %% --- SCREEN 1: DASHBOARD ---
       subgraph Page_Dashboard ["Page: Dashboard Home"]
           Guard -->|Authenticated Session| P1_DOM[Render Dashboard View]
           P1_DOM -->|Admin / Staff Views Overview Metrics| API_Dash_Admin["GET /api/v1/dashboard/summary"]
           P1_DOM -->|Field Agent Tracks Targets| API_Dash_Agent["GET /api/v1/dashboard/agent-targets"]
           P1_DOM -->|Buyer Views Status Summary| API_Dash_Buyer["GET /api/v1/dashboard/buyer-summary"]
       end


       %% --- SCREEN 2: PRODUCTS ---
       subgraph Page_Products ["Page: Product Catalog Management"]
           Guard -->|Requires permission: product:read| P2_DOM[Render Product Interface]
          
           P2_DOM -->|Everyone reads scoped price tier| API_Prod_Get["GET /api/v1/products"]
           P2_DOM -->|Admin Creates New Catalog Items| API_Prod_Post["POST /api/v1/products"]
           P2_DOM -->|Staff/Admin Updates Inventory Counts| API_Prod_Put["PUT /api/v1/products/:id"]
          
           %% Cart Interactions
           P2_DOM -->|Field Agent / Buyer Add Actions| LocalCart[(LocalStorage Cart Array Object)]
       end


       %% --- SCREEN 3: ORDERS ---
       subgraph Page_Orders ["Page: Order Pipeline Operations"]
           Guard -->|Requires permission: order:read| P3_DOM[Render Order System]
          
           LocalCart -->|Click Checkout / Submit Order| API_Order_Post["POST /api/v1/orders {buyer_id, items: [...]}"]
           API_Order_Post -->|On 201 Success Response| ClearCart[Wipe LocalStorage Cart to Empty Array]
          
           P3_DOM -->|Admin Reads Company-wide History| API_Order_Admin["GET /api/v1/orders?scope=all"]
           P3_DOM -->|Staff Updates Kanban Status Board| API_Order_Patch["PATCH /api/v1/orders/:id/status {status}"]
           P3_DOM -->|Agent Reads Self-Booked Logs Only| API_Order_Agent["GET /api/v1/orders (Backend uses secure context matching)"]
           P3_DOM -->|Buyer Reads Personal Orders Only| API_Order_Buyer["GET /api/v1/orders (Backend scopes via decoded token ID)"]
       end


       %% --- SCREEN 4: MARKET INSIGHTS ---
       subgraph Page_Market ["Page: Sales-Market Knowledge"]
           Guard -->|Filters out role: buyer| P4_DOM[Render Intel Hub]
          
           P4_DOM -->|Field Agent Logs Ground Rumors| API_Market_Post["POST /api/v1/market-intelligence {buyer_id, type, notes, geo}"]
           P4_DOM -->|Admin Views Feed & Route Map Proofs| API_Market_Get["GET /api/v1/market-intelligence"]
       end


       %% --- SCREEN 5: BILLING ---
       subgraph Page_Billing ["Page: Invoices & Document Vault"]
           Guard -->|Requires permission: invoice:read| P5_DOM[Render Billing Archive]
          
           P5_DOM -->|Staff / Admin Initiates Document| API_Invoice_Post["POST /api/v1/invoices/:order_id"]
           P5_DOM -->|Everyone Downloads Secure PDF file| API_Invoice_Get["GET /api/v1/invoices/:id/download"]
       end
   end


   %% ==========================================
   %% LAYER 4: BACKEND GATEWAY MIDDLEWARE SECURITY PIPELINE
   %% ==========================================
   subgraph Layer_4_Go_Secure_Middleware_Pipeline ["4. Go Backend Middleware Security Pipeline"]
       LS -->|Automatically added to every request header via Axios| Header_Token[Authorization: Bearer Token String]
      
       %% API endpoints route traffic through the header token mapping
       API_Dash_Admin & API_Dash_Agent & API_Dash_Buyer & API_Prod_Get & API_Prod_Post & API_Prod_Put & API_Order_Post & API_Order_Admin & API_Order_Patch & API_Order_Agent & API_Order_Buyer & API_Market_Post & API_Market_Get & API_Invoice_Post & API_Invoice_Get --> Header_Token
      
       Header_Token --> GIN_Router[Go Backend Gin Engine Router]
      
       GIN_Router --> MW_JWT{1. JWT Checkpoint}
       MW_JWT -->|Fails crypto validation signature| Res_401[Abort: 403 Unauthorized]
       MW_JWT -->|Passes: Saves variables into request context via c.Set| MW_Domain{2. Domain Cross-Check Checkpoint}
      
       MW_Domain -->|token tenant_id != URL host domain| Res_403A[Abort: 403 Forbidden Domain Mismatch]
       MW_Domain -->|Passes| MW_Perms{3. Permission Enforcement Checkpoint}
      
       MW_Perms -->|Context permissions array lacks key| Res_403B[Abort: 403 Forbidden Privileges Missing]
       MW_Perms -->|Passes| Core_Controllers[Target Core Operational Controller]
   end


   %% ==========================================
   %% LAYER 5: DATA ISOLATION ENGINE
   %% ==========================================
   subgraph Layer_5_Database_Storage_Layer ["5. Shared Database Multi-Tenant Isolation Engine"]
       Core_Controllers -->|Injects pre-verified context constraints| SQL_Scope[Enforced Scoped SQL Query Runner]
      
       SQL_Scope -->|WHERE tenant_id = context.tenant_id AND buyer_id = context.user_id| System_DB[(Single Production PostgreSQL Database)]
      
       System_DB --> DB_Products[(products table)]
       System_DB --> DB_Orders[(orders table)]
       System_DB --> DB_Invoices[(invoices table)]
       System_DB --> DB_Market[(market_knowledge table)]
   end


   %% Assign Node Classes for Visual Clarity
   class A,B,C,D,E,LS,Dec,SessionState,Guard entry;
   class P1_DOM,P2_DOM,P3_DOM,P4_DOM,P5_DOM,LocalCart page;
   class API_Verify,API_Check,API_Verify_Auth,API_Dash_Admin,API_Dash_Agent,API_Dash_Buyer,API_Prod_Get,API_Prod_Post,API_Prod_Put,API_Order_Post,API_Order_Admin,API_Order_Patch,API_Order_Agent,API_Order_Buyer,API_Market_Post,API_Market_Get,API_Invoice_Post,API_Invoice_Get api;
   class Header_Token,GIN_Router,MW_JWT,MW_Domain,MW_Perms,Core_Controllers,SQL_Scope middleware;
   class System_DB,DB_Tenants,DB_Users,DB_Products,DB_Orders,DB_Invoices,DB_Market db;




