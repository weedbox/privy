package privy

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *GormStorage {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	storage := NewGormStorage(db)
	if err := storage.Initialize(); err != nil {
		t.Fatalf("failed to initialize storage: %v", err)
	}

	return storage
}

func TestGormStorage_CreateAndGetResource(t *testing.T) {
	storage := setupTestDB(t)

	resource := &Resource{
		Key:         "article",
		Name:        "文章",
		Description: "新聞文章主體",
	}

	err := storage.CreateResource(resource)
	if err != nil {
		t.Fatalf("failed to create resource: %v", err)
	}

	if resource.ID == 0 {
		t.Error("expected resource ID to be set")
	}

	retrieved, err := storage.GetResource("article", nil)
	if err != nil {
		t.Fatalf("failed to get resource: %v", err)
	}

	if retrieved.Key != "article" {
		t.Errorf("expected key 'article', got '%s'", retrieved.Key)
	}
	if retrieved.Name != "文章" {
		t.Errorf("expected name '文章', got '%s'", retrieved.Name)
	}
}

func TestGormStorage_CreateResourceWithActions(t *testing.T) {
	storage := setupTestDB(t)

	resource := &Resource{
		Key:         "article",
		Name:        "文章",
		Description: "新聞文章主體",
	}

	err := storage.CreateResource(resource)
	if err != nil {
		t.Fatalf("failed to create resource: %v", err)
	}

	actions := []Action{
		{Key: "read", Name: "閱讀", Description: "閱讀文章內容"},
		{Key: "create", Name: "新增", Description: "建立新文章"},
		{Key: "update", Name: "更新", Description: "編輯既有文章"},
	}

	err = storage.CreateActions(resource.ID, actions)
	if err != nil {
		t.Fatalf("failed to create actions: %v", err)
	}

	retrieved, err := storage.GetResource("article", nil)
	if err != nil {
		t.Fatalf("failed to get resource: %v", err)
	}

	if len(retrieved.Actions) != 3 {
		t.Errorf("expected 3 actions, got %d", len(retrieved.Actions))
	}
}

func TestGormStorage_CreateSubResource(t *testing.T) {
	storage := setupTestDB(t)

	parent := &Resource{
		Key:         "article",
		Name:        "文章",
		Description: "新聞文章主體",
	}

	err := storage.CreateResource(parent)
	if err != nil {
		t.Fatalf("failed to create parent resource: %v", err)
	}

	child := &Resource{
		Key:         "comment",
		Name:        "留言",
		Description: "文章底下的留言",
		ParentID:    &parent.ID,
	}

	err = storage.CreateResource(child)
	if err != nil {
		t.Fatalf("failed to create child resource: %v", err)
	}

	retrieved, err := storage.GetResource("comment", &parent.ID)
	if err != nil {
		t.Fatalf("failed to get child resource: %v", err)
	}

	if retrieved.Key != "comment" {
		t.Errorf("expected key 'comment', got '%s'", retrieved.Key)
	}
	if retrieved.ParentID == nil || *retrieved.ParentID != parent.ID {
		t.Errorf("expected parent ID %d, got %v", parent.ID, retrieved.ParentID)
	}
}

func TestGormStorage_ListResources(t *testing.T) {
	storage := setupTestDB(t)

	resources := []*Resource{
		{Key: "article", Name: "文章", Description: "新聞文章主體"},
		{Key: "user", Name: "用戶", Description: "系統用戶"},
	}

	for _, r := range resources {
		if err := storage.CreateResource(r); err != nil {
			t.Fatalf("failed to create resource: %v", err)
		}
	}

	list, err := storage.ListResources(nil)
	if err != nil {
		t.Fatalf("failed to list resources: %v", err)
	}

	if len(list) != 2 {
		t.Errorf("expected 2 resources, got %d", len(list))
	}
}

func TestGormStorage_GetAction(t *testing.T) {
	storage := setupTestDB(t)

	resource := &Resource{
		Key:         "article",
		Name:        "文章",
		Description: "新聞文章主體",
	}

	err := storage.CreateResource(resource)
	if err != nil {
		t.Fatalf("failed to create resource: %v", err)
	}

	actions := []Action{
		{Key: "read", Name: "閱讀", Description: "閱讀文章內容"},
	}

	err = storage.CreateActions(resource.ID, actions)
	if err != nil {
		t.Fatalf("failed to create actions: %v", err)
	}

	action, err := storage.GetAction(resource.ID, "read")
	if err != nil {
		t.Fatalf("failed to get action: %v", err)
	}

	if action.Key != "read" {
		t.Errorf("expected key 'read', got '%s'", action.Key)
	}
	if action.Name != "閱讀" {
		t.Errorf("expected name '閱讀', got '%s'", action.Name)
	}
}

func TestGormStorage_CreateAndGetRole(t *testing.T) {
	storage := setupTestDB(t)

	role := &Role{
		Key:         "editor",
		Name:        "編輯者",
		Description: "可以編輯和發布文章",
		Permissions: []string{"article.read", "article.create", "article.update"},
	}

	err := storage.CreateRole(role)
	if err != nil {
		t.Fatalf("failed to create role: %v", err)
	}

	if role.ID == 0 {
		t.Error("expected role ID to be set")
	}

	retrieved, err := storage.GetRole("editor")
	if err != nil {
		t.Fatalf("failed to get role: %v", err)
	}

	if retrieved.Key != "editor" {
		t.Errorf("expected key 'editor', got '%s'", retrieved.Key)
	}
	if retrieved.Name != "編輯者" {
		t.Errorf("expected name '編輯者', got '%s'", retrieved.Name)
	}
	if len(retrieved.Permissions) != 3 {
		t.Errorf("expected 3 permissions, got %d", len(retrieved.Permissions))
	}
}

func TestGormStorage_UpdateRole(t *testing.T) {
	storage := setupTestDB(t)

	role := &Role{
		Key:         "editor",
		Name:        "編輯者",
		Description: "可以編輯和發布文章",
		Permissions: []string{"article.read", "article.create"},
	}

	err := storage.CreateRole(role)
	if err != nil {
		t.Fatalf("failed to create role: %v", err)
	}

	role.Permissions = append(role.Permissions, "article.update", "article.delete")

	err = storage.UpdateRole(role)
	if err != nil {
		t.Fatalf("failed to update role: %v", err)
	}

	retrieved, err := storage.GetRole("editor")
	if err != nil {
		t.Fatalf("failed to get role: %v", err)
	}

	if len(retrieved.Permissions) != 4 {
		t.Errorf("expected 4 permissions, got %d", len(retrieved.Permissions))
	}
}

func TestGormStorage_ListRoles(t *testing.T) {
	storage := setupTestDB(t)

	roles := []*Role{
		{Key: "editor", Name: "編輯者", Description: "可以編輯和發布文章"},
		{Key: "viewer", Name: "瀏覽者", Description: "只能瀏覽文章"},
	}

	for _, r := range roles {
		if err := storage.CreateRole(r); err != nil {
			t.Fatalf("failed to create role: %v", err)
		}
	}

	list, err := storage.ListRoles()
	if err != nil {
		t.Fatalf("failed to list roles: %v", err)
	}

	if len(list) != 2 {
		t.Errorf("expected 2 roles, got %d", len(list))
	}
}

func TestGormStorage_DeleteResource(t *testing.T) {
	storage := setupTestDB(t)

	resource := &Resource{
		Key:         "article",
		Name:        "文章",
		Description: "新聞文章主體",
	}

	err := storage.CreateResource(resource)
	if err != nil {
		t.Fatalf("failed to create resource: %v", err)
	}

	err = storage.DeleteResource(resource.ID)
	if err != nil {
		t.Fatalf("failed to delete resource: %v", err)
	}

	_, err = storage.GetResourceByID(resource.ID)
	if err != ErrResourceNotFound {
		t.Errorf("expected ErrResourceNotFound, got %v", err)
	}
}

func TestGormStorage_DeleteRole(t *testing.T) {
	storage := setupTestDB(t)

	role := &Role{
		Key:         "editor",
		Name:        "編輯者",
		Description: "可以編輯和發布文章",
	}

	err := storage.CreateRole(role)
	if err != nil {
		t.Fatalf("failed to create role: %v", err)
	}

	err = storage.DeleteRole(role.ID)
	if err != nil {
		t.Fatalf("failed to delete role: %v", err)
	}

	_, err = storage.GetRoleByID(role.ID)
	if err != ErrRoleNotFound {
		t.Errorf("expected ErrRoleNotFound, got %v", err)
	}
}
