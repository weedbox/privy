package privy

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestManager(t *testing.T) *Manager {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	storage := NewGormStorage(db)
	m := CreateManager(WithStorage(storage))

	return m
}

func TestManager_CreateResource(t *testing.T) {
	m := setupTestManager(t)

	r, err := m.CreateResource(ResourceConfig{
		Key:         "article",
		Name:        "文章",
		Description: "新聞文章主體",
		Actions: []Action{
			{Key: "read", Name: "閱讀", Description: "閱讀文章內容"},
			{Key: "create", Name: "新增", Description: "建立新文章"},
			{Key: "update", Name: "更新", Description: "編輯既有文章"},
		},
	})

	if err != nil {
		t.Fatalf("failed to create resource: %v", err)
	}

	if r.Key != "article" {
		t.Errorf("expected key 'article', got '%s'", r.Key)
	}

	if len(r.Actions) != 3 {
		t.Errorf("expected 3 actions, got %d", len(r.Actions))
	}
}

func TestManager_CreateResourceWithSubResources(t *testing.T) {
	m := setupTestManager(t)

	r, err := m.CreateResource(ResourceConfig{
		Key:         "article",
		Name:        "文章",
		Description: "新聞文章主體",
		Actions: []Action{
			DefineAction("read", "閱讀", "閱讀文章內容"),
			DefineAction("create", "新增", "建立新文章"),
		},
		SubResources: []Resource{
			{
				Key:         "comment",
				Name:        "留言",
				Description: "文章底下的留言",
				Actions: []Action{
					DefineAction("read", "閱讀留言", "讀取留言內容"),
					DefineAction("create", "新增留言", "新增一則留言"),
				},
			},
		},
	})

	if err != nil {
		t.Fatalf("failed to create resource: %v", err)
	}

	if len(r.SubResources) != 1 {
		t.Errorf("expected 1 sub-resource, got %d", len(r.SubResources))
	}

	if r.SubResources[0].Key != "comment" {
		t.Errorf("expected sub-resource key 'comment', got '%s'", r.SubResources[0].Key)
	}
}

func TestManager_AddActions(t *testing.T) {
	m := setupTestManager(t)

	_, err := m.CreateResource(ResourceConfig{
		Key:         "article",
		Name:        "文章",
		Description: "新聞文章主體",
		Actions: []Action{
			DefineAction("read", "閱讀", "閱讀文章內容"),
		},
	})

	if err != nil {
		t.Fatalf("failed to create resource: %v", err)
	}

	err = m.AddActions("article", []Action{
		DefineAction("share", "分享", "分享文章給其他人"),
		DefineAction("like", "按讚", "對文章按讚"),
	})

	if err != nil {
		t.Fatalf("failed to add actions: %v", err)
	}

	r, err := m.GetResource("article")
	if err != nil {
		t.Fatalf("failed to get resource: %v", err)
	}

	if len(r.Actions) != 3 {
		t.Errorf("expected 3 actions, got %d", len(r.Actions))
	}
}

func TestManager_CreateResources(t *testing.T) {
	m := setupTestManager(t)

	_, err := m.CreateResource(ResourceConfig{
		Key:         "article",
		Name:        "文章",
		Description: "新聞文章主體",
	})

	if err != nil {
		t.Fatalf("failed to create resource: %v", err)
	}

	err = m.CreateResources("article", []Resource{
		{
			Key:         "comment",
			Name:        "留言",
			Description: "文章底下的留言",
			Actions: []Action{
				DefineAction("edit", "編輯留言", "編輯既有留言"),
			},
		},
	})

	if err != nil {
		t.Fatalf("failed to create sub-resources: %v", err)
	}

	r, err := m.GetResource("article")
	if err != nil {
		t.Fatalf("failed to get resource: %v", err)
	}

	if len(r.SubResources) != 1 {
		t.Errorf("expected 1 sub-resource, got %d", len(r.SubResources))
	}
}

func TestManager_GetResourceByPath(t *testing.T) {
	m := setupTestManager(t)

	_, err := m.CreateResource(ResourceConfig{
		Key:         "article",
		Name:        "文章",
		Description: "新聞文章主體",
		SubResources: []Resource{
			{
				Key:         "comment",
				Name:        "留言",
				Description: "文章底下的留言",
			},
		},
	})

	if err != nil {
		t.Fatalf("failed to create resource: %v", err)
	}

	r, err := m.GetResource("article.comment")
	if err != nil {
		t.Fatalf("failed to get resource by path: %v", err)
	}

	if r.Key != "comment" {
		t.Errorf("expected key 'comment', got '%s'", r.Key)
	}
}

func TestManager_CreateRole(t *testing.T) {
	m := setupTestManager(t)

	role, err := m.CreateRole("editor", RoleConfig{
		Name:        "編輯者",
		Description: "可以編輯和發布文章",
		Permissions: []string{
			"article.read",
			"article.create",
			"article.update",
			"article.publish",
		},
	})

	if err != nil {
		t.Fatalf("failed to create role: %v", err)
	}

	if role.Key != "editor" {
		t.Errorf("expected key 'editor', got '%s'", role.Key)
	}

	if len(role.Permissions) != 4 {
		t.Errorf("expected 4 permissions, got %d", len(role.Permissions))
	}
}

func TestManager_AssignPermissions(t *testing.T) {
	m := setupTestManager(t)

	_, err := m.CreateRole("editor", RoleConfig{
		Name:        "編輯者",
		Description: "可以編輯和發布文章",
		Permissions: []string{
			"article.read",
			"article.create",
		},
	})

	if err != nil {
		t.Fatalf("failed to create role: %v", err)
	}

	err = m.AssignPermissions("editor", []string{
		"article.update",
		"article.delete",
	})

	if err != nil {
		t.Fatalf("failed to assign permissions: %v", err)
	}

	role, err := m.GetRole("editor")
	if err != nil {
		t.Fatalf("failed to get role: %v", err)
	}

	if len(role.Permissions) != 4 {
		t.Errorf("expected 4 permissions, got %d", len(role.Permissions))
	}
}

func TestManager_RemovePermissions(t *testing.T) {
	m := setupTestManager(t)

	_, err := m.CreateRole("editor", RoleConfig{
		Name:        "編輯者",
		Description: "可以編輯和發布文章",
		Permissions: []string{
			"article.read",
			"article.create",
			"article.update",
			"article.delete",
		},
	})

	if err != nil {
		t.Fatalf("failed to create role: %v", err)
	}

	err = m.RemovePermissions("editor", []string{
		"article.delete",
	})

	if err != nil {
		t.Fatalf("failed to remove permissions: %v", err)
	}

	role, err := m.GetRole("editor")
	if err != nil {
		t.Fatalf("failed to get role: %v", err)
	}

	if len(role.Permissions) != 3 {
		t.Errorf("expected 3 permissions, got %d", len(role.Permissions))
	}
}

func TestManager_CheckRolePermission(t *testing.T) {
	m := setupTestManager(t)

	_, err := m.CreateRole("editor", RoleConfig{
		Name:        "編輯者",
		Description: "可以編輯和發布文章",
		Permissions: []string{
			"article.read",
			"article.create",
			"article.update",
		},
	})

	if err != nil {
		t.Fatalf("failed to create role: %v", err)
	}

	// Test exact match
	hasPermission, err := m.CheckRolePermission("editor", "article.read")
	if err != nil {
		t.Fatalf("failed to check permission: %v", err)
	}
	if !hasPermission {
		t.Error("expected editor to have 'article.read' permission")
	}

	// Test group match
	hasPermission, err = m.CheckRolePermission("editor", "article")
	if err != nil {
		t.Fatalf("failed to check permission: %v", err)
	}
	if !hasPermission {
		t.Error("expected editor to have 'article' permission")
	}

	// Test no match
	hasPermission, err = m.CheckRolePermission("editor", "article.delete")
	if err != nil {
		t.Fatalf("failed to check permission: %v", err)
	}
	if hasPermission {
		t.Error("expected editor not to have 'article.delete' permission")
	}
}

func TestManager_ListResources(t *testing.T) {
	m := setupTestManager(t)

	_, err := m.CreateResource(ResourceConfig{
		Key:         "article",
		Name:        "文章",
		Description: "新聞文章主體",
	})

	if err != nil {
		t.Fatalf("failed to create resource: %v", err)
	}

	_, err = m.CreateResource(ResourceConfig{
		Key:         "user",
		Name:        "用戶",
		Description: "系統用戶",
	})

	if err != nil {
		t.Fatalf("failed to create resource: %v", err)
	}

	resources, err := m.ListResources()
	if err != nil {
		t.Fatalf("failed to list resources: %v", err)
	}

	if len(resources) != 2 {
		t.Errorf("expected 2 resources, got %d", len(resources))
	}
}

func TestManager_ListRoles(t *testing.T) {
	m := setupTestManager(t)

	_, err := m.CreateRole("editor", RoleConfig{
		Name:        "編輯者",
		Description: "可以編輯和發布文章",
	})

	if err != nil {
		t.Fatalf("failed to create role: %v", err)
	}

	_, err = m.CreateRole("viewer", RoleConfig{
		Name:        "瀏覽者",
		Description: "只能瀏覽文章",
	})

	if err != nil {
		t.Fatalf("failed to create role: %v", err)
	}

	roles, err := m.ListRoles()
	if err != nil {
		t.Fatalf("failed to list roles: %v", err)
	}

	if len(roles) != 2 {
		t.Errorf("expected 2 roles, got %d", len(roles))
	}
}
