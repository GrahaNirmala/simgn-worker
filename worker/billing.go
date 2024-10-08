package worker

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"firebase.google.com/go/messaging"
	"github.com/jinzhu/now"
	"github.com/Roofiif/sim-graha-nirmala-worker/entity"
	"github.com/Roofiif/sim-graha-nirmala-worker/logger"
)

func (w *worker) generateMonthlyBilling() {
	ctx := context.Background()

	// Load Jakarta time zone
	loc, _ := time.LoadLocation("Asia/Jakarta")

	occupants := make([]*entity.Occupant, 0)
	err := w.db.Conn().NewSelect().Model(&occupants).Scan(ctx)
	if err != nil {
		logger.Log().Error("failed to get occupants", "error", err)
		return
	}

	var billingConfig entity.BillingConfig
	err = w.db.Conn().NewSelect().Model(&billingConfig).Limit(1).Scan(ctx)
	if err != nil {
		logger.Log().Error("failed to get billing configuration", "error", err)
		return
	}

	var deviceToken entity.DeviceToken
	err = w.db.Conn().NewSelect().Model(&deviceToken).Limit(1).Scan(ctx)
	if err != nil {
		logger.Log().Error("failed to get device token", "error", err)
		return
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	for _, occupant := range occupants {
		var existingBillings []*entity.Billing
		err := w.db.Conn().NewSelect().
			Model(&existingBillings).
			Where("house_id = ?", occupant.HouseId).
			Where("period < ?", now.New(time.Now().In(loc)).BeginningOfMonth()). // Set location here
			Order("period ASC").
			Scan(ctx)
		if err != nil && err != sql.ErrNoRows {
			logger.Log().Error("failed to get previous billings", "error", err)
			return
		}

		for _, existingBilling := range existingBillings {
			if !existingBilling.IsPaid && isFirstDayOfMonth(loc) { // Pass loc here
				penaltyKey := fmt.Sprintf("%d-%s", occupant.HouseId, existingBilling.Period.Format("2006-01"))
				if !w.checked[penaltyKey] {
					totalDenda := calculatePenalty(existingBilling.Period, billingConfig.ExtraChargeBill, loc) // Pass loc here
					logger.Log().Info("Calculating penalty", "house_id", occupant.HouseId, "period", existingBilling.Period, "totalDenda", totalDenda)

					existingBilling.ExtraCharge += totalDenda

					_, err := w.db.Conn().NewUpdate().Model(existingBilling).WherePK().Exec(ctx)
					if err != nil {
						logger.Log().Error("failed to update billing with penalty", "error", err)
						return
					}

					logger.Log().Info("penalty added", "house_id", occupant.HouseId, "penalty", totalDenda)

					w.checked[penaltyKey] = true
				}
			}
		}

		var currentBilling entity.Billing
		err = w.db.Conn().NewSelect().
			Model(&currentBilling).
			Where("house_id = ?", occupant.HouseId).
			Where("period >= ?", now.New(time.Now().In(loc)).BeginningOfMonth()). // Set location here
			Where("period <= ?", now.New(time.Now().In(loc)).EndOfMonth()). // Set location here
			Scan(ctx)
		if err != nil && err != sql.ErrNoRows {
			logger.Log().Error("failed to get current billing", "error", err)
			return
		}

		if err == sql.ErrNoRows {
			billing := &entity.Billing{
				HouseId:     occupant.HouseId,
				Period:      now.New(time.Now().In(loc)).BeginningOfMonth(), // Set location here
				Amount:      billingConfig.AmountBill,
				IsPaid:      false,
				ExtraCharge: 0,
			}

			_, err = w.db.Conn().NewInsert().Model(billing).Exec(ctx)
			if err != nil {
				logger.Log().Error("failed to insert billing", "error", err)
				return
			}
			
			var deviceToken entity.DeviceToken
            err = w.db.Conn().NewSelect().
                Model(&deviceToken).
                Where("occupant_id = ?", occupant.Id).
                Limit(1).Scan(ctx)
            if err != nil {
                logger.Log().Error("failed to get device token for occupant", "occupant_id", occupant.Id, "error", err)
                return
            }

			message := &messaging.Message{
				Token: deviceToken.DeviceToken,
				Notification: &messaging.Notification{
					Title: "Tagihan Bulan ini",
					Body:  fmt.Sprintf("Tagihan anda Bulan ini sebesar Rp %d", billingConfig.AmountBill),
				},
			}

			_, err = w.fcmClient.Send(ctx, message)
			if err != nil {
				logger.Log().Error("failed to send push notification", "error", err)
				continue
			}

			logger.Log().Info("push notification sent", "house_id", occupant.HouseId)
			logger.Log().Info("billing generated", "house_id", occupant.HouseId)
		}
	}
}

func (w *worker) send15thDayNotification() {
	ctx := context.Background()

	// Load Jakarta time zone
	loc, _ := time.LoadLocation("Asia/Jakarta")

	occupants := make([]*entity.Occupant, 0)
	err := w.db.Conn().NewSelect().Model(&occupants).Scan(ctx)
	if err != nil {
		logger.Log().Error("failed to get occupants", "error", err)
		return
	}

	for _, occupant := range occupants {
		var currentBilling entity.Billing
		err = w.db.Conn().NewSelect().
			Model(&currentBilling).
			Where("house_id = ?", occupant.HouseId).
			Where("period >= ?", now.New(time.Now().In(loc)).BeginningOfMonth()). // Set location here
			Where("period <= ?", now.New(time.Now().In(loc)).EndOfMonth()). // Set location here
			Scan(ctx)
		if err != nil && err != sql.ErrNoRows {
			logger.Log().Error("failed to get current billing", "error", err)
			return
		}

		if currentBilling.IsPaid {
			continue
		}

		deviceTokens := make([]*entity.DeviceToken, 0)
		err = w.db.Conn().NewSelect().
			Model(&deviceTokens).
			Where("occupant_id = ?", occupant.Id).
			Scan(ctx)
		if err != nil {
			logger.Log().Error("failed to get device tokens", "error", err)
			return
		}

		for _, deviceToken := range deviceTokens {
			w.mu.Lock()
			lastSentDate, exists := w.lastNotificationDate[occupant.HouseId]
			w.mu.Unlock()

			today := time.Now().In(loc) // Use Jakarta time zone here

			if today.Day() == 15 && (!exists || lastSentDate.Before(today)) {
				message := &messaging.Message{
					Token: deviceToken.DeviceToken,
					Notification: &messaging.Notification{
						Title: "Tagihan Sudah 15 Hari",
						Body:  fmt.Sprintf("Tagihan anda untuk bulan %s sebesar Rp %d sudah 15 hari belum terbayar.", formatIndonesianMonthYear(currentBilling.Period), currentBilling.Amount),
					},
				}

				_, err = w.fcmClient.Send(ctx, message)
				if err != nil {
					logger.Log().Error("failed to send 15th day notification", "error", err)
					continue
				}

				w.mu.Lock()
				w.lastNotificationDate[occupant.HouseId] = today
				w.mu.Unlock()

				logger.Log().Info("15th day notification sent", "house_id", occupant.HouseId)
			}
		}
	}
}

func (w *worker) send22thDayNotification() {
	ctx := context.Background()

	// Load Jakarta time zone
	loc, _ := time.LoadLocation("Asia/Jakarta")

	occupants := make([]*entity.Occupant, 0)
	err := w.db.Conn().NewSelect().Model(&occupants).Scan(ctx)
	if err != nil {
		logger.Log().Error("failed to get occupants", "error", err)
		return
	}

	for _, occupant := range occupants {
		var currentBilling entity.Billing
		err = w.db.Conn().NewSelect().
			Model(&currentBilling).
			Where("house_id = ?", occupant.HouseId).
			Where("period >= ?", now.New(time.Now().In(loc)).BeginningOfMonth()). // Set location here
			Where("period <= ?", now.New(time.Now().In(loc)).EndOfMonth()). // Set location here
			Scan(ctx)
		if err != nil && err != sql.ErrNoRows {
			logger.Log().Error("failed to get current billing", "error", err)
			return
		}

		if currentBilling.IsPaid {
			continue
		}

		deviceTokens := make([]*entity.DeviceToken, 0)
		err = w.db.Conn().NewSelect().
			Model(&deviceTokens).
			Where("occupant_id = ?", occupant.Id).
			Scan(ctx)
		if err != nil {
			logger.Log().Error("failed to get device tokens", "error", err)
			return
		}

		for _, deviceToken := range deviceTokens {
			w.mu.Lock()
			lastSentDate, exists := w.lastNotificationDate[occupant.HouseId]
			w.mu.Unlock()

			today := time.Now().In(loc) // Use Jakarta time zone here

			if today.Day() == 22 && (!exists || lastSentDate.Before(today)) {
				message := &messaging.Message{
					Token: deviceToken.DeviceToken,
					Notification: &messaging.Notification{
						Title: "Tagihan Sudah 22 Hari",
						Body:  fmt.Sprintf("Tagihan anda untuk bulan %s sebesar Rp %d sudah 22 hari belum terbayar.", formatIndonesianMonthYear(currentBilling.Period), currentBilling.Amount),
					},
				}

				_, err = w.fcmClient.Send(ctx, message)
				if err != nil {
					logger.Log().Error("failed to send 22th day notification", "error", err)
					continue
				}

				w.mu.Lock()
				w.lastNotificationDate[occupant.HouseId] = today
				w.mu.Unlock()

				logger.Log().Info("22th day notification sent", "house_id", occupant.HouseId)
			}
		}
	}
}

// Helper functions

func getIndonesianMonth(month time.Month) string {
	months := map[time.Month]string{
		time.January:   "Januari",
		time.February:  "Februari",
		time.March:     "Maret",
		time.April:     "April",
		time.May:       "Mei",
		time.June:      "Juni",
		time.July:      "Juli",
		time.August:    "Agustus",
		time.September: "September",
		time.October:   "Oktober",
		time.November:  "November",
		time.December:  "Desember",
	}
	return months[month]
}

func formatIndonesianMonthYear(t time.Time) string {
	month := getIndonesianMonth(t.Month())
	year := t.Year()
	return fmt.Sprintf("%s %d", month, year)
}

func calculatePenalty(period time.Time, extraChargeBill int64, loc *time.Location) int64 { // Adjusted to use location
	now := time.Now().In(loc)

	if now.After(period) {
		logger.Log().Info("Penalty applied", "period", period, "now", now, "penalty", extraChargeBill)
		return extraChargeBill
	}

	return 0
}

func isFirstDayOfMonth(loc *time.Location) bool { // Adjusted to use location
	return time.Now().In(loc).Day() == 1
}
