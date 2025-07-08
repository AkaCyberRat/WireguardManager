package handlers

import (
	"net/http"

	"WireguardManager/internal/core"
	"WireguardManager/internal/services"
	"WireguardManager/internal/transport/rest"

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
		rest.NewErrorResponse(c, http.StatusBadRequest, rest.ErrImpossibleToBindModel)

		return
	}

	peer, err := h.peerService.Get(c.Request.Context(), &model)
	if err != nil {
		if err == core.ErrModelValidation {
			rest.NewErrorResponse(c, http.StatusBadRequest, err)

			return
		}

		if err == core.ErrPeerNotFound {
			rest.NewErrorResponse(c, http.StatusNotFound, err)

			return
		}

		logrus.Error("Get peer internal error: ", err.Error())
		rest.NewErrorResponse(c, http.StatusInternalServerError, rest.ErrInternalServer)

		return
	}

	c.JSON(http.StatusOK, peer)
}

// POST peer
func (h *PeerHandler) Create(c *gin.Context) {

	var model core.CreatePeer
	if err := c.BindJSON(&model); err != nil {
		rest.NewErrorResponse(c, http.StatusBadRequest, rest.ErrImpossibleToBindModel)
		return
	}

	peer, err := h.peerService.Create(c.Request.Context(), &model)
	if err != nil {
		if err == core.ErrModelValidation {
			rest.NewErrorResponse(c, http.StatusBadRequest, err)

			return
		}

		if err == core.ErrPeerLimitReached {
			rest.NewErrorResponse(c, http.StatusForbidden, err)
			return
		}

		logrus.Error("Create peer internal error: ", err.Error())
		rest.NewErrorResponse(c, http.StatusInternalServerError, rest.ErrInternalServer)
		return
	}

	c.JSON(http.StatusOK, peer)
}

// PATCH peer
func (h *PeerHandler) Update(c *gin.Context) {

	var model core.UpdatePeer
	if err := c.BindJSON(&model); err != nil {
		rest.NewErrorResponse(c, http.StatusBadRequest, rest.ErrImpossibleToBindModel)
		return
	}

	peer, err := h.peerService.Update(c.Request.Context(), &model)
	if err != nil {
		if err == core.ErrModelValidation {
			rest.NewErrorResponse(c, http.StatusBadRequest, err)

			return
		}

		if err == core.ErrPeerNotFound {
			rest.NewErrorResponse(c, http.StatusNotFound, err)

			return
		}

		logrus.Error("Update peer internal error: ", err.Error())
		rest.NewErrorResponse(c, http.StatusInternalServerError, rest.ErrInternalServer)
		return
	}

	c.JSON(http.StatusOK, peer)
}

// DELETE peer
func (h *PeerHandler) Delete(c *gin.Context) {

	var model core.DeletePeer
	if err := c.BindJSON(&model); err != nil {
		rest.NewErrorResponse(c, http.StatusBadRequest, rest.ErrImpossibleToBindModel)
		return
	}

	err := h.peerService.Delete(c.Request.Context(), &model)
	if err != nil {
		if err == core.ErrModelValidation {
			rest.NewErrorResponse(c, http.StatusBadRequest, err)

			return
		}

		if err == core.ErrPeerNotFound {
			rest.NewErrorResponse(c, http.StatusNotFound, err)

			return
		}

		logrus.Error("Delete peer internal error: ", err.Error())
		rest.NewErrorResponse(c, http.StatusInternalServerError, rest.ErrInternalServer)
		return
	}

	c.Status(http.StatusOK)
}
