package handlers

import (
	"fmt"
	"log"
	"strings"
)

// Button представляет интерактивную кнопку в интерфейсе сообщения.
type Button struct {
	Text         string
	CallbackData string
}

// MessageSender — интерфейс для отправки сообщений пользователю (должен быть внедрён в рантайме бота).
type MessageSender interface {
	SendMessage(userID int64, text string, buttons [][]Button) error
}

// Sender — пакетный отправщик сообщений. Должен быть установлен при инициализации приложения.
// По умолчанию логирует сообщения в stdout для локальной отладки.
var Sender MessageSender = defaultSender{}

// defaultSender реализует MessageSender с простым логированием для разработки/тестирования.
type defaultSender struct{}

func (d defaultSender) SendMessage(userID int64, text string, buttons [][]Button) error {
	log.Printf("SendMessage user=%d\n%s\nButtons:%+v\n", userID, text, buttons)
	return nil
}

// Service содержит данные об услуге.
type Service struct {
	ID          int
	Title       string
	Description string
	Rules       string
	Schedule    string // пустое, если не применяется
	// Options непустой для услуг с несколькими типами посещения (например, галереи)
	Options []string
	// HasBooking указывает, что услуга поддерживает мгновенное бронирование (например, спорт)
	HasBooking bool
}

// ErrServiceNotFound возвращается, когда услуга с указанным ID не найдена.
var ErrServiceNotFound = fmt.Errorf("service not found")

// inMemoryServices содержит примеры услуг и служит источником данных для хендлера.
var inMemoryServices = map[int]Service{
	1: {
		ID:          1,
		Title:       "🎨 Третьяковская галерея",
		Description: "Государственная Третьяковская галерея — крупнейшее собрание русского искусства.",
		Rules:       "Максимум 20 человек. Фото без вспышки.",
		Schedule:    "ПН-СР: 10:00-18:00",
		Options:     []string{"Приватный тур", "Групповой тур"},
		HasBooking:  false,
	},
	2: {
		ID:          2,
		Title:       "🏋️ Спортзал Dynamo",
		Description: "Современный спортивный комплекс с тренажёрным залом и бассейном.",
		Rules:       "Вход по абонементам и предварительной записи.",
		Schedule:    "Ежедневно: 06:00-23:00",
		Options:     nil,
		HasBooking:  true,
	},
}

// buildButtons формирует кнопки ответа в соответствии с настройками услуги.
func buildButtons(s Service) [][]Button {
	var row []Button
	// Для услуг с опциями (например, галереи) — отдельные варианты посещения
	if len(s.Options) > 0 {
		for idx, opt := range s.Options {
			cb := fmt.Sprintf("option:%d:%d", s.ID, idx) // option:<serviceID>:<optionIdx> — формат callback для опции
			row = append(row, Button{Text: opt, CallbackData: cb})
		}
	} else if s.HasBooking {
		// Для услуг с расписанием и возможностью бронирования
		row = append(row, Button{Text: "Забронировать", CallbackData: fmt.Sprintf("book_now:%d", s.ID)})
	}
	// Всегда добавляем кнопку 'Назад'
	row = append(row, Button{Text: "Назад", CallbackData: "back_to_box_solutions"})
	return [][]Button{row}
}

// composeMessage формирует текст сообщения для услуги.
func composeMessage(s Service) string {
	parts := []string{}
	parts = append(parts, s.Title)
	parts = append(parts, "")
	parts = append(parts, "Описание: "+s.Description)
	parts = append(parts, "Правила: "+s.Rules)
	if strings.TrimSpace(s.Schedule) != "" {
		parts = append(parts, "Расписание: "+s.Schedule)
	}
	return strings.Join(parts, "\n")
}

// HandleServiceDetail формирует и отправляет сообщение с деталями услуги указанному пользователю.
// Логирует user_id и service_id и возвращает ошибку, если услуга не найдена или отправка неудачна.
func HandleServiceDetail(serviceID int, userID int64) error {
	log.Printf("HandleServiceDetail called: user_id=%d, service_id=%d", userID, serviceID)
	service, ok := inMemoryServices[serviceID]
	if !ok {
		return ErrServiceNotFound
	}

	msg := composeMessage(service)
	buttons := buildButtons(service)
	// Добавляем подсказку, если у услуги есть опции
	if len(service.Options) > 0 {
		msg += "\n\nВыберите тип посещения:"
	}

	if err := Sender.SendMessage(userID, msg, buttons); err != nil {
		return fmt.Errorf("send message: %w", err)
	}
	return nil
}
