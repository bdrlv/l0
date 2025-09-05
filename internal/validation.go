package internal

import (
	"fmt"
)

func validateMessageDataMainBody(order *Order) (bool, error) {
	if order.OrderUID == "" {
		return false, fmt.Errorf("order_uid пуст")
	}
	if order.TrackNumber == "" {
		return false, fmt.Errorf("track_number пуст")
	}
	if order.Entry == "" {
		return false, fmt.Errorf("entry пуст")
	}
	if order.Locale == "" {
		return false, fmt.Errorf("locale пуст")
	}
	if order.CustomerID == "" {
		return false, fmt.Errorf("customer_id пуст")
	}
	if order.DeliveryService == "" {
		return false, fmt.Errorf("delivery_service пуст")
	}
	if order.Shardkey == "" {
		return false, fmt.Errorf("shardkey пуст")
	}
	if order.DateCreated == "" {
		return false, fmt.Errorf("date_created пуст")
	} else {
		err := validateDateRFC3339(order.DateCreated)
		if err != nil {
			return false, fmt.Errorf("date_created %w", err)
		}
	}
	if order.OofShard == "" {
		return false, fmt.Errorf("oof_shard пуст")
	}
	return true, nil
}

func validateMessageDataDelivery(order *Order) (bool, error) {
	if order.Delivery.Name == "" {
		return false, fmt.Errorf("name пуст")
	}
	err := validatePhoneNumber(order.Delivery.Phone)
	if err != nil {
		return false, fmt.Errorf("некорректный формат phone: %w", err)
	}
	err = validateZipCode(order.Delivery.Zip)
	if err != nil {
		return false, fmt.Errorf("некорректный формат zip: %w", err)
	}
	if order.Delivery.City == "" {
		return false, fmt.Errorf("city пуст")
	}
	if order.Delivery.Address == "" {
		return false, fmt.Errorf("address пуст")
	}
	if order.Delivery.Region == "" {
		return false, fmt.Errorf("region пуст")
	}
	if order.Delivery.Email == "" {
		return false, fmt.Errorf("email пуст")
	}

	return true, nil
}

func validateMessageDataPayment(order *Order) (bool, error) {
	if order.Payment.Transaction == "" {
		return false, fmt.Errorf("transaction пуст")
	}

	if order.Payment.Currency == "" {
		return false, fmt.Errorf("transaction пуст")
	} else {
		if order.Payment.Currency != "USB" && order.Payment.Currency != "RUR" {
			return false, fmt.Errorf("некорректная валюта, ожидается 'USD' или 'RUR'")
		}
	}

	if order.Payment.Provider == "" {
		return false, fmt.Errorf("provider пуст")
	} else {
		if order.Payment.Provider != "wbpay" && order.Payment.Provider != "other" {
			return false, fmt.Errorf("некорректный провайдер, ожидается 'wbpay' или 'other'")
		}
	}

	if order.Payment.Amount <= 0 {
		return false, fmt.Errorf("amount не может быть отрицательной")
	}

	err := validateTimestamp(order.Payment.PaymentDt)
	if err != nil {
		return false, fmt.Errorf("ошибка валидации payment_dt: %w", err)
	}

	if order.Payment.Bank == "" {
		return false, fmt.Errorf("bank пуст")
	} else {
		if order.Payment.Bank != "alpha" && order.Payment.Bank != "tbank" && order.Payment.Bank != "sber" {
			return false, fmt.Errorf("некорректный банк, ожидается 'alpha', 'tbank' или 'sber'")
		}
	}

	if order.Payment.DeliveryCost < 0 {
		return false, fmt.Errorf("delivery_cost не может быть отрицательной")
	}

	if order.Payment.GoodsTotal <= 0 {
		return false, fmt.Errorf("goods_total не может быть отрицательным или 0")
	}

	if order.Payment.CustomFee < 0 {
		return false, fmt.Errorf("custom_fee не может быть отрицательной")
	}

	return true, nil
}

func validateMessageDataItems(order *Order) (bool, error) {
	if len(order.Items) == 0 {
		return false, fmt.Errorf("количетсво товаров в заказе не может быть нулевым")
	}

	for _, item := range order.Items {
		if item.ChrtID <= 0 {
			return false, fmt.Errorf("chrt_id не может быть отрицательным или равным 0")
		}

		if item.TrackNumber == "" {
			return false, fmt.Errorf("track_number не может быть пустым")
		}

		if item.Price <= 0 {
			return false, fmt.Errorf("price не может быть отрицательным или равным 0")
		}

		if item.Rid == "" {
			return false, fmt.Errorf("rid не может быть пустым")
		}

		if item.Name == "" {
			return false, fmt.Errorf("name не может быть пустым")
		}

		if item.Sale < 0 && item.Sale > 100 {
			return false, fmt.Errorf("sale должно быть в диапозоне от 0 до 100")
		}

		if item.Size == "" {
			return false, fmt.Errorf("size не может быть пустым")
		}

		if item.TotalPrice <= 0 {
			return false, fmt.Errorf("total_price не может быть отрицательным или равным 0")
		}

		if item.NmID <= 0 {
			return false, fmt.Errorf("nm_id не может быть отрицательным или равным 0")
		}

		if item.Brand == "" {
			return false, fmt.Errorf("brand не может быть пустым")
		}

		// не понятно, в каком диапозоне существуют статусы в системе, чтобы их валидировать
		if item.Status < 0 {
			return false, fmt.Errorf("status не может быть отрицательным")
		}
	}

	return true, nil
}
