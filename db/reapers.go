package db

import (
	"log"
	"time"

	"github.com/jinzhu/gorm"
)

// These functions usually run in the background of the backend
// service and delete all stale entries from the corresponding tables
// that were created longer than a supplied offset ago.

// For offers and requests.
func OfferRequestReaper(db *gorm.DB, table string, sleepOffset time.Duration) {

	log.Printf("%s reaper started.\n", table)

	// Buffer up to 10 items before bulk expire.
	expireBufferSize := 10

	for {

		// Initialize slice to buffer items that are about to be set to 'expired'.
		expireBuffer := make([]string, 0, expireBufferSize)
		i := 0

		// Retrieve all items in ascending order by validity period.
		if table == "Offers" {

			var Items []Offer
			db.Where("\"expired\" = ?", "false").Order("\"validity_period\" ASC").Find(&Items)

			log.Println("Size of offers list:", len(Items))

			for _, item := range Items {

				// If the validity period of an item is less than the
				// current time, add the item's ID to expireBuffer.
				if item.ValidityPeriod.Before(time.Now()) {

					log.Println("offer time:", item.ValidityPeriod)
					log.Println("now   time:", time.Now())

					expireBuffer = append(expireBuffer, item.ID)

					log.Println("len of expireBuffer:", len(expireBuffer))
					log.Println("Offer with ID", item.ID, "was added to expire buffer:", expireBuffer[i], " (i is", i, ")")

					i++
				}

				// If expire buffer is full, issue a bulk update.
				if i == expireBufferSize {

					log.Println("Expire buffer full")

					db.Model(&Offer{}).Where("\"id\" IN (?)", expireBuffer).Update("expired", true)
					expireBuffer = make([]string, 0, expireBufferSize)
					i = 0

					log.Println("expireBuffer:", expireBuffer)
					log.Println("i:", i)
				}
			}

			if len(expireBuffer) > 0 {
				log.Println("Loop done, but expireBuffer contains", len(expireBuffer), "elements. Expiring.")
				db.Model(&Offer{}).Where("\"id\" IN (?)", expireBuffer).Update("expired", true)
			}

		} else if table == "Requests" {

			var Items []Request
			db.Where("\"expired\" = ?", "false").Order("\"validity_period\" ASC").Find(&Items)

			log.Println("Size of requests list:", len(Items))

			for _, item := range Items {

				// If the validity period of an item is less than the
				// current time, add the item's ID to expireBuffer.
				if item.ValidityPeriod.Before(time.Now()) {

					log.Println("request time:", item.ValidityPeriod)
					log.Println("now     time:", time.Now())

					expireBuffer = append(expireBuffer, item.ID)

					log.Println("len of expireBuffer:", len(expireBuffer))
					log.Println("Request with ID", item.ID, "was added to expire buffer:", expireBuffer[i], " (i is", i, ")")

					i++
				}

				// If expire buffer is full, issue a bulk update.
				if i == expireBufferSize {

					log.Println("Expire buffer full")

					db.Model(&Request{}).Where("\"id\" IN (?)", expireBuffer).Update("expired", true)
					expireBuffer = make([]string, 0, expireBufferSize)
					i = 0

					log.Println("expireBuffer:", expireBuffer)
					log.Println("i:", i)
				}
			}

			if len(expireBuffer) > 0 {
				log.Println("Loop done, but expireBuffer contains", len(expireBuffer), "elements. Expiring.")
				db.Model(&Request{}).Where("\"id\" IN (?)", expireBuffer).Update("expired", true)
			}

		} else {
			log.Println("[OfferRequestReaper] Invalid table name supplied.")
			return
		}

		log.Printf("%s reaper done. Sleeping for %s.\n", table, sleepOffset.String())

		// Let function execution sleep until next round.
		time.Sleep(sleepOffset)
	}
}

// For notifications.
func NotificationReaper(db *gorm.DB, expiryOffset time.Duration, sleepOffset time.Duration) {

	log.Println("Notifications reaper started.")

	// Buffer up to 10 items before bulk delete.
	deleteBufferSize := 10

	for {

		// Retrieve all read notifications in ascending order by date.
		var Notifications []Notification
		db.Where("\"read\" = ?", "true").Order("\"created_at\" ASC").Find(&Notifications)

		// Initialize slice to buffer to-be-deleted notifications.
		deleteBuffer := make([]string, 0, deleteBufferSize)
		i := 0

		for _, notification := range Notifications {

			// If notification was created more than expiryOffset duration
			// ago, add ID of notification to deletion buffer.
			if notification.CreatedAt.Add(expiryOffset).Before(time.Now()) {

				deleteBuffer = append(deleteBuffer, notification.ID)
				i++
			}

			// If delete buffer is full, issue a bulk deletion.
			if i == deleteBufferSize {

				db.Where("\"id\" IN (?)", deleteBuffer).Delete(&Notification{})
				deleteBuffer = make([]string, 0, deleteBufferSize)
				i = 0
			}
		}

		if len(deleteBuffer) > 0 {
			db.Where("\"id\" IN (?)", deleteBuffer).Delete(&Notification{})
		}

		log.Printf("Notifications reaper done. Sleeping for %s.\n", sleepOffset.String())

		// Let function execution sleep until next round.
		time.Sleep(sleepOffset)
	}
}
