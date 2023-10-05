package http

import (
	"github.com/gin-gonic/gin"
	"github.com/temukan-co/monolith/config"
	"github.com/temukan-co/monolith/core/entity"
	"github.com/temukan-co/monolith/core/usecase"
	"github.com/temukan-co/monolith/pkg/logger"
	"net/http"
	"time"
)

type talentRoutes struct {
	uc  usecase.Talent
	log logger.Interface
	cfg *config.Config
}

func NewTalentRoutes(handler *gin.RouterGroup, uc usecase.Talent, l logger.Interface, cfg *config.Config) {
	handler.Use(gin.Logger())
	handler.Use(gin.Recovery())

	r := talentRoutes{uc, l, cfg}
	h := handler.Group("talent")
	{
		h.POST("/create", r.createTalentHandler)
		h.PUT("/update", r.updateTalentHandler)
	}
}

// All godoc
// @Tags Unified
// @Summary Unified
// @Description Put all mandatory parameter
// @Param channelId header string true "WEB" default(WEB)
// @Param customerSessionId header string true "d41d8cd98f00b204e9800998ecf8427e" default(d41d8cd98f00b204e9800998ecf8427)
// @Param lang header string true "en" default(en)
// @Param requestId header string true "23123123" default(23123123)
// @Param serviceId header string true "gateway" default(gateway)
// @Param username header string true "username" default(username)
// @Param version header string true "version" default(1)
// @Param customerUserAgent header string true "Chrome" default(Chrome)
// @Param customerIPAddress header string true "192.168.1.1" default(192.168.1.1)
// @Param request body hotel.SearchRequestParamDto true "SearchRequestParamDto"
// @Param X-Loyalty-Level header string false "X-Loyalty-Level"
// @Param X-Account-Id header string false "X-Account-Id"
// @Param X-Identity header string false "X-Identity"
// @Param Authorization header string false "Authorization"
// @Accept  json
// @Produce  json
// @Success 200 {object} historyResponse
// @Failure 500 {object} BaseResponse
// @Router /translation/history [post]

// @Summary Create a new talent
// @Description Creates a new talent in the system.
// @ID createTalent
// @Tags Talent Handler
// @Accept json
// @Produce json
// @Param request body entity.TalentRequest true "The talent request"
// @Success     200 {object} BaseResponse
// @Failure     500 {object} BaseResponse
// @Router      /talent/create [post]
func (r *talentRoutes) createTalentHandler(c *gin.Context) {
	var request entity.TalentRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		r.log.Error(err, "http - createTalentHandler - ShouldBindJSON")
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	err = r.uc.Create(c.Request.Context(), request)
	if err != nil {
		r.log.Error(err, "http - createTalentHandler - failed call use case")
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, BaseResponse{
		Code:       "200",
		Data:       "success",
		ServerTime: time.Now().Unix(),
	})
}

// @Summary Update an existing talent
// @Description Updates an existing talent in the system.
// @ID updateTalent
// @Tags Talent Handler
// @Accept json
// @Produce json
// @Param request body entity.TalentRequest true "The talent request"
// @Success     200 {object} BaseResponse
// @Failure     500 {object} BaseResponse
// @Router      /talent/update [put]
func (r *talentRoutes) updateTalentHandler(c *gin.Context) {
	var request entity.TalentRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		r.log.Error(err, "http- updateTalentHandler - ShouldBindJSON")
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	err = r.uc.Update(c.Request.Context(), request)
	if err != nil {
		r.log.Error(err, "http - updateTalentHandler - failed call use case")
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, BaseResponse{
		Code:       "200",
		Data:       "success",
		ServerTime: time.Now().Unix(),
	})
}
