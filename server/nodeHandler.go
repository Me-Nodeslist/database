package server

import (
	"net/http"
	"strconv"

	"github.com/Me-Nodeslist/database/database"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
)

type NodeInfo struct {
	NodeID                     uint32 `json:"nodeID"`
	NodeAddress                string `json:"nodeAddress"`
	Recipient                  string `json:"recipient"`
	Active                     bool `json:"active"`
	CommissionRate             uint8 `json:"commissionRate"`
	DelegationAmount           uint16 `json:"delegationAmount"`
	SelfTotalReward            string `json:"selfTotalReward"`
	SelfWithdrawedReward       string `json:"selfWithdrawedReward"`
	DelegationReward           string `json:"delegationReward"`
	CommissionRateLastModifyAt string `json:"commissionRateLastModifyAt"`
	RegisterDate               string `json:"registerDate"`
	OnlineDays                 int64 `json:"onlineDays"`
	OnlineDays_RecentMonth     int64 `json:"onlineDays_RecentMonth"`
	OnlineDays_RecentWeek      int64 `json:"onlineDays_RecentWeek"`
}

type NodeInfos struct {
	Infos []NodeInfo `json:"infos"`
}

// @Summary Get the amount of all registered nodes
// @Description Query the amount of the registered nodes in nodelist server
// @Tags Node
// @Accept json
// @Produce json
// @Success 200 {object} map[string]int "return the nodes amount successfully"
// @Failure 500 {object} map[string]string "internal server error"
// @Router /node/amount [get]
func GetNodeAmount() gin.HandlerFunc {
	return func(c *gin.Context) {
		amount, err := database.GetNodeAmount()
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

// @Summary Get all nodes information(paging)
// @Description Query the information of all nodes, support paging
// @Tags Node
// @Accept json
// @Produce json
// @Param offset query int false "paging start index (default 0)"
// @Param limit query int false "number of items to return per page(default 10)"
// @Success 200 {object} NodeInfos "return node info list successfully"
// @Failure 400 {object} map[string]string "request parameter error"
// @Failure 500 {object} map[string]string "internal server error"
// @Router /node/info [get]
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
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"infos": infos,
		})
	}
}

// @Summary Get the node information of the owner
// @Description Query the node information through the owner address who owned the node
// @Tags Node
// @Accept json
// @Produce json
// @Param address path string true "owner address(an ethereum address with prefix '0x')"
// @Success 200 {object} NodeInfo "return the node information successfully"
// @Failure 500 {object} map[string]string "internal server error"
// @Router /node/info/owner/{address} [get]
func GetNodeInfoOfOwner() gin.HandlerFunc {
	return func(c *gin.Context) {
		address := c.Param("address")
		owner := common.HexToAddress(address)

		info, err := database.GetNodeInfoByNodeAddress(owner)
		if err != nil {
			logger.Error(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"info": info,
		})
	}
}

// @Summary Get the nodes information by the recipient
// @Description Query the nodes information list through the recipient address who receives the node reward
// @Tags Node
// @Accept json
// @Produce json
// @Param address path string true "recipient address(an ethereum address with prefix '0x')"
// @Param offset query int false "paging start index (default 0)"
// @Param limit query int false "number of items to return per page(default 10)"
// @Success 200 {object} NodeInfos "return the nodes information list successfully"
// @Failure 500 {object} map[string]string "internal server error"
// @Router /node/info/recipient/{address} [get]
func GetNodeInfosOfRecipient() gin.HandlerFunc {
	return func(c *gin.Context) {
		address := c.Param("address")
		recipient := common.HexToAddress(address)

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

		info, err := database.GetNodeInfosByRecipient(recipient, offset, limit)
		if err != nil {
			logger.Error(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"info": info,
		})
	}
}

// @Summary Get the nodes information delegated by a specific owner
// @Description Query the information of the nodes that the specific owner has delegated his licenses to
// @Tags Node
// @Accept json
// @Produce json
// @Param address path string true "owner address(an ethereum address with prefix '0x')"
// @Success 200 {object} NodeInfos "return delegated nodes information successfully"
// @Failure 400 {object} map[string]string "request parameter error"
// @Failure 500 {object} map[string]string "internal server error"
// @Router /node/info/delegation/{address} [get]
func GetNodeInfosOfdelegation() gin.HandlerFunc {
	return func(c *gin.Context) {
		address := c.Param("address")
		owner := common.HexToAddress(address)

		// get license amount
		amount, err := database.GetLicenseAmountByOwner(owner)
		if err != nil {
			logger.Error(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		// get license infos
		infos, err := database.GetLicenseInfosByOwner(owner, 0, int(amount))
		if err != nil {
			logger.Error(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		// get delegated nodeInfo of license
		nodeInfos := make([]database.NodeInfo,0)
		for i:=0;i<int(amount);i++{
			if infos[i].Delegated {
				delegatedNode := infos[i].DelegatedNode
				nodeInfo, err := database.GetNodeInfoByNodeAddress(common.HexToAddress(delegatedNode))
				if err != nil {
					logger.Error(err.Error())
					c.JSON(http.StatusInternalServerError, gin.H{
						"error": err.Error(),
					})
					return
				}
				nodeInfos = append(nodeInfos, nodeInfo)
			}
		}
		c.JSON(http.StatusOK, gin.H{
			"infos": nodeInfos,
		})
	}
}
