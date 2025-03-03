package server

import (
	"errors"
	"math/big"
	"net/http"

	"github.com/Me-Nodeslist/database/database"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
)

type RewardInfo struct {
	TotalLicenseRewards          *big.Int `json:"totalLicenseRewards" example:"1000000"`
	TotalWithdrawedLicenseRewards *big.Int `json:"totalWithdrawedLicenseRewards" example:"500000"`
	NodeReward                   *big.Int `json:"nodeReward" example:"200000"`
	WithdrawedNodeReward         *big.Int `json:"withdrawedNodeReward" example:"100000"`
}

type RedeemInfo struct {
	RedeemingDelMEMOAmount   *big.Int   `json:"redeemingDelMEMOAmount" example:"1000"`
	LockedMEMOAmount         *big.Int   `json:"lockedMEMOAmount" example:"500"`
	UnlockedMEMOAmount       *big.Int   `json:"unlockedMEMOAmount" example:"1500"`
	WithdrawedMEMOAmount     *big.Int   `json:"withdrawedMEMOAmount" example:"800"`
	UnclaimedRedeemIDs       []string   `json:"unclaimedRedeemIDs" example:"['1', '2']"`
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

		// get all licenses
		amount, err := database.GetLicenseAmountByOwner(owner)
		if err != nil {
			logger.Error(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		infos, err := database.GetLicenseInfosByOwner(owner, 0, int(amount))
		if err != nil {
			logger.Error(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		// get all delegation rewards
		totalDelegationReward := big.NewInt(0)
		totalWithdrawedDelegationReward := big.NewInt(0)
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

		// get node rewards
		nodeInfo, err := database.GetNodeInfoByNodeAddress(owner)
		if err != nil {
			logger.Error(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		// return all rewards and all withdrawed rewards(withdraw delMEMO)
		c.JSON(http.StatusOK, gin.H{
			"totalLicenseRewards":           totalDelegationReward,
			"totalWithdrawedLicenseRewards": totalWithdrawedDelegationReward,
			"nodeReward":                    nodeInfo.SelfTotalReward,
			"withdrawedNodeReward":          nodeInfo.SelfWithdrawedReward,
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
			"redeemingDelMEMOAmount": redeemingAmount,
			"lockedMEMOAmount":       lockedMemoAmount,
			"unlockedMEMOAmount":     unlockedMemoAmount,
			"withdrawedMEMOAmount":   withdrawedMemoAmount,
			"unclaimedRedeemIDs":     redeemIDs,
		})
	}
}
