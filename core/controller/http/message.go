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
