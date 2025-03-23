package model

type UpdateUserRequest struct {
	FirstName  string  `json:"first_name"`
	LastName   string  `json:"last_name"`
	MiddleName *string `json:"middle_name,omitempty"`
	Phone      *string `json:"phone,omitempty"`
}
