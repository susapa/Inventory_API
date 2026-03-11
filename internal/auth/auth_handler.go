package auth

import (
	"net/http"

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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User already exists or other error"})
		return
	}

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
	c.SetCookie("jwt", token, 3600*24, "/", "", false, true)

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
