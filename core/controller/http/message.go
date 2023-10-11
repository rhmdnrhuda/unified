package http

import (
	"github.com/gin-gonic/gin"
	"github.com/rhmdnrhuda/unified/core/entity"
	"github.com/rhmdnrhuda/unified/core/usecase"
	"github.com/rhmdnrhuda/unified/pkg/logger"
	"net/http"
	"strings"
	"time"
)

type messageRoutes struct {
	uc  usecase.Message
	log logger.Interface
}

func NewMessageRoutes(handler *gin.RouterGroup, uc usecase.Message, l logger.Interface) {
	handler.Use(gin.Logger())
	handler.Use(gin.Recovery())

	r := messageRoutes{uc: uc, log: l}
	h := handler.Group("message")
	{
		h.POST("/webhook", r.receiveMessage)
		h.POST("", r.receiveMessage)
	}

	newHandler := handler.Group("payment")
	{
		newHandler.GET("/callback", r.callback)
	}

	cronHandler := handler.Group("cron")
	{
		cronHandler.GET("", r.cron)
	}

}

// @Summary Receive message
// @Description Receive new message in the system.
// @ID Message
// @Tags Message Handler
// @Accept json
// @Produce json
// @Param request body entity.MessageRequest true "The message request"
// @Success     200 {object} BaseResponse
// @Failure     500 {object} BaseResponse
// @Router      /message [post]
func (r *messageRoutes) receiveMessage(c *gin.Context) {
	var request entity.MessageRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		r.log.Error(err, "http - createTalentHandler - ShouldBindJSON")
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if !strings.EqualFold(request.EventType, "Message") {
		return
	}

	resp, err := r.uc.ProcessMessage(c.Request.Context(), request)
	if err != nil {
		r.log.Error(err, "http - receiveMessage - failed call use case")
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, BaseResponse{
		Code:       "200",
		Data:       resp,
		ServerTime: time.Now().Unix(),
	})
}

// @Summary Receive Payment Callback
// @Description Receive new message in the system.
// @ID Callback
// @Tags Payment Callback Handler
// @Accept json
// @Produce json
// @Param phone query string true "phone number"
// @Success     200 {object} BaseResponse
// @Failure     500 {object} BaseResponse
// @Router      /payment/callback [get]
func (r *messageRoutes) callback(c *gin.Context) {
	number := c.Query("phone")

	r.uc.PaymentCallback(c.Request.Context(), number)

	c.JSON(http.StatusOK, BaseResponse{
		Code:       "200",
		Data:       "Success",
		ServerTime: time.Now().Unix(),
	})
}

// @Summary Cron Alert
// @Description Run Cron Job For User Alert.
// @ID Cron Alert
// @Tags Cron Alert Handler
// @Accept json
// @Produce json
// @Success     200 {object} BaseResponse
// @Failure     500 {object} BaseResponse
// @Router      /cron [get]
func (r *messageRoutes) cron(c *gin.Context) {

	r.uc.RunCron(c.Request.Context())

	c.JSON(http.StatusOK, BaseResponse{
		Code:       "200",
		Data:       "Success",
		ServerTime: time.Now().Unix(),
	})
}
