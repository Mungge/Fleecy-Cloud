package handlers

import (
	"net/http"

	"github.com/Mungge/Fleecy-Cloud/models"
	"github.com/gin-gonic/gin"
)

// POST /aggregator/estimate
func EstimateHandler(c *gin.Context){
	var input models.UserInput
	if err:=c.ShouldBindJSON(&input); err!=nil{
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	estimation := models.EstimateResources(input)
	c.JSON(http.StatusOK, estimation)
}

// POST /aggregator/recommend
func RecommendHandler(c *gin.Context) {
	var input models.UserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	estimation := models.EstimateResources(input)
	recommendations := models.RecommendInstance(estimation)

	c.JSON(http.StatusOK, gin.H{
		"estimate":        estimation,
		"recommendations": recommendations,
	})
}