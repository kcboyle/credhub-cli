package models

type SecretBody struct {
	ContentType string            `json:"type" binding:"required"`
	Value       interface{}       `json:"value,omitempty"`
	Parameters  *SecretParameters `json:"parameters,omitempty"`
	UpdatedAt   string            `json:"updated_at,omitempty"`
}
