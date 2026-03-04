package handlers

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"

	"github.com/yandex-development-2-team/Go/internal/models"
)

type BookingRepository interface {
	SaveBooking(ctx context.Context, state *models.BookingState) error
	GetAvailableDates(ctx context.Context, serviceID int) ([]time.Time, error)
}

type BookingFormHandler struct {
	bot   *tgbotapi.BotAPI
	db    BookingRepository
	log   *zap.Logger
	store map[int64]*models.BookingState
	mu    sync.RWMutex
}

func NewBookingFormHandler(
	bot *tgbotapi.BotAPI,
	db BookingRepository,
	log *zap.Logger,
) *BookingFormHandler {
	return &BookingFormHandler{
		bot:   bot,
		db:    db,
		log:   log,
		store: make(map[int64]*models.BookingState),
	}
}

func (h *BookingFormHandler) Start(userID int64, serviceID int, visitType string) {
	h.mu.Lock()
	h.store[userID] = &models.BookingState{
		UserID:    userID,
		ServiceID: serviceID,
		VisitType: visitType,
		Step:      models.BookingStepSelectDate,
		CreatedAt: time.Now(),
	}
	h.mu.Unlock()
}

func (h *BookingFormHandler) HandleUpdate(update tgbotapi.Update) {
	if update.CallbackQuery != nil {
		h.handleCallback(update.CallbackQuery)
		return
	}

	if update.Message != nil {
		h.handleMessage(update.Message)
	}
}

func (h *BookingFormHandler) getState(userID int64) (*models.BookingState, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	state, ok := h.store[userID]
	return state, ok
}

func (h *BookingFormHandler) setState(userID int64, state *models.BookingState) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.store[userID] = state
}

func (h *BookingFormHandler) clearState(userID int64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.store, userID)
}

func (h *BookingFormHandler) handleCallback(q *tgbotapi.CallbackQuery) {
	state, ok := h.getState(q.From.ID)
	if !ok {
		return
	}

	switch state.Step {
	case models.BookingStepSelectDate:
		date, err := time.Parse("2006-01-02", q.Data)
		if err != nil {
			return
		}

		state.SelectedDate = date
		state.Step = models.BookingStepGuestName
		h.setState(q.From.ID, state)

		msg := tgbotapi.NewMessage(q.Message.Chat.ID, "Введите ФИО:")
		h.bot.Send(msg)

	case models.BookingStepConfirm:
		if q.Data == "confirm_yes" {
			err := h.db.SaveBooking(context.Background(), state)
			if err != nil {
				h.log.Error("save booking error", zap.Error(err))
				return
			}

			h.clearState(q.From.ID)

			msg := tgbotapi.NewMessage(q.Message.Chat.ID, "Готово! Возврат на главную")
			h.bot.Send(msg)

			// тут можно вызвать функцию возврата в меню
		}

		if q.Data == "confirm_no" {
			state.Step = models.BookingStepSelectDate
			h.setState(q.From.ID, state)
			h.sendDateSelection(q.Message.Chat.ID, state.ServiceID)
		}
	}
}

func (h *BookingFormHandler) handleMessage(msg *tgbotapi.Message) {
	state, ok := h.getState(msg.From.ID)
	if !ok {
		return
	}

	text := strings.TrimSpace(msg.Text)

	switch state.Step {

	case models.BookingStepGuestName:
		if len(text) < 3 || len(text) > 100 {
			h.bot.Send(tgbotapi.NewMessage(msg.Chat.ID,
				"Ошибка: ФИО должно быть от 3 до 100 символов. Введите снова:"))
			return
		}

		state.GuestName = text
		state.Step = models.BookingStepOrg
		h.setState(msg.From.ID, state)

		h.bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Введите организацию:"))

	case models.BookingStepOrg:
		if len(text) < 2 || len(text) > 255 {
			h.bot.Send(tgbotapi.NewMessage(msg.Chat.ID,
				"Ошибка: организация должна быть от 2 до 255 символов. Введите снова:"))
			return
		}

		state.GuestOrganization = text
		state.Step = models.BookingStepPosition
		h.setState(msg.From.ID, state)

		h.bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Введите должность:"))

	case models.BookingStepPosition:
		if len(text) < 2 || len(text) > 100 {
			h.bot.Send(tgbotapi.NewMessage(msg.Chat.ID,
				"Ошибка: должность должна быть от 2 до 100 символов. Введите снова:"))
			return
		}

		state.GuestPosition = text
		state.Step = models.BookingStepConfirm
		h.setState(msg.From.ID, state)

		h.sendConfirmation(msg.Chat.ID, state)
	}
}

func (h *BookingFormHandler) sendDateSelection(chatID int64, serviceID int) {
	dates, err := h.db.GetAvailableDates(context.Background(), serviceID)
	if err != nil {
		h.log.Error("get dates error", zap.Error(err))
		return
	}

	var buttons [][]tgbotapi.InlineKeyboardButton
	for _, d := range dates {
		dateStr := d.Format("2006-01-02")
		btn := tgbotapi.NewInlineKeyboardButtonData(dateStr, dateStr)
		buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(btn))
	}

	msg := tgbotapi.NewMessage(chatID, "Выберите дату:")
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(buttons...)

	h.bot.Send(msg)
}

func (h *BookingFormHandler) sendConfirmation(chatID int64, state *models.BookingState) {
	text := fmt.Sprintf(
		"Подтвердите бронирование:\n\nДата: %s\nФИО: %s\nОрганизация: %s\nДолжность: %s",
		state.SelectedDate.Format("02.01.2006"),
		state.GuestName,
		state.GuestOrganization,
		state.GuestPosition,
	)

	yes := tgbotapi.NewInlineKeyboardButtonData("Подтвердить", "confirm_yes")
	no := tgbotapi.NewInlineKeyboardButtonData("Изменить дату", "confirm_no")

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(yes, no),
	)

	h.bot.Send(msg)
}