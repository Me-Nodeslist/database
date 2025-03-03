package server

import (
	"net/http"
	"strconv"

	"github.com/Me-Nodeslist/database/database"
	"github.com/gin-gonic/gin"
)

func GetNodeAmount() gin.HandlerFunc {
	return func(c *gin.Context) {
		amount, err := database.GetNodeAmount()
		if err != nil {
			logger.Error(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": c.Errors[0].Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"amount": amount,
		})
		return
	}
}

func GetNodeInfos() gin.HandlerFunc {
	return func(c *gin.Context) {
		offsetStr := c.Query("offset")
		limitStr := c.Query("limit")

		offset, err := strconv.Atoi(offsetStr)
		if err != nil {
			logger.Error(err.Error())
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			logger.Error(err.Error())
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		infos, err := database.GetNodeInfos(offset, limit)
		if err != nil {
			logger.Error(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": c.Errors[0].Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"infos": infos,
		})
		return
	}
}
