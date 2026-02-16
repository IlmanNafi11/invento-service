package rbac

// RBAC Resources - correspond to Casbin policy objects
const (
	ResourcePermission = "Permission"
	ResourceRole       = "Role"
	ResourceUser       = "User"
	ResourceProject    = "Project"
	ResourceModul      = "Modul"
)

// RBAC Actions - correspond to Casbin policy actions
const (
	ActionRead     = "read"
	ActionCreate   = "create"
	ActionUpdate   = "update"
	ActionDelete   = "delete"
	ActionDownload = "download"
)
