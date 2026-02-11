package handlers

import (
	"testing"
)

type fakeSender struct {
	lastUser int64
	lastText string
	lastBtns [][]Button
}

func (f *fakeSender) SendMessage(userID int64, text string, buttons [][]Button) error {
	f.lastUser = userID
	f.lastText = text
	f.lastBtns = buttons
	return nil
}

func TestHandleServiceDetail_Gallery(t *testing.T) {
	fs := &fakeSender{}
	Sender = fs
	defer func() { Sender = defaultSender{} }()

	if err := HandleServiceDetail(1, 42); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Проверяем, что сообщение содержит ожидаемые разделы
	if fs.lastUser != 42 {
		t.Fatalf("expected user 42, got %d", fs.lastUser)
	}
	if fs.lastText == "" {
		t.Fatal("expected non-empty message text")
	}
	if !contains(fs.lastText, "Описание:") || !contains(fs.lastText, "Правила:") || !contains(fs.lastText, "Расписание:") {
		t.Fatalf("message missing expected sections: %s", fs.lastText)
	}
	// Кнопки: опции + Назад
	if len(fs.lastBtns) != 1 {
		t.Fatalf("expected 1 row of buttons, got %d", len(fs.lastBtns))
	}
	row := fs.lastBtns[0]
	if len(row) < 2 {
		t.Fatalf("expected at least 2 buttons, got %d", len(row))
	}
	if row[len(row)-1].CallbackData != "back_to_box_solutions" {
		t.Fatalf("expected last button to be back_to_box_solutions, got %s", row[len(row)-1].CallbackData)
	}
}

func TestHandleServiceDetail_Sport(t *testing.T) {
	fs := &fakeSender{}
	Sender = fs
	defer func() { Sender = defaultSender{} }()

	if err := HandleServiceDetail(2, 100); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !contains(fs.lastText, "Забронировать") && !hasBookingButton(fs.lastBtns) {
		t.Fatalf("expected booking button in message or buttons")
	}
}

func TestHandleServiceDetail_NotFound(t *testing.T) {
	fs := &fakeSender{}
	Sender = fs
	defer func() { Sender = defaultSender{} }()

	if err := HandleServiceDetail(999, 1); err == nil {
		t.Fatalf("expected error for unknown service, got nil")
	}
}

// вспомогательные функции
func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || (len(s) > len(sub) && (index(s, sub) >= 0)))
}

func index(s, sep string) int {
	for i := 0; i+len(sep) <= len(s); i++ {
		if s[i:i+len(sep)] == sep {
			return i
		}
	}
	return -1
}

func hasBookingButton(btns [][]Button) bool {
	for _, row := range btns {
		for _, b := range row {
			if b.CallbackData == "book_now:2" || b.Text == "Забронировать" {
				return true
			}
		}
	}
	return false
}
