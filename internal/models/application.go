package models

import "time"

type ApplicationType string

type ApplicationSource string

type ApplicationStatus string

const (
	ApplicationTypeBox            ApplicationType = "box"
	ApplicationTypeSpecialProject ApplicationType = "special_project"

	ApplicationSourceTelegramBot ApplicationSource = "telegram_bot"
	ApplicationSourceManual      ApplicationSource = "manual"

	ApplicationStatusQueue      ApplicationStatus = "queue"
	ApplicationStatusInProgress ApplicationStatus = "in_progress"
	ApplicationStatusDone       ApplicationStatus = "done"
)

type ApplicationCreateRequest struct {
	Type             ApplicationType   `json:"type"`
	Source           ApplicationSource `json:"source"`
	CustomerName     string            `json:"customer_name"`
	ContactInfo      string            `json:"contact_info"`
	ProjectName      *string           `json:"project_name,omitempty"`
	BoxID            *int64            `json:"box_id,omitempty"`
	SpecialProjectID *int64            `json:"special_project_id,omitempty"`
}

type Application struct {
	ID               int64             `db:"id" json:"id"`
	Type             ApplicationType   `db:"type" json:"type"`
	Source           ApplicationSource `db:"source" json:"source"`
	Status           ApplicationStatus `db:"status" json:"status"`
	CustomerName     string            `db:"customer_name" json:"customer_name"`
	ContactInfo      string            `db:"contact_info" json:"contact_info"`
	ProjectName      *string           `db:"project_name" json:"project_name,omitempty"`
	BoxID            *int64            `db:"box_id" json:"box_id,omitempty"`
	SpecialProjectID *int64            `db:"special_project_id" json:"special_project_id,omitempty"`
	ManagerID        *int64            `db:"manager_id" json:"manager_id,omitempty"`
	CreatedAt        time.Time         `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time         `db:"updated_at" json:"updated_at"`
}

// ApplicationListItem is used in the GET /api/v1/applications response items.
type ApplicationListItem = Application

type ApplicationFilter struct {
	Status    *ApplicationStatus `json:"status,omitempty"`
	Type      *ApplicationType   `json:"type,omitempty"`
	ManagerID *int64             `json:"manager_id,omitempty"`
	DateFrom  *time.Time         `json:"date_from,omitempty"`
	DateTo    *time.Time         `json:"date_to,omitempty"`
	Limit     int                `json:"limit,omitempty"`
	Offset    int                `json:"offset,omitempty"`
}

type Pagination struct {
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}
