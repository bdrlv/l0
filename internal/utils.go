package internal

import (
	"fmt"
	"strings"
	"time"
	"unicode"
)

func validateDateRFC3339(dateString string) error {
	_, err := time.Parse(time.RFC3339, dateString)
	if err != nil {
		return fmt.Errorf("некорректный формат даты %w", err)
	}
	return nil
}

func validatePhoneNumber(phoneNumberString string) error {
	if phoneNumberString == "" {
		return fmt.Errorf("пустое значение")
	}

	if !strings.HasPrefix(phoneNumberString, "+") {
		return fmt.Errorf("значение должно начинаться с '+'")
	}

	phoneNumberWithoutPrefix := phoneNumberString[1:]
	if len(phoneNumberWithoutPrefix) != 11 {
		return fmt.Errorf("длина номера не равна 11 цифрам")
	}

	for _, item := range phoneNumberWithoutPrefix {
		if !unicode.IsDigit(item) {
			return fmt.Errorf("содержит недопустимый символ %c", item)
		}
	}

	return nil
}

func validateZipCode(zipString string) error {
	if zipString == "" {
		return fmt.Errorf("пустое значение")
	}

	if len(zipString) < 5 || len(zipString) > 7 {
		return fmt.Errorf("длина кода должна быть больше 5 и меньше 7 цифр")
	}

	for _, item := range zipString {
		if !unicode.IsDigit(item) {
			return fmt.Errorf("содержит недопустимый символ %c", item)
		}
	}

	return nil
}

func validateTimestamp(stamp int64) error {
	if stamp <= 0 {
		return fmt.Errorf("значение не может быть <= 0")
	}
	ut := time.Unix(stamp, 0)
	delta := ut.Year() - time.Now().Year()
	if delta < 0 {
		delta = -delta
	}
	if delta > 1 { // сделал так, если заказ был оплачен в новый год и, в теории, мы получили дельту в 1 год между полатежом и процессингом заказа
		return fmt.Errorf("год платежа отличается от текущего более, чем на 1")
	}
	return nil
}
