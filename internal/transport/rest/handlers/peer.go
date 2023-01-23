package handlers

import (
	"net/http"

	"WireguardManager/internal/core"
	"WireguardManager/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type PeerHandler struct {
	peerService services.PeerService
}

func NewPeerHandler(s services.PeerService) *PeerHandler {
	return &PeerHandler{peerService: s}
}

// ShowAccount godoc
// @Summary      Show a peer
// @Description  get peer by Id
// @Tags         peer
// @Accept       json
// @Produce      json
// @Success      200  {object}  core.Peer
// @Failure      400  {object}  handlers.ErrorResponse
// @Failure      401  {object}  handlers.ErrorResponse
// @Failure      403  {object}  handlers.ErrorResponse
// @Failure      404  {object}  handlers.ErrorResponse
// @Failure      500  {object}  handlers.ErrorResponse
// @Router       /api/peer [get]
func (h *PeerHandler) Get(c *gin.Context) {

	var model core.GetPeer
	if err := c.BindJSON(&model); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err)

		return
	}

	peer, err := h.peerService.Get(c.Request.Context(), &model)
	if err != nil {
		if err == core.ErrModelValidation {
			newErrorResponse(c, http.StatusBadRequest, err)

			return
		}

		if err == core.ErrPeerNotFound {
			newErrorResponse(c, http.StatusNotFound, err)

			return
		}

		logrus.Error("Get peer internal error: ", err.Error())
		newErrorResponse(c, http.StatusInternalServerError, ErrInternalServer)

		return
	}

	newErrorResponse(c, http.StatusOK, peer)
}

// POST peer
func (h *PeerHandler) Create(c *gin.Context) {

	var model core.CreatePeer
	if err := c.BindJSON(&model); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	peer, err := h.peerService.Create(c.Request.Context(), &model)
	if err != nil {
		if err == core.ErrModelValidation {
			newErrorResponse(c, http.StatusBadRequest, err)

			return
		}

		if err == core.ErrPeerLimitReached {
			newErrorResponse(c, http.StatusForbidden, err)
			return
		}

		logrus.Error("Create peer internal error: ", err.Error())
		newErrorResponse(c, http.StatusInternalServerError, ErrInternalServer)
		return
	}

	c.JSON(http.StatusOK, peer)
}

// PATCH peer
func (h *PeerHandler) Update(c *gin.Context) {

	var model core.UpdatePeer
	if err := c.BindJSON(&model); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	peer, err := h.peerService.Update(c.Request.Context(), &model)
	if err != nil {
		if err == core.ErrModelValidation {
			newErrorResponse(c, http.StatusBadRequest, err)

			return
		}

		if err == core.ErrPeerNotFound {
			newErrorResponse(c, http.StatusNotFound, err)

			return
		}

		logrus.Error("Update peer internal error: ", err.Error())
		newErrorResponse(c, http.StatusInternalServerError, ErrInternalServer)
		return
	}

	c.JSON(http.StatusOK, peer)
}

// DELETE peer
func (h *PeerHandler) Delete(c *gin.Context) {

	var model core.DeletePeer
	if err := c.BindJSON(&model); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	err := h.peerService.Delete(c.Request.Context(), &model)
	if err != nil {
		if err == core.ErrModelValidation {
			newErrorResponse(c, http.StatusBadRequest, err)

			return
		}

		if err == core.ErrPeerNotFound {
			newErrorResponse(c, http.StatusNotFound, err)

			return
		}

		logrus.Error("Delete peer internal error: ", err.Error())
		newErrorResponse(c, http.StatusInternalServerError, ErrInternalServer)
		return
	}

	c.Status(http.StatusOK)
}
