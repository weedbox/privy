package main

import (
	"fmt"
	"log"

	"github.com/weedbox/privy"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	// Initialize GORM with SQLite
	db, err := gorm.Open(sqlite.Open("rbac.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// Create RBAC manager with GORM storage
	m := privy.CreateManager(
		privy.WithStorage(privy.NewGormStorage(db)),
	)

	fmt.Println("=== Creating Resources ===")

	// Define resources and actions
	r, err := m.CreateResource(privy.ResourceConfig{
		Key:         "article",
		Name:        "Article",
		Description: "News article entity",
		Actions: []privy.Action{
			privy.DefineAction("read", "Read", "Read article content"),
			privy.DefineAction("create", "Create", "Create new article"),
			privy.DefineAction("update", "Update", "Edit existing article"),
			privy.DefineAction("delete", "Delete", "Delete article"),
			privy.DefineAction("publish", "Publish", "Publish article"),
		},
		SubResources: []privy.Resource{
			{
				Key:         "comment",
				Name:        "Comment",
				Description: "Article comments",
				Actions: []privy.Action{
					privy.DefineAction("read", "Read Comment", "Read comment content"),
					privy.DefineAction("create", "Create Comment", "Create a new comment"),
					privy.DefineAction("delete", "Delete Comment", "Delete comment"),
				},
			},
		},
	})
	if err != nil {
		log.Printf("resource may already exist: %v", err)
	} else {
		fmt.Printf("Created resource: %s (%s)\n", r.Key, r.Name)
		fmt.Printf("  Actions: %d\n", len(r.Actions))
		fmt.Printf("  Sub-resources: %d\n", len(r.SubResources))
	}

	fmt.Println("\n=== Extending Resources ===")

	// Add more actions to existing resource
	err = m.AddActions("article", []privy.Action{
		privy.DefineAction("share", "Share", "Share article with others"),
		privy.DefineAction("like", "Like", "Like an article"),
	})
	if err != nil {
		log.Printf("failed to add actions: %v", err)
	} else {
		fmt.Println("Added 'share' and 'like' actions to article resource")
	}

	// Add more sub-resources
	err = m.CreateResources("article", []privy.Resource{
		{
			Key:         "tag",
			Name:        "Tag",
			Description: "Article tags",
			Actions: []privy.Action{
				privy.DefineAction("assign", "Assign Tag", "Assign tag to article"),
			},
		},
	})
	if err != nil {
		log.Printf("failed to create sub-resource: %v", err)
	} else {
		fmt.Println("Added 'tag' sub-resource to article")
	}

	fmt.Println("\n=== Creating Roles ===")

	// Create editor role
	editorRole, err := m.CreateRole("editor", privy.RoleConfig{
		Name:        "Editor",
		Description: "Can edit and publish articles",
		Permissions: []string{
			"article.read",
			"article.create",
			"article.update",
			"article.publish",
			"article.comment.read",
			"article.comment.create",
		},
	})
	if err != nil {
		log.Printf("role may already exist: %v", err)
	} else {
		fmt.Printf("Created role: %s (%s)\n", editorRole.Key, editorRole.Name)
		fmt.Printf("  Permissions: %v\n", editorRole.Permissions)
	}

	// Create viewer role
	viewerRole, err := m.CreateRole("viewer", privy.RoleConfig{
		Name:        "Viewer",
		Description: "Can only view articles",
		Permissions: []string{
			"article.read",
			"article.comment.read",
		},
	})
	if err != nil {
		log.Printf("role may already exist: %v", err)
	} else {
		fmt.Printf("Created role: %s (%s)\n", viewerRole.Key, viewerRole.Name)
		fmt.Printf("  Permissions: %v\n", viewerRole.Permissions)
	}

	fmt.Println("\n=== Assigning Additional Permissions ===")

	// Assign more permissions to editor
	err = m.AssignPermissions("editor", []string{
		"article.delete",
		"article.comment.delete",
	})
	if err != nil {
		log.Printf("failed to assign permissions: %v", err)
	} else {
		fmt.Println("Assigned 'delete' permissions to editor role")
	}

	fmt.Println("\n=== Checking Permissions ===")

	// Check various permission scenarios
	tests := []struct {
		role       string
		permission string
	}{
		{"editor", "article.read"},
		{"editor", "article.delete"},
		{"editor", "article"},
		{"viewer", "article.read"},
		{"viewer", "article.delete"},
		{"viewer", "article"},
	}

	for _, test := range tests {
		hasPermission, err := m.CheckRolePermission(test.role, test.permission)
		if err != nil {
			log.Printf("error checking permission: %v", err)
			continue
		}
		status := "❌"
		if hasPermission {
			status = "✅"
		}
		fmt.Printf("%s Role '%s' has permission '%s': %v\n",
			status, test.role, test.permission, hasPermission)
	}

	fmt.Println("\n=== CheckPermission Function Examples ===")

	permTests := []struct {
		required string
		given    string
	}{
		{"user.create", "user.create"},
		{"user.create", "user"},
		{"infrastructure.vm", "infrastructure.vm.start"},
		{"infrastructure", "infrastructure.vm.stop"},
		{"user.delete", "user.update"},
		{"user.create", "infrastructure.vm.start"},
	}

	for _, test := range permTests {
		result := privy.CheckPermission(test.required, test.given)
		status := "❌"
		if result {
			status = "✅"
		}
		fmt.Printf("%s CheckPermission(%q, %q) = %v\n",
			status, test.required, test.given, result)
	}

	fmt.Println("\n=== Listing Resources ===")

	resources, err := m.ListResources()
	if err != nil {
		log.Printf("failed to list resources: %v", err)
	} else {
		for _, res := range resources {
			fmt.Printf("- %s (%s): %d actions, %d sub-resources\n",
				res.Key, res.Name, len(res.Actions), len(res.SubResources))
		}
	}

	fmt.Println("\n=== Listing Roles ===")

	roles, err := m.ListRoles()
	if err != nil {
		log.Printf("failed to list roles: %v", err)
	} else {
		for _, role := range roles {
			fmt.Printf("- %s (%s): %d permissions\n",
				role.Key, role.Name, len(role.Permissions))
		}
	}

	fmt.Println("\n✅ Example completed successfully!")
}
