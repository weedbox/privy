package privy

import "testing"

func TestCheckPermission(t *testing.T) {
	tests := []struct {
		name               string
		requiredPermission string
		givenPermission    string
		expected           bool
	}{
		{
			name:               "exact match",
			requiredPermission: "user.create",
			givenPermission:    "user.create",
			expected:           true,
		},
		{
			name:               "group match - given is parent",
			requiredPermission: "user.create",
			givenPermission:    "user",
			expected:           true,
		},
		{
			name:               "hierarchical match - given is child",
			requiredPermission: "infrastructure.vm",
			givenPermission:    "infrastructure.vm.start",
			expected:           true,
		},
		{
			name:               "hierarchical match - given is parent",
			requiredPermission: "infrastructure",
			givenPermission:    "infrastructure.vm.stop",
			expected:           true,
		},
		{
			name:               "different permissions",
			requiredPermission: "user.delete",
			givenPermission:    "user.update",
			expected:           false,
		},
		{
			name:               "different groups",
			requiredPermission: "user.create",
			givenPermission:    "infrastructure.vm.start",
			expected:           false,
		},
		{
			name:               "deep hierarchy match",
			requiredPermission: "infrastructure.vm.compute.start",
			givenPermission:    "infrastructure",
			expected:           true,
		},
		{
			name:               "partial match should fail",
			requiredPermission: "user",
			givenPermission:    "username",
			expected:           false,
		},
		{
			name:               "reverse partial match should fail",
			requiredPermission: "username",
			givenPermission:    "user",
			expected:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CheckPermission(tt.requiredPermission, tt.givenPermission)
			if result != tt.expected {
				t.Errorf("CheckPermission(%q, %q) = %v, want %v",
					tt.requiredPermission, tt.givenPermission, result, tt.expected)
			}
		})
	}
}

func TestCheckPermissions(t *testing.T) {
	tests := []struct {
		name               string
		requiredPermission string
		givenPermissions   []string
		expected           bool
	}{
		{
			name:               "has permission in list",
			requiredPermission: "article.read",
			givenPermissions:   []string{"article.read", "article.create"},
			expected:           true,
		},
		{
			name:               "has parent permission in list",
			requiredPermission: "article.comment.read",
			givenPermissions:   []string{"article", "user.read"},
			expected:           true,
		},
		{
			name:               "no matching permission",
			requiredPermission: "article.delete",
			givenPermissions:   []string{"article.read", "article.create"},
			expected:           false,
		},
		{
			name:               "empty permission list",
			requiredPermission: "article.read",
			givenPermissions:   []string{},
			expected:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CheckPermissions(tt.requiredPermission, tt.givenPermissions)
			if result != tt.expected {
				t.Errorf("CheckPermissions(%q, %v) = %v, want %v",
					tt.requiredPermission, tt.givenPermissions, result, tt.expected)
			}
		})
	}
}

func TestManager_CheckRolesPermission(t *testing.T) {
	m := setupTestManager(t)

	// Create roles
	_, err := m.CreateRole("editor", RoleConfig{
		Name:        "編輯者",
		Description: "可以編輯和發布文章",
		Permissions: []string{"article.read", "article.create", "article.update"},
	})
	if err != nil {
		t.Fatalf("failed to create editor role: %v", err)
	}

	_, err = m.CreateRole("viewer", RoleConfig{
		Name:        "瀏覽者",
		Description: "只能瀏覽文章",
		Permissions: []string{"article.read"},
	})
	if err != nil {
		t.Fatalf("failed to create viewer role: %v", err)
	}

	tests := []struct {
		name               string
		roleKeys           []string
		requiredPermission string
		expected           bool
	}{
		{
			name:               "has permission in first role",
			roleKeys:           []string{"editor", "viewer"},
			requiredPermission: "article.update",
			expected:           true,
		},
		{
			name:               "has permission in second role",
			roleKeys:           []string{"viewer", "editor"},
			requiredPermission: "article.update",
			expected:           true,
		},
		{
			name:               "has permission in both roles",
			roleKeys:           []string{"editor", "viewer"},
			requiredPermission: "article.read",
			expected:           true,
		},
		{
			name:               "no permission in any role",
			roleKeys:           []string{"viewer"},
			requiredPermission: "article.delete",
			expected:           false,
		},
		{
			name:               "non-existent role should be skipped",
			roleKeys:           []string{"nonexistent", "viewer"},
			requiredPermission: "article.read",
			expected:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := m.CheckRolesPermission(tt.roleKeys, tt.requiredPermission)
			if err != nil {
				t.Fatalf("failed to check roles permission: %v", err)
			}
			if result != tt.expected {
				t.Errorf("CheckRolesPermission(%v, %q) = %v, want %v",
					tt.roleKeys, tt.requiredPermission, result, tt.expected)
			}
		})
	}
}
