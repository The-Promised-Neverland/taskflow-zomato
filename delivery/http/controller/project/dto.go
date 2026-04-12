package project_controller

type createRequest struct {
	Name        string `json:"name" validate:"required,min=1"`
	Description string `json:"description"`
}

type updateRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}
