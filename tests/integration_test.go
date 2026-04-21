package tests

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"prompt-management/internal/config"
	pgrepo "prompt-management/internal/repository/postgres"
	"prompt-management/internal/service"
	"prompt-management/pkg/auth"
)

func TestIntegration_HappyPath(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Spin up PostgreSQL container
	postgresContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:16-alpine"),
		postgres.WithDatabase("prompt_management"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(1).WithStartupTimeout(60*time.Second)),
	)
	if err != nil {
		t.Fatalf("failed to start container: %v", err)
	}

	defer func() {
		if err := postgresContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %v", err)
		}
	}()

	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	// Wait a bit for initialization
	time.Sleep(1 * time.Second)

	// Set up dependencies
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatalf("failed to connect to db: %v", err)
	}
	defer pool.Close()

	// Apply quick manual schema instead of golang-migrate to ease test container logic 
	// Or assuming migrator is available (we create it here just to make sure tests work)
	schema := `
	CREATE TABLE users (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		email VARCHAR(255) UNIQUE NOT NULL,
		username VARCHAR(50) UNIQUE NOT NULL,
		full_name VARCHAR(255) NOT NULL,
		password_hash VARCHAR(255) NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		deleted_at TIMESTAMP WITH TIME ZONE
	);

	CREATE TABLE prompt_management (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		client VARCHAR(100) NOT NULL,
		use_case VARCHAR(100) NOT NULL,
		document_type VARCHAR(100) NOT NULL,
		category VARCHAR(100) NOT NULL,
		stage_name VARCHAR(100) NOT NULL,
		active_item_id UUID,
		created_by UUID REFERENCES users(id),
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		deleted_at TIMESTAMP WITH TIME ZONE
	);

	CREATE TABLE prompt_item (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		management_id UUID REFERENCES prompt_management(id) ON DELETE CASCADE,
		question_key VARCHAR(100) NOT NULL,
		prompt_text TEXT NOT NULL,
		vector_prompt TEXT,
		generation_config JSONB,
		response_schema JSONB,
		top_k NUMERIC,
		ranking_method VARCHAR(50),
		version_no VARCHAR(20) NOT NULL,
		status VARCHAR(20) NOT NULL DEFAULT 'draft',
		change_log TEXT,
		created_by UUID REFERENCES users(id),
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		deleted_at TIMESTAMP WITH TIME ZONE,
		UNIQUE (management_id, question_key, version_no)
	);
	`
	_, err = pool.Exec(ctx, schema)
	if err != nil {
		t.Fatalf("failed to setup schema: %v", err)
	}

	cfg := &config.Config{
		JWTSecret: "test-secret-integration",
	}

	authRepo := pgrepo.NewUserRepository(pool)
	mgmtRepo := pgrepo.NewManagementRepository(pool)
	itemRepo := pgrepo.NewItemRepository(pool)

	authSvc := service.NewAuthService(cfg, authRepo)
	mgmtSvc := service.NewManagementService(pool, mgmtRepo, itemRepo)
	itemSvc := service.NewItemService(itemRepo)

	// Step 1: Register User
	t.Log("Registering user...")
	email := "integration@example.com"
	pass := "secure123"
	user, err := authSvc.Register(ctx, email, "intuser", "Int User", pass)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// Step 2: Login User
	t.Log("Logging in...")
	token, err := authSvc.Login(ctx, email, pass)
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	claims, err := auth.ValidateToken(token, cfg.JWTSecret)
	if err != nil {
		t.Fatalf("Token validation failed: %v", err)
	}
	if claims.UserID != user.ID {
		t.Errorf("UserID mismatch in token")
	}

	// Step 3: Create Prompt Management Group
	t.Log("Creating Management Group...")
	mgmt, err := mgmtSvc.Create(ctx, service.CreateRequest{
		Client:       "IntClient",
		UseCase:      "IntUseCase",
		DocumentType: "PDF",
		Category:     "Legal",
		StageName:    "Extraction",
	}, user.ID)
	if err != nil {
		t.Fatalf("Create mgmt failed: %v", err)
	}

	// Step 4: Add New Draft Item
	t.Log("Adding draft item (v1.0.0)...")
	item1, err := itemSvc.Add(ctx, service.AddItemRequest{
		ManagementID: mgmt.ID,
		QuestionKey:  "field1",
		PromptText:   "Extract name",
	}, user.ID)
	if err != nil {
		t.Fatalf("Add item failed: %v", err)
	}
	if item1.Status != "draft" || item1.VersionNo != "v1.0.0" {
		t.Errorf("Invalid item1 state: %s, %s", item1.Status, item1.VersionNo)
	}

	// Step 5: Promote Item
	t.Log("Promoting item...")
	if err := itemSvc.Promote(ctx, mgmt.ID, item1.ID); err != nil {
		t.Fatalf("Promote failed: %v", err)
	}

	// Verify status update
	item1db, _ := itemSvc.GetByID(ctx, item1.ID)
	if item1db.Status != "active" {
		t.Errorf("expected item1 active, got %s", item1db.Status)
	}

	// Step 6: Add updated item draft
	t.Log("Adding updated draft item (v1.0.1)...")
	item2, err := itemSvc.Add(ctx, service.AddItemRequest{
		ManagementID: mgmt.ID,
		QuestionKey:  "field1",
		PromptText:   "Extract complete legal name",
	}, user.ID)
	if err != nil {
		t.Fatalf("Add item2 failed: %v", err)
	}
	if item2.VersionNo != "v1.0.1" {
		t.Errorf("expected item2 version v1.0.1, got %s", item2.VersionNo)
	}

	// Step 7: Promote new item
	t.Log("Promoting item2...")
	if err := itemSvc.Promote(ctx, mgmt.ID, item2.ID); err != nil {
		t.Fatalf("Promote item2 failed: %v", err)
	}

	// Verify final state: item1 archived, item2 active
	item1Final, _ := itemSvc.GetByID(ctx, item1.ID)
	item2Final, _ := itemSvc.GetByID(ctx, item2.ID)

	if item1Final.Status != "archived" {
		t.Errorf("expected item1 archived, got %s", item1Final.Status)
	}
	if item2Final.Status != "active" {
		t.Errorf("expected item2 active, got %s", item2Final.Status)
	}
}

func TestIntegration_BulkCreateAndAggregatedFetch(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Spin up PostgreSQL container
	postgresContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:16-alpine"),
		postgres.WithDatabase("prompt_management_bulk"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(1).WithStartupTimeout(60*time.Second)),
	)
	if err != nil {
		t.Fatalf("failed to start container: %v", err)
	}
	defer postgresContainer.Terminate(ctx)

	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}
	time.Sleep(2 * time.Second)

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatalf("failed to connect to db: %v", err)
	}
	defer pool.Close()

	// Setup Schema
	schema := `
	CREATE TABLE users (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		email VARCHAR(255) UNIQUE NOT NULL,
		username VARCHAR(50) UNIQUE NOT NULL,
		full_name VARCHAR(255) NOT NULL,
		password_hash VARCHAR(255) NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		deleted_at TIMESTAMP WITH TIME ZONE
	);

	CREATE TABLE prompt_management (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		client VARCHAR(100) NOT NULL,
		use_case VARCHAR(100) NOT NULL,
		document_type VARCHAR(100) NOT NULL,
		category VARCHAR(100) NOT NULL,
		stage_name VARCHAR(100) NOT NULL,
		active_item_id UUID,
		created_by UUID REFERENCES users(id),
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		deleted_at TIMESTAMP WITH TIME ZONE
	);

	CREATE TABLE prompt_item (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		management_id UUID REFERENCES prompt_management(id) ON DELETE CASCADE,
		question_key VARCHAR(100) NOT NULL,
		prompt_text TEXT NOT NULL,
		vector_prompt TEXT,
		generation_config JSONB,
		response_schema JSONB,
		top_k NUMERIC,
		ranking_method VARCHAR(50),
		version_no VARCHAR(20) NOT NULL,
		status VARCHAR(20) NOT NULL DEFAULT 'draft',
		change_log TEXT,
		created_by UUID REFERENCES users(id),
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		deleted_at TIMESTAMP WITH TIME ZONE,
		UNIQUE (management_id, question_key, version_no)
	);
	`
	_, err = pool.Exec(ctx, schema)
	if err != nil {
		t.Fatalf("failed to setup schema: %v", err)
	}

	cfg := &config.Config{JWTSecret: "test-bulk-secret"}
	authRepo := pgrepo.NewUserRepository(pool)
	mgmtRepo := pgrepo.NewManagementRepository(pool)
	itemRepo := pgrepo.NewItemRepository(pool)

	authSvc := service.NewAuthService(cfg, authRepo)
	mgmtSvc := service.NewManagementService(pool, mgmtRepo, itemRepo)

	// Register & Setup User
	user, _ := authSvc.Register(ctx, "bulk@example.com", "bulkuser", "Bulk User", "pass123")

	// Step 1: Perform Bulk Creation
	t.Log("Testing Bulk Creation...")
	bulkReq := service.BulkCreateRequest{
		CreateRequest: service.CreateRequest{
			Client:       "BulkClient",
			UseCase:      "BulkUseCase",
			DocumentType: "JSON",
			Category:     "Dev",
			StageName:    "Alpha",
		},
		Items: []service.AddItemRequest{
			{
				QuestionKey: "key1", 
				PromptText: "Prompt 1",
				ChangeLog: func(s string) *string { return &s }("Bulk init log"),
				TopK: func(f float64) *float64 { return &f }(5.0),
				RankingMethod: func(s string) *string { return &s }("cosine"),
			},
			{QuestionKey: "key2", PromptText: "Prompt 2"},
		},
	}

	pm, err := mgmtSvc.CreateFull(ctx, bulkReq, user.ID)
	if err != nil {
		t.Fatalf("Bulk CreateFull failed: %v", err)
	}

	if pm.Client != "BulkClient" || len(pm.Prompts) != 2 {
		t.Errorf("Bulk creation verification failed. Items: %d", len(pm.Prompts))
	}

	// Verify auto-promotion and field persistence
	for _, p := range pm.Prompts {
		if p.Status != "active" || p.VersionNo != "v1.0.0" {
			t.Errorf("Auto-promotion failed for %s: %s (%s)", p.QuestionKey, p.Status, p.VersionNo)
		}
		if p.QuestionKey == "key1" {
			if p.ChangeLog == nil || *p.ChangeLog != "Bulk init log" {
				t.Errorf("ChangeLog persistence failed: expected 'Bulk init log', got %v", p.ChangeLog)
			}
			if p.TopK == nil || *p.TopK != 5.0 {
				t.Errorf("TopK persistence failed: expected 5.0, got %v", p.TopK)
			}
		}
	}

	// Step 2: Test Aggregated Fetch (GetByID)
	t.Log("Testing Aggregated Fetch...")
	fetched, err := mgmtSvc.GetByID(ctx, pm.ID)
	if err != nil {
		t.Fatalf("Aggregated GetByID failed: %v", err)
	}

	if len(fetched.Prompts) != 2 {
		t.Errorf("Aggregated fetch failed to return correct item count: %d", len(fetched.Prompts))
	}

	// Ensure prompts are linked back
	if fetched.Prompts[0].ManagementID != pm.ID {
		t.Errorf("Prompt linking failed in aggregated fetch")
	}
}
