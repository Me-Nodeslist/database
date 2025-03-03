package server

import (
	"net/http"
	"strconv"

	"github.com/Me-Nodeslist/database/database"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
)

type LicenseInfo struct {
	TokenID          string `json:"tokenID"`
	Owner            string `json:"owner"`
	Delegated        bool   `json:"delegated"`
	DelegatedNode    string `json:"delegatedNode"`
	TotalReward      string `json:"totalReward"`
	InitialReward    string `json:"initialReward"`
	WithdrawedReward string `json:"withdrawedReward"`
}

type LicenseInfos struct {
	Infos []LicenseInfo `json:"infos"`
}

// @Summary Get all license amount and delegated license amount
// @Description Get all license amount that have been sold, and all license amount that have been delegated
// @Tags License
// @Accept json
// @Produce json
// @Success 200 {object} map[string]int "return the amount"
// @Failure 500 {object} map[string]string "internal server error"
// @Router /license/amount [get]
func GetLicenseAmount() gin.HandlerFunc {
	return func(c *gin.Context) {
		amount, err := database.GetLicenseAmount()
		if err != nil {
			logger.Error(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		delegatedAmount, err := database.GetDelegatedLicenseAmount()
		if err != nil {
			logger.Error(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"amount":          amount,
			"delegatedAmount": delegatedAmount,
		})
	}
}

// @Summary Get all license amount of the owner
// @Description Get the license amount that the wallet address has purchased
// @Tags License
// @Accept json
// @Produce json
// @Param address path string true "owner address(an ethereum address with prefix '0x')""
// @Success 200 {object} map[string]int "return amount"
// @Failure 400 {object} map[string]string "request parameter error"
// @Failure 500 {object} map[string]string "internal server error"
// @Router /license/amount/owner/{address} [get]
func GetLicenseAmountOfOwner() gin.HandlerFunc {
	return func(c *gin.Context) {
		address := c.Param("address")
		owner := common.HexToAddress(address)
		amount, err := database.GetLicenseAmountByOwner(owner)
		if err != nil {
			logger.Error(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"amount": amount,
		})
	}
}

// @Summary Get the license information of the specified owner in pages
// @Description Query the license information owned by the owner through the wallet address, support paging
// @Tags License
// @Accept json
// @Produce json
// @Param address path string true "owner address(an ethereum address with prefix '0x')""
// @Param offset query int false "paging start index (default 0)"
// @Param limit query int false  "number of items to return per page(default 10)"
// @Success 200 {object}  LicenseInfos "return license info list successfully"
// @Failure 400 {object} map[string]string "request parameter error"
// @Failure 500 {object} map[string]string "internal server error"
// @Router /license/info/owner/{address} [get]
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
	}
}
