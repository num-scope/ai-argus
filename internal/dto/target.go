package dto

import "time"

type TargetRequest struct {
	Name               string `form:"name" json:"name"`
	Protocol           string `form:"protocol" json:"protocol"`
	URL                string `form:"url" json:"url"`
	Model              string `form:"model" json:"model"`
	APIKey             string `form:"api_key" json:"api_key"`
	CustomContentField string `form:"custom_content_field" json:"custom_content_field"`
	ExtraHeadersJSON   string `form:"extra_headers_json" json:"extra_headers_json"`
	ExtraBodyJSON      string `form:"extra_body_json" json:"extra_body_json"`
}

type TargetResponse struct {
	ID                 int64     `json:"id"`
	Name               string    `json:"name"`
	Protocol           string    `json:"protocol"`
	URL                string    `json:"url"`
	Model              string    `json:"model"`
	HasAPIKey          bool      `json:"has_api_key"`
	CustomContentField string    `json:"custom_content_field,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}
