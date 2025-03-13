package server

import (
	"errors"
	"math/big"
	"net/http"

	"github.com/Me-Nodeslist/database/database"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RewardInfo struct {
	TotalLicenseRewards           string `json:"totalLicenseRewards" example:"1000000"`
	TotalWithdrawedLicenseRewards string `json:"totalWithdrawedLicenseRewards" example:"500000"`
	NodeReward                    string `json:"nodeReward" example:"200000"`
	WithdrawedNodeReward          string `json:"withdrawedNodeReward" example:"100000"`
}

type RedeemInfo struct {
	RedeemingDelMEMOAmount string   `json:"redeemingDelMEMOAmount" example:"1000"`
	LockedMEMOAmount       string   `json:"lockedMEMOAmount" example:"500"`
	UnlockedMEMOAmount     string   `json:"unlockedMEMOAmount" example:"1500"`
	WithdrawedMEMOAmount   string   `json:"withdrawedMEMOAmount" example:"800"`
	UnclaimedRedeemIDs     []string `json:"unclaimedRedeemIDs" example:"['1', '2']"`
}

// @Summary Get the account's reward information
// @Description Query all license rewards and node reward information of the specific owner, include total and withdrawed
// @Tags Reward
// @Accept json
// @Produce json
// @Param address path string true "owner address(an ethereum address with prefix '0x')"
// @Success 200 {object} RewardInfo "return reward information successfully"
// @Failure 400 {object} map[string]string "request parameter error"
// @Failure 500 {object} map[string]string "internal server error"
// @Router /reward/info/{address} [get]
func GetRewardInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		address := c.Param("address")
		owner := common.HexToAddress(address)

		totalDelegationReward := big.NewInt(0)
		totalWithdrawedDelegationReward := big.NewInt(0)

		// get all licenses
		amount, err := database.GetLicenseAmountByOwner(owner)
		if err != nil {
			logger.Error(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		if amount > 0 {
			infos, err := database.GetLicenseInfosByOwner(owner, 0, int(amount))
			if err != nil {
				logger.Error(err.Error())
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}

			// get all delegation rewards
			for i := 0; i < int(amount); i++ {
				reward, ok := new(big.Int).SetString(infos[i].TotalReward, 10)
				if !ok {
					logger.Error(infos[i].TotalReward, errors.New("transfer string to big.Int error"))
					c.JSON(http.StatusInternalServerError, gin.H{
						"error": "transfer string to big.Int error",
					})
					return
				}
				totalDelegationReward.Add(totalDelegationReward, reward)
				reward, ok = new(big.Int).SetString(infos[i].WithdrawedReward, 10)
				if !ok {
					logger.Error(infos[i].WithdrawedReward, errors.New("transfer string to big.Int error"))
					c.JSON(http.StatusInternalServerError, gin.H{
						"error": "transfer string to big.Int error",
					})
					return
				}
				totalWithdrawedDelegationReward.Add(totalWithdrawedDelegationReward, reward)
			}
		}

		// get node rewards
		nodeReward := "0"
		withdrawedNodeReward := "0"
		nodeInfo, err := database.GetNodeInfoByNodeAddress(owner)
		if err != nil && err != gorm.ErrRecordNotFound {
			logger.Error(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		if err == nil {
			nodeReward = nodeInfo.SelfTotalReward
			withdrawedNodeReward = nodeInfo.SelfWithdrawedReward
		}

		// return all rewards and all withdrawed rewards(withdraw delMEMO)
		c.JSON(http.StatusOK, gin.H{
			"totalLicenseRewards":           totalDelegationReward.String(),
			"totalWithdrawedLicenseRewards": totalWithdrawedDelegationReward.String(),
			"nodeReward":                    nodeReward,
			"withdrawedNodeReward":          withdrawedNodeReward,
		})
	}
}

// @Summary Get the account's redeem information
// @Description Query all locked, unlocked, withdrawed MEMOs and redeeming DelMEMOs of specific owner, as well as unclaimed redeemIDs
// @Tags Redeem
// @Accept json
// @Produce json
// @Param address path string true "owner address(an ethereum address with prefix '0x')"
// @Success 200 {object} RedeemInfo "return redeem information successfully"
// @Failure 400 {object} map[string]string "request parameter error"
// @Failure 500 {object} map[string]string "internal server error"
// @Router /reward/redeem/info/{address} [get]
func GetRedeemInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		address := c.Param("address")
		owner := common.HexToAddress(address)

		// get redeeming delMemo amount
		redeemingAmount, err := database.GetRedeemingAmountByInitiator(owner)
		if err != nil {
			logger.Error(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		// get locked Memo amount
		lockedMemoAmount, err := database.GetLockedAmountByInitiator(owner)
		if err != nil {
			logger.Error(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		// get unlocked Memo amount
		unlockedMemoAmount, err := database.GetUnClaimedAmountByInitiator(owner)
		if err != nil {
			logger.Error(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		// get withdrawed Memo amount
		withdrawedMemoAmount, err := database.GetClaimedAmountByInitiator(owner)
		if err != nil {
			logger.Error(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		// get unclaimed redeemIDs
		redeemIDs, err := database.GetUnClaimedRedeemIDsByInitiator(owner)
		if err != nil {
			logger.Error(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		// return all rewards and all withdrawed rewards(withdraw delMEMO)
		c.JSON(http.StatusOK, gin.H{
			"redeemingDelMEMOAmount": redeemingAmount.String(),
			"lockedMEMOAmount":       lockedMemoAmount.String(),
			"unlockedMEMOAmount":     unlockedMemoAmount.String(),
			"withdrawedMEMOAmount":   withdrawedMemoAmount.String(),
			"unclaimedRedeemIDs":     redeemIDs,
		})
	}
}
