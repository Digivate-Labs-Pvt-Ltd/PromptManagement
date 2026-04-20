package service

import (
	"context"
	"testing"
	"time"

	"prompt-management/internal/domain"
)

// MockItemStore is an in-memory implementation of ItemStore for testing.
type MockItemStore struct {
	items map[string]*domain.PromptItem
}

func NewMockItemStore() *MockItemStore {
	return &MockItemStore{
		items: make(map[string]*domain.PromptItem),
	}
}

func (m *MockItemStore) Insert(ctx context.Context, item *domain.PromptItem) error {
	item.ID = "generated-uuid-item"
	item.CreatedAt = time.Now()
	item.UpdatedAt = time.Now()
	m.items[item.ID] = item
	return nil
}

func (m *MockItemStore) GetByID(ctx context.Context, id string) (*domain.PromptItem, error) {
	item, ok := m.items[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return item, nil
}

func (m *MockItemStore) Promote(ctx context.Context, managementID string, itemID string) error {
	// First, verify item exists and is draft
	item, ok := m.items[itemID]
	if !ok {
		return domain.ErrNotFound
	}
	if item.ManagementID != managementID {
		return domain.ErrValidation
	}

	// Archive any existing active item for this managementID
	for _, i := range m.items {
		if i.ManagementID == managementID && i.Status == "active" && i.ID != itemID {
			i.Status = "archived"
			i.UpdatedAt = time.Now()
		}
	}

	// Promote the requested item
	item.Status = "active"
	item.UpdatedAt = time.Now()
	return nil
}

func (m *MockItemStore) Archive(ctx context.Context, id string) error {
	item, ok := m.items[id]
	if !ok {
		return domain.ErrNotFound
	}
	item.Status = "archived"
	item.UpdatedAt = time.Now()
	return nil
}

func (m *MockItemStore) List(ctx context.Context, filters domain.ItemListFilters) ([]*domain.PromptItem, int, error) {
	var result []*domain.PromptItem
	for _, item := range m.items {
		if item.ManagementID == filters.ManagementID {
			if filters.Status == nil || item.Status == *filters.Status {
				result = append(result, item)
			}
		}
	}
	return result, len(result), nil
}

func (m *MockItemStore) GetLatestVersionByGroupAndKey(ctx context.Context, managementID, questionKey string) (*string, error) {
	var latest *string
	for _, item := range m.items {
		if item.ManagementID == managementID && item.QuestionKey == questionKey {
			// For testing we just return the first one found or we could do proper sorting
			if latest == nil || item.VersionNo > *latest {
				v := item.VersionNo
				latest = &v
			}
		}
	}
	return latest, nil
}

func TestItemService_Add(t *testing.T) {
	store := NewMockItemStore()
	svc := NewItemService(store)
	ctx := context.Background()

	userID := "user-123"

	t.Run("Valid Add Item (First Draft)", func(t *testing.T) {
		req := AddItemRequest{
			ManagementID: "mgmt-1",
			QuestionKey:  "q1",
			PromptText:   "Write a test",
		}
		item, err := svc.Add(ctx, req, userID)
		if err != nil {
			t.Fatalf("failed to add item: %v", err)
		}
		if item.VersionNo != "v1.0.0" {
			t.Errorf("expected version v1.0.0, got %s", item.VersionNo)
		}
		if item.Status != "draft" {
			t.Errorf("expected status draft, got %s", item.Status)
		}
	})

	t.Run("Missing Validation Fields", func(t *testing.T) {
		req := AddItemRequest{
			ManagementID: "mgmt-1",
		} // missing prompt text and question key
		_, err := svc.Add(ctx, req, userID)
		if err == nil {
			t.Error("expected validation error, got nil")
		}
	})

	t.Run("Add Subsequent Item", func(t *testing.T) {
		req := AddItemRequest{
			ManagementID: "mgmt-1",
			QuestionKey:  "q1",
			PromptText:   "Write a better test",
		}
		item, err := svc.Add(ctx, req, userID)
		if err != nil {
			t.Fatalf("failed to add item: %v", err)
		}
		if item.VersionNo != "v1.0.1" {
			t.Errorf("expected bumped version v1.0.1, got %s", item.VersionNo)
		}
	})
}

func TestBumpVersion(t *testing.T) {
	tests := []struct {
		input    *string
		expected string
	}{
		{nil, "v1.0.0"},
		{stringPtr(""), "v1.0.0"},
		{stringPtr("v1.0.0"), "v1.0.1"},
		{stringPtr("v2.4.9"), "v2.4.10"},
		{stringPtr("1.0.5"), "v1.0.6"},
		{stringPtr("invalid"), "invalid-1"},
	}

	for _, tt := range tests {
		actual := bumpVersion(tt.input)
		if actual != tt.expected {
			t.Errorf("bumpVersion(%v) expected %s, got %s", deref(tt.input), tt.expected, actual)
		}
	}
}

func TestItemService_Promote(t *testing.T) {
	store := NewMockItemStore()
	svc := NewItemService(store)
	ctx := context.Background()

	// Insert active item
	store.items["item-1"] = &domain.PromptItem{ID: "item-1", ManagementID: "mgmt-1", Status: "active"}
	// Insert new draft item
	store.items["item-2"] = &domain.PromptItem{ID: "item-2", ManagementID: "mgmt-1", Status: "draft"}

	t.Run("Successful Promote", func(t *testing.T) {
		if err := svc.Promote(ctx, "mgmt-1", "item-2"); err != nil {
			t.Fatalf("failed to promote: %v", err)
		}
		if store.items["item-1"].Status != "archived" {
			t.Errorf("expected item-1 to be archived, got %s", store.items["item-1"].Status)
		}
		if store.items["item-2"].Status != "active" {
			t.Errorf("expected item-2 to be active, got %s", store.items["item-2"].Status)
		}
	})

	t.Run("Validation Error", func(t *testing.T) {
		if err := svc.Promote(ctx, "", "item-2"); err == nil {
			t.Error("expected validation error for empty mgmt id, got nil")
		}
	})
}

func TestItemService_ListAndArchive(t *testing.T) {
	store := NewMockItemStore()
	svc := NewItemService(store)
	ctx := context.Background()

	store.items["1"] = &domain.PromptItem{ID: "1", ManagementID: "mgmt-1", Status: "draft"}

	t.Run("List Items", func(t *testing.T) {
		items, _, err := svc.List(ctx, domain.ItemListFilters{ManagementID: "mgmt-1"})
		if err != nil {
			t.Fatalf("failed to list: %v", err)
		}
		if len(items) != 1 {
			t.Errorf("expected 1 item, got %d", len(items))
		}
	})

	t.Run("Archive Item", func(t *testing.T) {
		if err := svc.Archive(ctx, "1"); err != nil {
			t.Fatalf("failed to archive: %v", err)
		}
		if store.items["1"].Status != "archived" {
			t.Errorf("expected status to be archived, got %s", store.items["1"].Status)
		}
	})
}

func stringPtr(s string) *string {
	return &s
}

func deref(s *string) string {
	if s == nil {
		return "nil"
	}
	return *s
}
