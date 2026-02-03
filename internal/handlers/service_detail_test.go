package handlers

import (
	"strings"
	"testing"
)

func TestHandleServiceDetail_Gallery(t *testing.T) {
	var capturedUser int64
	var capturedText string
	var capturedButtons [][]Button

	Sender = func(userID int64, text string, buttons [][]Button) error {
		capturedUser = userID
		capturedText = text
		capturedButtons = buttons
		return nil
	}

	err := HandleServiceDetail(1, 12345)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedUser != 12345 {
		t.Fatalf("unexpected user, got %d", capturedUser)
	}
	if !strings.Contains(capturedText, "Третьяковская галерея") {
		t.Fatalf("missing service name in message: %s", capturedText)
	}
	if !strings.Contains(capturedText, "Описание:") || !strings.Contains(capturedText, "Правила:") {
		t.Fatalf("missing required sections in message: %s", capturedText)
	}
	if !strings.Contains(capturedText, "Расписание:") {
		t.Fatalf("expected schedule to be present for this service: %s", capturedText)
	}

	// Проверяем кнопки: приватная, групповая, назад
	want := map[string]bool{"service_1:private_view": false, "service_1:public_view": false, "back_to_box_solutions": false}
	for _, row := range capturedButtons {
		for _, b := range row {
			if _, ok := want[b.Callback]; ok {
				want[b.Callback] = true
			}
		}
	}
	for k, v := range want {
		if !v {
			t.Fatalf("missing button callback %s", k)
		}
	}
}

func TestHandleServiceDetail_Sport(t *testing.T) {
	var capturedText string
	var capturedButtons [][]Button

	Sender = func(userID int64, text string, buttons [][]Button) error {
		capturedText = text
		capturedButtons = buttons
		return nil
	}

	err := HandleServiceDetail(2, 555)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(capturedText, "Футбольный зал") {
		t.Fatalf("missing service name in message: %s", capturedText)
	}
	// должно содержать расписание
	if !strings.Contains(capturedText, "Расписание:") {
		t.Fatalf("missing schedule for sport service: %s", capturedText)
	}
	// должны быть кнопки book_now и назад
	foundBook := false
	foundBack := false
	for _, row := range capturedButtons {
		for _, b := range row {
			if b.Callback == "service_2:book_now" {
				foundBook = true
			}
			if b.Callback == "back_to_box_solutions" {
				foundBack = true
			}
		}
	}
	if !foundBook || !foundBack {
		t.Fatalf("missing required buttons for sport: book=%v back=%v", foundBook, foundBack)
	}
}

func TestHandleServiceDetail_Museum(t *testing.T) {
	var capturedText string
	var capturedButtons [][]Button

	Sender = func(userID int64, text string, buttons [][]Button) error {
		capturedText = text
		capturedButtons = buttons
		return nil
	}

	err := HandleServiceDetail(4, 999)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(capturedText, "Музей истории") {
		t.Fatalf("missing museum name in message: %s", capturedText)
	}
	// музей должен иметь приватную/групповую и кнопку назад
	foundPrivate := false
	foundPublic := false
	foundBack := false
	for _, row := range capturedButtons {
		for _, b := range row {
			if b.Callback == "service_4:private_view" {
				foundPrivate = true
			}
			if b.Callback == "service_4:public_view" {
				foundPublic = true
			}
			if b.Callback == "back_to_box_solutions" {
				foundBack = true
			}
		}
	}
	if !foundPrivate || !foundPublic || !foundBack {
		t.Fatalf("missing required buttons for museum: private=%v public=%v back=%v", foundPrivate, foundPublic, foundBack)
	}
}

func TestHandleServiceDetail_NotFound(t *testing.T) {
	Sender = func(userID int64, text string, buttons [][]Button) error { return nil }
	if err := HandleServiceDetail(9999, 1); err == nil {
		t.Fatalf("expected error for unknown service")
	}
}
