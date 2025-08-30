package controllers

import (
	"encoding/json"
	"net/http"
	"strings"

	"alkaukaba-backend/config"
	"alkaukaba-backend/database"
	"alkaukaba-backend/models"
	"alkaukaba-backend/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var googleConf *oauth2.Config

func InitAuth(cfg config.Config) {
	googleConf = &oauth2.Config{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		RedirectURL:  cfg.GoogleRedirectURL,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}
}

// ===== Register (local) =====
type RegisterInput struct {
	Name     string `json:"name" binding:"required,min=2"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

func Register(c *gin.Context) {
	var in RegisterInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// check existing email
	var exists int64
	database.DB.Model(&models.User{}).Where("email = ?", strings.ToLower(in.Email)).Count(&exists)
	if exists > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email already registered"})
		return
	}

	hash, _ := utils.HashPassword(in.Password)
	user := models.User{
		Name:     in.Name,
		Email:    strings.ToLower(in.Email),
		Password: hash,
		Provider: "local",
	}
	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}
	tok, _ := utils.GenerateToken(user.ID)
	c.JSON(http.StatusOK, gin.H{"message": "registered", "token": tok, "user": user})
}

// ===== Login (local) =====
type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func Login(c *gin.Context) {
	var in LoginInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := database.DB.Where("email = ?", strings.ToLower(in.Email)).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	if user.Provider != "local" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "use Google login for this account"})
		return
	}
	if !utils.CheckPasswordHash(in.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	tok, _ := utils.GenerateToken(user.ID)
	c.JSON(http.StatusOK, gin.H{"message": "logged in", "token": tok, "user": user})
}

// ===== Google OAuth =====
func GoogleLogin(c *gin.Context) {
	if googleConf == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "google oauth not configured"})
		return
	}
	state := "state-token" // TODO: replace with random/CSRF state stored in cookie/session
	url := googleConf.AuthCodeURL(state, oauth2.AccessTypeOffline)
	c.Redirect(http.StatusFound, url)
}

func GoogleCallback(c *gin.Context) {
	if googleConf == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "google oauth not configured"})
		return
	}
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing code"})
		return
	}
	tok, err := googleConf.Exchange(c, code)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "oauth exchange failed"})
		return
	}
	client := googleConf.Client(c, tok)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil || resp.StatusCode != 200 {
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to fetch userinfo"})
		return
	}
	defer resp.Body.Close()
	var gu struct {
		ID      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&gu); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "invalid userinfo"})
		return
	}

	// Upsert user by GoogleID (fallback to email)
	var user models.User
	if err := database.DB.Where("google_id = ?", gu.ID).First(&user).Error; err != nil {
		// Not found by GoogleID, try email
		if err := database.DB.Where("email = ?", strings.ToLower(gu.Email)).First(&user).Error; err != nil {
			// Create new
			user = models.User{
				Name:       gu.Name,
				Email:      strings.ToLower(gu.Email),
				Provider:   "google",
				GoogleID:   gu.ID,
				PictureURL: gu.Picture,
			}
			_ = database.DB.Create(&user).Error
		} else {
			// Link existing local account to Google
			database.DB.Model(&user).Updates(map[string]interface{}{
				"provider":    "google",
				"google_id":   gu.ID,
				"picture_url": gu.Picture,
			})
		}
	} else {
		// Update profile pic/name if changed
		database.DB.Model(&user).Updates(map[string]interface{}{"name": gu.Name, "picture_url": gu.Picture})
	}

	jwt, _ := utils.GenerateToken(user.ID)
	// For web apps, you could set cookie instead of JSON
	c.JSON(http.StatusOK, gin.H{"message": "google logged in", "token": jwt, "user": user})
}

// ===== Me (protected) =====
func Me(c *gin.Context) {
	uidVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no user"})
		return
	}
	uid := uidVal.(uint)
	var user models.User
	if err := database.DB.First(&user, uid).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user})
}

// ===== Update User (name & picture) =====
type UpdateUserInput struct {
	Name       string `json:"name"`
	PictureURL string `json:"picture_url"`
}

func UpdateUser(c *gin.Context) {
	uidVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	uid := uidVal.(uint)

	var in UpdateUserInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := database.DB.First(&user, uid).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	updates := map[string]interface{}{}
	if in.Name != "" {
		updates["name"] = in.Name
	}
	if in.PictureURL != "" {
		updates["picture_url"] = in.PictureURL
	}

	database.DB.Model(&user).Updates(updates)
	c.JSON(http.StatusOK, gin.H{"message": "user updated", "user": user})
}

// ===== Update Password (need old password) =====
type UpdatePasswordInput struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

func UpdatePassword(c *gin.Context) {
	uidVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	uid := uidVal.(uint)

	var in UpdatePasswordInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := database.DB.First(&user, uid).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	if !utils.CheckPasswordHash(in.OldPassword, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "old password incorrect"})
		return
	}

	hash, _ := utils.HashPassword(in.NewPassword)
	database.DB.Model(&user).Update("password", hash)
	c.JSON(http.StatusOK, gin.H{"message": "password updated"})
}
