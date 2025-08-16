package controllers

import (
    "net/http"
    "alkaukaba-backend/database"
    "github.com/gin-gonic/gin"
)

func HealthCheck(c *gin.Context) {
    sqlDB, err := database.DB.DB()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"status": "db error", "error": err.Error()})
        return
    }

    if err := sqlDB.Ping(); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"status": "db not reachable", "error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"status": "ok", "db": "connected"})
}
