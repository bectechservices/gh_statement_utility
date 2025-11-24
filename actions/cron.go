package actions

import (
	"log"
	"ng-statement-app/models"

	"github.com/robfig/cron"
)

func StartCronScheduler() *cron.Cron {
	c := cron.New()
	c.Start()
	//_ = c.AddFunc("0 * * * * *", models.IsloggedInResetForAllUsers)
	_ = c.AddFunc("@every 5m", models.IsloggedInResetForAllUsers)
	_ = c.AddFunc("@every 24h", func() {
		if err := models.DormancyOnLastLogin(models.GormDB); err != nil {
			log.Println("Dormancy job error:", err)
		}
	})
	_ = c.AddFunc("@every 24h", func() {
		if err := models.DeletedUsersUpdate(models.GormDB); err != nil {
			log.Println("Deleted_at Update job error:", err)
		}
	})
	log.Println("Starting Cron...........")
	return c
}
