package handler

import (
	"net/http"
	"shorturl/service"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type ShortURLHandler struct {
	shortURLService *service.ShortURLService
	statsService    *service.StatsService
}

func NewShortURLHandler(shortURLService *service.ShortURLService, statsService *service.StatsService) *ShortURLHandler {
	return &ShortURLHandler{
		shortURLService: shortURLService,
		statsService:    statsService,
	}
}

type CreateRequest struct {
	URL      string `json:"url" binding:"required,url"`
	ExpireAt int64  `json:"expire_at,omitempty"`
}

type CreateResponse struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	ShortCode string `json:"short_code"`
	ShortURL  string `json:"short_url"`
}

func (h *ShortURLHandler) Create(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "invalid request: " + err.Error(),
		})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "user not authenticated",
		})
		return
	}

	var expireAt *time.Time
	if req.ExpireAt > 0 {
		t := time.Unix(req.ExpireAt, 0)
		expireAt = &t
	}

	ctx := c.Request.Context()
	shortCode, err := h.shortURLService.CreateShortURL(ctx, userID.(uint64), req.URL, expireAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "create short url failed: " + err.Error(),
		})
		return
	}

	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	shortURL := scheme + "://" + c.Request.Host + "/s/" + shortCode

	c.JSON(http.StatusOK, CreateResponse{
		Code:      0,
		Message:   "success",
		ShortCode: shortCode,
		ShortURL:  shortURL,
	})
}

func (h *ShortURLHandler) List(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "user not authenticated",
		})
		return
	}

	page := 1
	if p := c.Query("page"); p != "" {
		if n, err := strconv.Atoi(p); err == nil {
			page = n
		}
	}

	pageSize := 10
	if ps := c.Query("page_size"); ps != "" {
		if n, err := strconv.Atoi(ps); err == nil {
			pageSize = n
		}
	}

	ctx := c.Request.Context()
	list, total, err := h.shortURLService.ListShortURLs(ctx, userID.(uint64), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "list short urls failed: " + err.Error(),
		})
		return
	}

	for i := range list {
		list[i].OriginalURL = ""
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"list":  list,
			"total": total,
			"page":  page,
		},
	})
}

func (h *ShortURLHandler) Delete(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "user not authenticated",
		})
		return
	}

	shortCode := c.Param("code")
	if shortCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "short code is required",
		})
		return
	}

	ctx := c.Request.Context()
	if err := h.shortURLService.DeleteShortURL(ctx, shortCode, userID.(uint64)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "delete short url failed: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

func (h *ShortURLHandler) Redirect(c *gin.Context) {
	shortCode := c.Param("code")
	if shortCode == "" {
		c.Status(http.StatusNotFound)
		return
	}

	ctx := c.Request.Context()
	originalURL, err := h.shortURLService.GetOriginalURL(ctx, shortCode)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	if originalURL == "" {
		c.Status(http.StatusNotFound)
		return
	}

	h.statsService.RecordVisit(service.VisitEvent{
		ShortCode: shortCode,
		IP:        c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
		Referer:   c.GetHeader("Referer"),
	})

	c.Redirect(http.StatusFound, originalURL)
}
