package model

type RoleRequest struct {
	Name string `json:"name"`
}

type PermissionRequest struct {
	Name string `json:"name"`
}

type CreateUserRequest struct {
	FirstName            string              `json:"first_name"`
	LastName             string              `json:"last_name"`
	MiddleName           string              `json:"middle_name"`
	Email                string              `json:"email"`
	Password             string              `json:"password"`
	PasswordConfirmation string              `json:"password_confirmation"`
	Phone                string              `json:"phone"`
	Position             string              `json:"position"`
	Active               int                 `json:"active"`
	DataRole             string              `json:"data_role"`
	Roles                []RoleRequest       `json:"roles"`
	Permissions          []PermissionRequest `json:"permissions"`
}
