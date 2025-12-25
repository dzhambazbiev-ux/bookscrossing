package transport

import (
	"net/http"
	"strconv"

	"github.com/dasler-fw/bookcrossing/internal/dto"
	"github.com/dasler-fw/bookcrossing/internal/middleware"
	"github.com/dasler-fw/bookcrossing/internal/services"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userServ services.UserService
}

func NewUserHandler(userServ services.UserService) *UserHandler {
	return &UserHandler{userServ: userServ}
}

func (h *UserHandler) RegisterRoutes(r *gin.Engine) {
	users := r.Group("/users")
	{
		users.POST("/register", h.Register)
		users.POST("/login", h.Login)
		users.GET("/:id",middleware.JWTAuth(), h.GetProfile)
		users.PATCH("/:id", middleware.JWTAuth(), h.UpdateProfile)
		users.GET("/:id/exchanges",middleware.JWTAuth(), h.GetUserExchanges)

	}

}

func (h *UserHandler) Register(c *gin.Context) {
	var req dto.UserCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	token, err := h.userServ.Register(req)
	if err != nil {
		if err.Error() == "email уже используется" {
			c.JSON(http.StatusConflict, gin.H{
				"error": "email уже используется",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "не удалось зарегистрировать пользователя",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"token": token,
	})

}

func (h *UserHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	token, err := h.userServ.Login(req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (h *UserHandler) GetProfile(c *gin.Context) {

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный идентификатор пользователя"})
		return
	}

	profile, err := h.userServ.GetProfile(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "пользователь не найден"})
		return
	}

	c.JSON(http.StatusOK, profile)
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	authUserID := c.GetUint("user_id")

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный идентификатор пользователя"})
		return
	}

	if authUserID != uint(id) {
		c.JSON(http.StatusForbidden, gin.H{"error": "доступ запрещён: нельзя редактировать чужой профиль"})
		return
	}

	var req dto.UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректное тело запроса"})
		return
	}
	if _, err := h.userServ.GetUserByID(uint(id)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "пользователь не найден",
		})
		return
	}

	if err := h.userServ.UpdateProfile(uint(id), req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "не удалось обновить профиль пользователя",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "профиль пользователя успешно обновлён",
	})

}

func (h *UserHandler) GetUserExchanges(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный идентификатор пользователя"})
		return
	}

	status := c.Query("status")

	exchanges, err := h.userServ.GetUserExchanges(uint(id), status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось получить историю обменов"})
		return
	}

	c.JSON(http.StatusOK, exchanges)
}
