package handlers

import (
	"fmt"
	"net/http"

	"github.com/Mungge/Fleecy-Cloud/services"
	"github.com/gin-gonic/gin"
)

type SSHKeypairHandler struct {
	service *services.SSHKeypairService
}

func NewSSHKeypairHandler(service *services.SSHKeypairService) *SSHKeypairHandler {
	return &SSHKeypairHandler{service: service}
}

// GetKeypairByAggregatorID 집계자 ID로 SSH 키페어 조회 (Private Key 제외)
// GET /api/keypairs/aggregator/:aggregatorId
func (h *SSHKeypairHandler) GetKeypairByAggregatorID(c *gin.Context) {
	aggregatorID := c.Param("aggregatorId")
	if aggregatorID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "aggregator ID is required",
		})
		return
	}

	keypair, err := h.service.GetKeypairByAggregatorID(aggregatorID)
	if err != nil {
		if err.Error() == "keypair not found for aggregator: "+aggregatorID {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "keypair not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get keypair",
		})
		return
	}

	c.JSON(http.StatusOK, keypair)
}

// DownloadPrivateKey Private Key 다운로드
// GET /api/keypairs/aggregator/:aggregatorId/private-key
func (h *SSHKeypairHandler) DownloadPrivateKey(c *gin.Context) {
	aggregatorID := c.Param("aggregatorId")
	if aggregatorID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "aggregator ID is required",
		})
		return
	}

	keypairWithPrivateKey, err := h.service.GetKeypairWithPrivateKey(aggregatorID)
	if err != nil {
		if err.Error() == "keypair not found for aggregator: "+aggregatorID {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "keypair not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get private key",
		})
		return
	}

	// Private Key를 파일로 다운로드
	filename := keypairWithPrivateKey.KeyName + ".pem"
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/x-pem-file")
	c.String(http.StatusOK, keypairWithPrivateKey.PrivateKey)
}

// ListKeypairsByUser 사용자의 모든 키페어 목록 조회
// GET /api/keypairs/user/:userId
func (h *SSHKeypairHandler) ListKeypairsByUser(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "user ID is required",
		})
		return
	}

	// userID를 int64로 변환
	var userIDInt int64
	if _, err := fmt.Sscanf(userID, "%d", &userIDInt); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid user ID format",
		})
		return
	}

	keypairs, err := h.service.ListKeypairsByUserID(userIDInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to list keypairs",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": keypairs,
	})
}

// DeleteKeypairByAggregatorID 집계자 ID로 키페어 삭제
// DELETE /api/keypairs/aggregator/:aggregatorId
func (h *SSHKeypairHandler) DeleteKeypairByAggregatorID(c *gin.Context) {
	aggregatorID := c.Param("aggregatorId")
	if aggregatorID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "aggregator ID is required",
		})
		return
	}

	err := h.service.DeleteKeypairByAggregatorID(aggregatorID)
	if err != nil {
		if err.Error() == "keypair not found for aggregator: "+aggregatorID {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "keypair not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to delete keypair",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "keypair deleted successfully",
	})
}
