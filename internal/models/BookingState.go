package models

import "time"

const (
	BookingStepSelectDate = 1
	BookingStepGuestName  = 2
	BookingStepOrg        = 3
	BookingStepPosition   = 4
	BookingStepConfirm    = 5
)

type BookingState struct {
	UserID            int64
	ServiceID         int
	VisitType         string // private/public
	SelectedDate      time.Time
	GuestName         string
	GuestOrganization string
	GuestPosition     string
	Step              int
	CreatedAt         time.Time
}