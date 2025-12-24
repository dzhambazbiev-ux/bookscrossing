package transport

import (
	"net/http"
	"strconv"

	"github.com/dasler-fw/bookcrossing/internal/dto"
	"github.com/dasler-fw/bookcrossing/internal/middleware"
	"github.com/dasler-fw/bookcrossing/internal/services"
	"github.com/gin-gonic/gin"
	
)

type ReviewHandler struct {
	service services.ReviewService
}

func NewReviewHandler(service services.ReviewService) *ReviewHandler {
	return &ReviewHandler{service: service}
}

func (h *ReviewHandler) RegisterReviewRoutes(r *gin.Engine) {
	r.POST("/review",middleware.JWTAuth(), h.Create)
	r.DELETE("/review/:id",middleware.JWTAuth(), h.Delete)
	r.GET("/users/:id/review", h.GetByUser)
	r.GET("/book/:id/review", h.GetByBook)
}

func (h *ReviewHandler) Create(c *gin.Context) {
	var req dto.CreateReviewRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	authorID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
		return
	}

	if err := h.service.Create(authorID.(uint), req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "review created",
	})
}

func (h *ReviewHandler) GetByUser(c *gin.Context) {
	userIDParam := c.Param("id")

	userID, err := strconv.Atoi(userIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid user id",
		})
		return
	}

	reviews, err := h.service.GetByUserID(uint(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get reviews",
		})
		return
	}

	c.JSON(http.StatusOK, reviews)
}

func (h *ReviewHandler) GetByBook(c *gin.Context) {
	bookIDParam := c.Param("id")

	bookID, err := strconv.Atoi(bookIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid book id",
		})
		return
	}

	review, err := h.service.GetByBookID(uint(bookID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get review",
		})
		return
	}

	c.JSON(http.StatusOK, review)
}

func (h *ReviewHandler) Delete(c *gin.Context) {
	reviewIDParam := c.Param("id")

	reviewID, err := strconv.Atoi(reviewIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid review id",
		})
		return
	}

	authorID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
		return
	}

	if err := h.service.Delete(uint(reviewID), authorID.(uint)); err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "review deleted",
	})
}
