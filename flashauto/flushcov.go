package flushauto

import (
	"log"
	"net/http"
	"os"
	"runtime/coverage"
	"time"

	"github.com/gin-gonic/gin"
)

func init() {
	go func() {
		time.Sleep(1 * time.Second)

		// the main app should use gin.Default()
		r := gin.Default()

		r.GET("/flushcov", func(c *gin.Context) {
			_ = os.Setenv("GOCOVERDIR", ".coverdata")

			if err := coverage.WriteMetaDir(".coverdata"); err != nil {
				log.Println("[flushauto] Failed to write meta:", err)
			}
			if err := coverage.WriteCountersDir(".coverdata"); err != nil {
				log.Println("[flushauto] Failed to write counters:", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			log.Println("[flushauto] Coverage data flushed.")
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		go r.Run(":9999") // Expose flushcov on another port
		log.Println("[flushauto] Registered /flushcov on :9999")
	}()
}
