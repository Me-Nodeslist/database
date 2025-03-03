package server

import (
	"net/http"
	"strconv"

	"github.com/Me-Nodeslist/database/database"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
)

func GetLicenseAmount() gin.HandlerFunc {
	return func(c *gin.Context) {
		amount, err := database.GetLicenseAmount()
		if err != nil {
			logger.Error(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": c.Errors[0].Error(),
			})
			return
		}
		delegatedAmount, err := database.GetDelegatedLicenseAmount()
		if err != nil {
			logger.Error(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": c.Errors[0].Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"amount":          amount,
			"delegatedAmount": delegatedAmount,
		})
		return
	}
}

func GetLicenseAmountOfOwner() gin.HandlerFunc {
	return func(c *gin.Context) {
		address := c.Param("address")
		owner := common.HexToAddress(address)
		amount, err := database.GetLicenseAmountByOwner(owner)
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

func GetLicenseInfosOfOwner() gin.HandlerFunc {
	return func(c *gin.Context) {
		address := c.Param("address")
		offsetStr := c.Query("offset")
		limitStr := c.Query("limit")

		owner := common.HexToAddress(address)
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

		infos, err := database.GetLicenseInfosByOwner(owner, offset, limit)
		if err != nil {
			logger.Error(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"infos": infos,
		})
		return
	}
}
