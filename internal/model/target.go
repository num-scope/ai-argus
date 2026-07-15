package model

import "time"

type Target struct {
	ID                 int64     `gorm:"primaryKey" json:"id"`
	Name               string    `gorm:"size:120;not null;uniqueIndex" json:"name"`
	Protocol           string    `gorm:"size:24;not null" json:"protocol"`
	URL                string    `gorm:"size:1000;not null" json:"url"`
	Model              string    `gorm:"size:200;not null" json:"model"`
	APIKey             string    `gorm:"type:text" json:"-"`
	CustomContentField string    `gorm:"size:120" json:"custom_content_field"`
	ExtraHeadersJSON   string    `gorm:"type:text" json:"-"`
	ExtraBodyJSON      string    `gorm:"type:text" json:"-"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}
