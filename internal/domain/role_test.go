package domain

import (
	"testing"
	"time"

	"invento-service/internal/dto"
)

func TestRoleStruct(t *testing.T) {
	t.Parallel()
	t.Run("Role struct initialization", func(t *testing.T) {
		t.Parallel()
		now := time.Now()
		role := Role{
			ID:        1,
			NamaRole:  "Administrator",
			CreatedAt: now,
			UpdatedAt: now,
		}

		if role.ID != 1 {
			t.Errorf("Expected ID 1, got %d", role.ID)
		}
		if role.NamaRole != "Administrator" {
			t.Errorf("Expected NamaRole 'Administrator', got %s", role.NamaRole)
		}
	})
}

func TestPermissionStruct(t *testing.T) {
	t.Parallel()
	t.Run("Permission struct initialization", func(t *testing.T) {
		t.Parallel()
		now := time.Now()
		perm := Permission{
			ID:        1,
			Resource:  "projects",
			Action:    "create",
			Label:     "Create projects",
			CreatedAt: now,
			UpdatedAt: now,
		}

		if perm.Resource != "projects" {
			t.Errorf("Expected Resource 'projects', got %s", perm.Resource)
		}
		if perm.Action != "create" {
			t.Errorf("Expected Action 'create', got %s", perm.Action)
		}
		if perm.Label != "Create projects" {
			t.Errorf("Expected Label 'Create projects', got %s", perm.Label)
		}
	})
}

func TestRolePermissionStruct(t *testing.T) {
	t.Parallel()
	t.Run("RolePermission struct with relations", func(t *testing.T) {
		t.Parallel()
		now := time.Now()
		role := Role{ID: 1, NamaRole: "Admin"}
		perm := Permission{ID: 1, Resource: "users", Action: "delete"}

		rolePerm := RolePermission{
			ID:           1,
			RoleID:       1,
			PermissionID: 1,
			CreatedAt:    now,
			Role:         role,
			Permission:   perm,
		}

		if rolePerm.RoleID != 1 {
			t.Errorf("Expected RoleID 1, got %d", rolePerm.RoleID)
		}
		if rolePerm.PermissionID != 1 {
			t.Errorf("Expected PermissionID 1, got %d", rolePerm.PermissionID)
		}
		if rolePerm.Role.NamaRole != "Admin" {
			t.Errorf("Expected Role NamaRole 'Admin', got %s", rolePerm.Role.NamaRole)
		}
		if rolePerm.Permission.Resource != "users" {
			t.Errorf("Expected Permission Resource 'users', got %s", rolePerm.Permission.Resource)
		}
	})
}

func TestRoleCreateRequest(t *testing.T) {
	t.Parallel()
	t.Run("dto.RoleCreateRequest with permissions", func(t *testing.T) {
		t.Parallel()
		req := dto.RoleCreateRequest{
			NamaRole: "Editor",
			Permissions: map[string][]string{
				"projects": {"create", "read", "update"},
				"users":    {"read"},
			},
		}

		if req.NamaRole != "Editor" {
			t.Errorf("Expected NamaRole 'Editor', got %s", req.NamaRole)
		}
		if len(req.Permissions) != 2 {
			t.Errorf("Expected 2 permission resources, got %d", len(req.Permissions))
		}
		if len(req.Permissions["projects"]) != 3 {
			t.Errorf("Expected 3 project permissions, got %d", len(req.Permissions["projects"]))
		}
	})

	t.Run("dto.RoleCreateRequest empty permissions", func(t *testing.T) {
		t.Parallel()
		req := dto.RoleCreateRequest{
			NamaRole:    "Viewer",
			Permissions: map[string][]string{},
		}

		if req.NamaRole != "Viewer" {
			t.Errorf("Expected NamaRole 'Viewer', got %s", req.NamaRole)
		}
		if len(req.Permissions) != 0 {
			t.Errorf("Expected 0 permissions, got %d", len(req.Permissions))
		}
	})
}

func TestRoleUpdateRequest(t *testing.T) {
	t.Parallel()
	t.Run("dto.RoleUpdateRequest with updated permissions", func(t *testing.T) {
		t.Parallel()
		req := dto.RoleUpdateRequest{
			NamaRole: "SuperAdmin",
			Permissions: map[string][]string{
				"projects": {"create", "read", "update", "delete"},
				"users":    {"create", "read", "update", "delete"},
				"roles":    {"read"},
			},
		}

		if req.NamaRole != "SuperAdmin" {
			t.Errorf("Expected NamaRole 'SuperAdmin', got %s", req.NamaRole)
		}
		if len(req.Permissions) != 3 {
			t.Errorf("Expected 3 permission resources, got %d", len(req.Permissions))
		}
	})
}

func TestRoleListQueryParams(t *testing.T) {
	t.Parallel()
	t.Run("dto.RoleListQueryParams with all fields", func(t *testing.T) {
		t.Parallel()
		params := dto.RoleListQueryParams{
			Search: "admin",
			Page:   1,
			Limit:  25,
		}

		if params.Search != "admin" {
			t.Errorf("Expected Search 'admin', got %s", params.Search)
		}
		if params.Page != 1 {
			t.Errorf("Expected Page 1, got %d", params.Page)
		}
		if params.Limit != 25 {
			t.Errorf("Expected Limit 25, got %d", params.Limit)
		}
	})

	t.Run("dto.RoleListQueryParams default values", func(t *testing.T) {
		t.Parallel()
		params := dto.RoleListQueryParams{}

		if params.Search != "" {
			t.Errorf("Expected empty Search, got %s", params.Search)
		}
		if params.Page != 0 {
			t.Errorf("Expected Page 0, got %d", params.Page)
		}
	})
}

func TestRoleListItem(t *testing.T) {
	t.Parallel()
	t.Run("dto.RoleListItem struct", func(t *testing.T) {
		t.Parallel()
		now := time.Now()
		item := dto.RoleListItem{
			ID:                1,
			NamaRole:          "Admin",
			JumlahPermission:  15,
			TanggalDiperbarui: now,
		}

		if item.ID != 1 {
			t.Errorf("Expected ID 1, got %d", item.ID)
		}
		if item.NamaRole != "Admin" {
			t.Errorf("Expected NamaRole 'Admin', got %s", item.NamaRole)
		}
		if item.JumlahPermission != 15 {
			t.Errorf("Expected JumlahPermission 15, got %d", item.JumlahPermission)
		}
	})
}

func TestRolePermissionDetail(t *testing.T) {
	t.Parallel()
	t.Run("dto.RolePermissionDetail struct", func(t *testing.T) {
		t.Parallel()
		detail := dto.RolePermissionDetail{
			Resource: "projects",
			Actions:  []string{"create", "read", "update", "delete"},
		}

		if detail.Resource != "projects" {
			t.Errorf("Expected Resource 'projects', got %s", detail.Resource)
		}
		if len(detail.Actions) != 4 {
			t.Errorf("Expected 4 actions, got %d", len(detail.Actions))
		}
	})
}

func TestRoleDetailResponse(t *testing.T) {
	t.Parallel()
	t.Run("dto.RoleDetailResponse with permissions", func(t *testing.T) {
		t.Parallel()
		now := time.Now()
		permissions := []dto.RolePermissionDetail{
			{Resource: "projects", Actions: []string{"create", "read"}},
			{Resource: "users", Actions: []string{"read", "update"}},
		}

		resp := dto.RoleDetailResponse{
			ID:               1,
			NamaRole:         "Editor",
			Permissions:      permissions,
			JumlahPermission: 4,
			CreatedAt:        now,
			UpdatedAt:        now,
		}

		if resp.NamaRole != "Editor" {
			t.Errorf("Expected NamaRole 'Editor', got %s", resp.NamaRole)
		}
		if len(resp.Permissions) != 2 {
			t.Errorf("Expected 2 permission groups, got %d", len(resp.Permissions))
		}
		if resp.JumlahPermission != 4 {
			t.Errorf("Expected JumlahPermission 4, got %d", resp.JumlahPermission)
		}
	})
}

func TestPermissionItem(t *testing.T) {
	t.Parallel()
	t.Run("dto.PermissionItem struct", func(t *testing.T) {
		t.Parallel()
		item := dto.PermissionItem{
			Action: "delete",
			Label:  "Delete resource",
		}

		if item.Action != "delete" {
			t.Errorf("Expected Action 'delete', got %s", item.Action)
		}
		if item.Label != "Delete resource" {
			t.Errorf("Expected Label 'Delete resource', got %s", item.Label)
		}
	})
}

func TestResourcePermissions(t *testing.T) {
	t.Parallel()
	t.Run("dto.ResourcePermissions struct", func(t *testing.T) {
		t.Parallel()
		perms := []dto.PermissionItem{
			{Action: "create", Label: "Create"},
			{Action: "read", Label: "Read"},
			{Action: "update", Label: "Update"},
		}

		res := dto.ResourcePermissions{
			Name:        "projects",
			Permissions: perms,
		}

		if res.Name != "projects" {
			t.Errorf("Expected Name 'projects', got %s", res.Name)
		}
		if len(res.Permissions) != 3 {
			t.Errorf("Expected 3 permissions, got %d", len(res.Permissions))
		}
	})
}

func TestRoleListData(t *testing.T) {
	t.Parallel()
	t.Run("dto.RoleListData with pagination", func(t *testing.T) {
		t.Parallel()
		items := []dto.RoleListItem{
			{ID: 1, NamaRole: "Admin", JumlahPermission: 10},
			{ID: 2, NamaRole: "Editor", JumlahPermission: 5},
		}

		data := dto.RoleListData{
			Items: items,
			Pagination: dto.PaginationData{
				Page:       1,
				Limit:      10,
				TotalItems: 2,
				TotalPages: 1,
			},
		}

		if len(data.Items) != 2 {
			t.Errorf("Expected 2 items, got %d", len(data.Items))
		}
		if data.Pagination.TotalItems != 2 {
			t.Errorf("Expected TotalItems 2, got %d", data.Pagination.TotalItems)
		}
	})
}
