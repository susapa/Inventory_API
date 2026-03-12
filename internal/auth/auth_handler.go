package auth

import (
	"log"
	"net/http"

	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/susapa/Inventory_API/internal/database"
	"github.com/susapa/Inventory_API/internal/models"
	"golang.org/x/crypto/bcrypt"
)

type RegisterInput struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginInput struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Remember bool   `json:"remember"`
}

type ForgotPasswordInput struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordInput struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// Register godoc
// @Summary      Register a new user
// @Description  Create a new user with username, email, and password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        user  body      RegisterInput  true  "User Registration Info"
// @Success      200   {object}  map[string]string
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /auth/register [post]
func Register(c *gin.Context) {
	var input RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	user := models.User{
		Username: input.Username,
		Email:    input.Email,
		Password: string(hashedPassword),
	}

	if err := database.DB.Create(&user).Error; err != nil {
		log.Printf("[AUTH] Registration failed for username %s: %v", input.Username, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User already exists or other error"})
		return
	}

	log.Printf("[AUTH] User registered: %s (%s)", input.Username, input.Email)
	c.JSON(http.StatusOK, gin.H{"message": "Registration successful"})
}

// Login godoc
// @Summary      Login a user
// @Description  Authenticate user and return a JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        credentials  body      LoginInput  true  "User Login Credentials"
// @Success      200          {object}  map[string]string
// @Failure      400          {object}  map[string]string
// @Failure      410          {object}  map[string]string
// @Router       /auth/login [post]
func Login(c *gin.Context) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := database.DB.Where("username = ?", input.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	token, err := GenerateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Set JWT in Cookie
	c.SetSameSite(http.SameSiteLaxMode)

	if input.Remember {
		c.SetCookie("jwt", token, 3600*24*10, "/", "", false, true)
		user.Remember = true
	} else {
		c.SetCookie("jwt", token, 3600*24, "/", "", false, true)
	}

	log.Printf("[AUTH] User logged in: %s", user.Username)
	c.JSON(http.StatusOK, gin.H{"message": "Login successful", "role": user.Role})
}

// Logout godoc
// @Summary      Logout a user
// @Description  Clear the authentication cookie
// @Tags         auth
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       /auth/logout [post]
func Logout(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("jwt", "", -1, "/", "", false, true)
	log.Printf("[AUTH] User logged out")
	c.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
}

// GetProfile godoc
// @Summary      Get user profile
// @Description  Return the profile of the current authenticated user
// @Tags         auth
// @Produce      json
// @Success      200  {object}  models.User
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Security     BearerAuth
// @Router       /api/profile [get]
func GetProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// GenerateRandomToken generates a hex-encoded random string
func GenerateRandomToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// ForgotPassword godoc
// @Summary      Forgot password
// @Description  Request a password reset link by providing email
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input  body      ForgotPasswordInput  true  "User Email"
// @Success      200    {object}  map[string]string
// @Failure      400    {object}  map[string]string
// @Failure      404    {object}  map[string]string
// @Failure      500    {object}  map[string]string
// @Router       /auth/forgot-password [post]
func ForgotPassword(c *gin.Context) {
	var input ForgotPasswordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := database.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User with this email not found"})
		return
	}

	token, err := GenerateRandomToken(32)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate reset token"})
		return
	}

	expiry := time.Now().Add(time.Hour) // Token expires in 1 hour
	user.ResetToken = token
	user.ResetTokenExpiresAt = &expiry

	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save reset token"})
		return
	}

	// Use the origin from where the request was made, fall back to current host if not present
	origin := c.Request.Header.Get("Origin")
	if origin == "" {
		scheme := "http"
		if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}
		origin = fmt.Sprintf("%s://%s", scheme, c.Request.Host)
	}

	resetURL := fmt.Sprintf("%s/forgot-password?token=%s", origin, token)

	log.Printf("[AUTH] Password reset requested for email: %s. Link: %s", input.Email, resetURL)

	c.JSON(http.StatusOK, gin.H{
		"message":   "Password reset link generated",
		"reset_url": resetURL,
	})
}

// ResetPassword godoc
// @Summary      Reset password
// @Description  Reset user password using a valid token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input  body      ResetPasswordInput  true  "Reset Token and New Password"
// @Success      200    {object}  map[string]string
// @Failure      400    {object}  map[string]string
// @Failure      401    {object}  map[string]string
// @Failure      500    {object}  map[string]string
// @Router       /auth/reset-password [post]
func ResetPassword(c *gin.Context) {
	var input ResetPasswordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := database.DB.Where("reset_token = ?", input.Token).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
		return
	}

	if user.ResetTokenExpiresAt == nil || user.ResetTokenExpiresAt.Before(time.Now()) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token has expired"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash new password"})
		return
	}

	user.Password = string(hashedPassword)
	user.ResetToken = ""           // Clear token
	user.ResetTokenExpiresAt = nil // Clear expiry

	if err := database.DB.Save(&user).Error; err != nil {
		log.Printf("[AUTH] Failed to update password for user %s: %v", user.Username, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	log.Printf("[AUTH] Password reset successful for user: %s", user.Username)
	c.JSON(http.StatusOK, gin.H{"message": "Password has been reset successfully"})
}
