package models

import "time"

type Box struct {
	ID          int64     `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`               // Название
	Slug        string    `json:"slug" db:"slug"`               // URL-идентификатор
	Description string    `json:"description" db:"description"` // Описание
	Rules       string    `json:"rules" db:"rules"`             // Правила участия
	Date        string    `json:"date" db:"date"`               // Дата проведения (DDMMYYYY)
	Time        string    `json:"time" db:"time"`               // Время начала (HH:MM)
	Location    string    `json:"location" db:"location"`       // Место проведения
	Price       int       `json:"price" db:"price"`             // Цена
	Image       string    `json:"image" db:"image"`             // URL обложки
	Status      string    `json:"status" db:"status"`           // active, hidden, draft, processed
	Organizer   string    `json:"organizer" db:"organizer"`     // Организатор
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	CreatedBy   int64     `json:"created_by" db:"created_by"` // ID создателя
}
