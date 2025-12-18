package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kolesaev/alertmanager-discord/alertmanager"
	"github.com/kolesaev/alertmanager-discord/config"
	"github.com/kolesaev/alertmanager-discord/discord"
)

func main() {
	configs := config.LoadUserConfig()

	router := gin.Default()
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "Application is healthy!",
		})
	})

	router.POST("/:channel", func(c *gin.Context) {
		channelName := c.Param("channel")

		var alertmanagerBody alertmanager.MessageBody
		if err := c.ShouldBindJSON(&alertmanagerBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := discord.SendAlerts(channelName, alertmanagerBody, *configs); err != nil {
			log.Println("[ERROR] ", err)
		}

		c.String(http.StatusOK, "Channel: %s", channelName)
	})

	s := &http.Server{
		Addr:           configs.ListenAddress,
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	s.ListenAndServe()

}
