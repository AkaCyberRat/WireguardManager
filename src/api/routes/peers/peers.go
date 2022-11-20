package peers

import (
	"WireguardManager/src/api/routes/peers/viewmodels"
	"WireguardManager/src/db/models"
	"WireguardManager/src/services/vpn"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func headPeersEndpoint(ctx *gin.Context) {
	count := vpn.GetPeersCountByStatus([]models.Status{models.Disabled, models.Enabled})
	ctx.Header("X-Total-Count", fmt.Sprint(count))

	ctx.Status(http.StatusNoContent)
}

func getPeersEndpoint(ctx *gin.Context) {
	offset, err1 := strconv.Atoi(ctx.DefaultQuery("offset", "0"))
	limit, err2 := strconv.Atoi(ctx.DefaultQuery("limit", "50"))

	count := vpn.GetPeersCountByStatus([]models.Status{models.Disabled, models.Enabled})
	ctx.Header("X-Total-Count", fmt.Sprint(count))

	if err1 != nil || err2 != nil || offset >= count || limit < 1 {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": false, "message": "wrong query param"})
		return
	}

	respModels := []viewmodels.PeerResponse{}
	peers := vpn.GetPeersByStatus([]models.Status{models.Enabled, models.Disabled}, offset, limit)

	for _, peer := range peers {
		respModel := new(viewmodels.PeerResponse).BindModel(peer)
		respModels = append(respModels, *respModel)
	}

	b, err := json.Marshal(respModels)
	if err != nil {
		panic(err)
	}

	ctx.JSON(http.StatusOK, gin.H{"status": false, "data": string(b)})
}

func getPeerEndpoint(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": false, "message": "wrong id param"})
		return
	}

	peer, err := vpn.GetPeerById(id)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"status": false, "message": "wrong id param"})
		return
	}

	respModel := new(viewmodels.PeerResponse).BindModel(peer)
	b, err := json.Marshal(*respModel)
	if err != nil {
		panic(err)
	}

	ctx.JSON(http.StatusOK, gin.H{"status": false, "data": string(b)})
}

func addPeerEndpoint(ctx *gin.Context) {
	reqModel := viewmodels.PeerAddRequest{}
	if err := ctx.ShouldBindJSON(&reqModel); err != nil {
		panic(err)
	}

	if validateModel(reqModel) {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": false, "message": "Model validation error"})
		return
	}

	if vpn.GetUsedPeersCount() < vpn.PeersLimit() {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": false, "message": "there are no free peers"})
		return
	}

	peer := vpn.GetPeersByStatus([]models.Status{models.Unused}, 0, 1)[0]
	reqModel.BindPeer(&peer)
	vpn.UpdatePeer(peer)

	respModel := new(viewmodels.PeerResponse).BindModel(peer)
	b, err := json.Marshal(*respModel)
	if err != nil {
		panic(err)
	}

	ctx.JSON(http.StatusOK, gin.H{"status": false, "data": string(b)})
}

func removePeerEndpoint(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": false, "message": "wrong id param"})
		return
	}

	err = vpn.DropPeerById(id)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"status": false, "message": "wrong id param"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": true})
}

func updatePeerEndpoint(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": false, "message": "wrong id param"})
		return
	}

	reqModel := viewmodels.PeerPatchRequest{}
	if err := ctx.ShouldBindJSON(&reqModel); err != nil {
		panic(err)
	}

	if validateModel(reqModel) {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": false, "message": "model validation error"})
		return
	}

	if peer, err := vpn.GetPeerById(id); err != nil || peer.Status != models.Unused {
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"status": false, "message": "there are no such peer"})
		return
	}

	peer := models.Peer{}
	reqModel.BindPeer(&peer)
	vpn.UpdatePeer(peer)

	respModel := new(viewmodels.PeerResponse).BindModel(peer)
	b, err := json.Marshal(*respModel)
	if err != nil {
		panic(err)
	}

	ctx.JSON(http.StatusOK, gin.H{"status": false, "data": string(b)})
}

func validateModel(model interface{}) bool {
	err := validator.New().Struct(model)
	return err == nil && viewmodels.HasNotNullField(model)
}
