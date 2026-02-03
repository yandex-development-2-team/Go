package handlers

import (
	"fmt"
	"log"
)

// ServiceType — тип услуги
type ServiceType int

const (
	ServiceTypeOther ServiceType = iota
	ServiceTypeGallery
	ServiceTypeMuseum
	ServiceTypeSport
)

// Service — минимальная модель услуги, используемая обработчиком
type Service struct {
	ID          int
	Name        string
	Description string
	Rules       string
	Schedule    string // опционально
	Type        ServiceType
}

// services — in-memory каталог, используемый обработчиком. Заменить репозиторием в реальном приложении.
var services = map[int]Service{
	1: {ID: 1, Name: "Третьяковская галерея", Description: "Государственная Третьяковская галерея...", Rules: "Максимум 20 человек...", Schedule: "ПН-СР: 10:00-18:00", Type: ServiceTypeGallery},
	2: {ID: 2, Name: "Футбольный зал", Description: "Зал для спортивных тренировок.", Rules: "Возраст 6+", Schedule: "ВС-ПТ: 08:00-22:00", Type: ServiceTypeSport},
	3: {ID: 3, Name: "Обычная услуга", Description: "Описание услуги.", Rules: "Общие правила.", Type: ServiceTypeOther},
	4: {ID: 4, Name: "Музей истории", Description: "Музей с экспонатами.", Rules: "Не трогать экспонаты.", Type: ServiceTypeMuseum},
}

// Button — модель кнопки с callback-данными
type Button struct {
	Text     string
	Callback string
}

// SenderFunc используется для отправки сообщений; заменить реальным отправщиком Telegram при интеграции.
var Sender func(userID int64, text string, buttons [][]Button) error = func(userID int64, text string, buttons [][]Button) error {
	// По умолчанию no-op отправщик — логирует полезную нагрузку для локальной/разработческой среды
	log.Printf("SendMessage user_id=%d text=%q buttons=%v", userID, text, buttons)
	return nil
}

// HandleServiceDetail формирует и отправляет сообщение с деталями услуги пользователю.
// Логирует user_id и service_id. Возвращает ошибку, если услуга не найдена
// или при ошибке отправки.
func HandleServiceDetail(serviceID int, userID int64) error {
	svc, ok := services[serviceID]
	if !ok {
		return fmt.Errorf("service %d not found", serviceID)
	}
	// Логирование доступа
	log.Printf("HandleServiceDetail user_id=%d service_id=%d", userID, serviceID)

	// Эмодзи заголовка по типу
	headerEmoji := "🔧"
	switch svc.Type {
	case ServiceTypeGallery:
		headerEmoji = "🎨"
	case ServiceTypeMuseum:
		headerEmoji = "🏛️"
	case ServiceTypeSport:
		headerEmoji = "🏃"
	}

	// Формируем сообщение
	msg := fmt.Sprintf("%s %s\n\nОписание: %s\nПравила: %s", headerEmoji, svc.Name, svc.Description, svc.Rules)
	if svc.Schedule != "" {
		msg += fmt.Sprintf("\nРасписание: %s", svc.Schedule)
	}

	// Добавляем подсказку в зависимости от типа
	if svc.Type == ServiceTypeGallery || svc.Type == ServiceTypeMuseum {
		msg += "\n\nВыберите тип посещения:"
	} else if svc.Type == ServiceTypeSport {
		msg += "\n\nДоступные действия:"
	}

	// Формируем кнопки
	var buttons [][]Button
	// Для галерей и музеев: private_view, public_view
	if svc.Type == ServiceTypeGallery || svc.Type == ServiceTypeMuseum {
		buttons = append(buttons, []Button{
			{Text: "Приватный тур", Callback: fmt.Sprintf("service_%d:private_view", svc.ID)},
			{Text: "Групповой тур", Callback: fmt.Sprintf("service_%d:public_view", svc.ID)},
		})
	}
	// Для спорта: book_now
	if svc.Type == ServiceTypeSport && svc.Schedule != "" {
		buttons = append(buttons, []Button{{Text: "Забронировать", Callback: fmt.Sprintf("service_%d:book_now", svc.ID)}})
	}
	// Всегда добавляем кнопку назад
	buttons = append(buttons, []Button{{Text: "Назад", Callback: "back_to_box_solutions"}})

	// Отправляем сообщение
	if err := Sender(userID, msg, buttons); err != nil {
		return fmt.Errorf("send message: %w", err)
	}
	return nil
}
