package models

type Settings struct {
	Notifications NotificationsSettings `json:"notifications"`
	Booking       BookingSettings       `json:"booking"`
	General       GeneralSettings       `json:"general"`
}

type NotificationsSettings struct {
	TelegramBotToken    string `json:"telegram_bot_token"`
	AutoReminders       bool   `json:"auto_reminders"`
	ReminderHoursBefore int    `json:"reminder_hours_before"`
}

type BookingSettings struct {
	MaxSlotsPerEvent         int  `json:"max_slots_per_event"`
	AllowOverbooking         bool `json:"allow_overbooking"`
	CancellationAllowedHours int  `json:"cancellation_allowed_hours"`
}

type GeneralSettings struct {
	SiteName     string `json:"site_name"`
	ContactEmail string `json:"contact_email"`
	ContactPhone string `json:"contact_phone"`
}
