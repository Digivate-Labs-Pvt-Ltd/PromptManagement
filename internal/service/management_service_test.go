package service

import (
	"context"
	"testing"
	"time"

	"prompt-management/internal/domain"
)

// MockManagementStore is an in-memory implementation of ManagementStore for testing.
type MockManagementStore struct {
	data map[string]*domain.PromptManagement
}

func NewMockManagementStore() *MockManagementStore {
	return &MockManagementStore{
		data: make(map[string]*domain.PromptManagement),
	}
}



func (m *MockManagementStore) Insert(ctx context.Context, pm *domain.PromptManagement) error {
	pm.ID = "generated-uuid-mgmt"
	now := time.Now()
	pm.CreatedAt = now
	pm.UpdatedAt = now
	m.data[pm.ID] = pm
	return nil
}

func (m *MockManagementStore) Update(ctx context.Context, pm *domain.PromptManagement) error {
	if _, ok := m.data[pm.ID]; !ok {
		return domain.ErrNotFound
	}
	pm.UpdatedAt = time.Now()
	m.data[pm.ID] = pm
	return nil
}

func (m *MockManagementStore) GetByID(ctx context.Context, id string) (*domain.PromptManagement, error) {
	pm, ok := m.data[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return pm, nil
}

func (m *MockManagementStore) List(ctx context.Context, filters domain.ListFilters) ([]*domain.PromptManagement, int, error) {
	var result []*domain.PromptManagement
	for _, pm := range m.data {
		match := true
		if filters.Client != "" && pm.Client != filters.Client {
			match = false
		}
		if filters.UseCase != "" && pm.UseCase != filters.UseCase {
			match = false
		}
		if match {
			result = append(result, pm)
		}
	}
	return result, len(result), nil
}

func (m *MockManagementStore) Delete(ctx context.Context, id string) error {
	if _, ok := m.data[id]; !ok {
		return domain.ErrNotFound
	}
	delete(m.data, id)
	return nil
}

func TestManagementService_Create(t *testing.T) {
	store := NewMockManagementStore()
	itemStore := NewMockItemStore()
	svc := NewManagementService(nil, store, itemStore)
	ctx := context.Background()

	userID := "user-123"

	t.Run("Valid Create", func(t *testing.T) {
		req := CreateRequest{
			Client:       "Acme Corp",
			UseCase:      "Customer Support",
			DocumentType: "Email",
			Category:     "Template",
			StageName:    "Initial",
		}
		pm, err := svc.Create(ctx, req, userID)
		if err != nil {
			t.Fatalf("failed to create: %v", err)
		}
		if pm.ID == "" {
			t.Error("expected ID to be generated")
		}
		if pm.Client != req.Client {
			t.Errorf("expected client %s, got %s", req.Client, pm.Client)
		}
	})

	t.Run("Missing Fields", func(t *testing.T) {
		req := CreateRequest{
			Client: "Acme Corp",
		} // missing use_case, etc.
		_, err := svc.Create(ctx, req, userID)
		if err == nil {
			t.Error("expected validation error, got nil")
		}
		if err != domain.ErrValidation {
			t.Errorf("expected %v, got %v", domain.ErrValidation, err)
		}
	})
}

func TestManagementService_GetByID_Aggregated(t *testing.T) {
	store := NewMockManagementStore()
	itemStore := NewMockItemStore()
	svc := NewManagementService(nil, store, itemStore)
	ctx := context.Background()

	// Seed data
	id := "group-1"
	store.data[id] = &domain.PromptManagement{ID: id, Client: "Test"}
	itemStore.items["item-1"] = &domain.PromptItem{ID: "item-1", ManagementID: id, PromptText: "P1"}
	itemStore.items["item-2"] = &domain.PromptItem{ID: "item-2", ManagementID: id, PromptText: "P2"}

	t.Run("Fetch with Items", func(t *testing.T) {
		pm, err := svc.GetByID(ctx, id)
		if err != nil {
			t.Fatalf("failed to get: %v", err)
		}
		if len(pm.Prompts) != 2 {
			t.Errorf("expected 2 prompts, got %d", len(pm.Prompts))
		}
	})
}

func TestManagementService_Update(t *testing.T) {
	store := NewMockManagementStore()
	itemStore := NewMockItemStore()
	svc := NewManagementService(nil, store, itemStore)
	ctx := context.Background()

	// Insert initial
	pm := &domain.PromptManagement{
		ID:           "test-id",
		Client:       "Old Client",
		UseCase:      "Old UseCase",
		DocumentType: "Old Type",
		Category:     "Old Category",
		StageName:    "Old Stage",
	}
	store.data[pm.ID] = pm

	t.Run("Partial Update", func(t *testing.T) {
		req := UpdateRequest{
			ID:      "test-id",
			Client:  "New Client",
			UseCase: "New UseCase",
		}

		updated, err := svc.Update(ctx, req)
		if err != nil {
			t.Fatalf("failed to update: %v", err)
		}

		if updated.Client != "New Client" {
			t.Errorf("expected New Client, got %s", updated.Client)
		}
		if updated.DocumentType != "Old Type" {
			t.Errorf("expected DocumentType to remain unchanged, got %s", updated.DocumentType)
		}
	})

	t.Run("Update Non-existent", func(t *testing.T) {
		req := UpdateRequest{
			ID:     "not-found",
			Client: "New Client",
		}
		_, err := svc.Update(ctx, req)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestManagementService_List(t *testing.T) {
	store := NewMockManagementStore()
	itemStore := NewMockItemStore()
	svc := NewManagementService(nil, store, itemStore)
	ctx := context.Background()

	store.data["1"] = &domain.PromptManagement{ID: "1", Client: "Client A", Category: "X"}
	store.data["2"] = &domain.PromptManagement{ID: "2", Client: "Client B", Category: "Y"}
	store.data["3"] = &domain.PromptManagement{ID: "3", Client: "Client A", Category: "Y"}

	t.Run("Filter by Client", func(t *testing.T) {
		res, total, err := svc.List(ctx, domain.ListFilters{Client: "Client A"})
		if err != nil {
			t.Fatalf("failed to list: %v", err)
		}
		if total != 2 {
			t.Errorf("expected 2 items, got %d", total)
		}
		if len(res) != 2 {
			t.Errorf("expected len 2, got %d", len(res))
		}
	})
}

func TestManagementService_Delete(t *testing.T) {
	store := NewMockManagementStore()
	itemStore := NewMockItemStore()
	svc := NewManagementService(nil, store, itemStore)
	ctx := context.Background()

	store.data["1"] = &domain.PromptManagement{ID: "1"}

	t.Run("Successful Delete", func(t *testing.T) {
		if err := svc.Delete(ctx, "1"); err != nil {
			t.Fatalf("failed to delete: %v", err)
		}
		_, ok := store.data["1"]
		if ok {
			t.Error("expected item to be deleted from store")
		}
	})

	t.Run("Delete Non-existent", func(t *testing.T) {
		err := svc.Delete(ctx, "999")
		if err == nil {
			t.Error("expected error for non-existent item, got nil")
		}
	})
}
