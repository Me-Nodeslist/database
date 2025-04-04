package server

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/Me-Nodeslist/database/database"
	"github.com/Me-Nodeslist/database/dumper"
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

type LicensePrice struct {
	Usdt string // xxxUSDT/1License
	Eth  string // xxxETH/1License
}

type MintRequest struct {
	Receiver string
	Amount   int64
	Value    string // pay how many wei
	TxHash   string // the transaction hash that receiver transfer eth to admin
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
			"amount":          amount + 1542,
			"delegatedAmount": delegatedAmount + 536,
		})
	}
}

// @Summary Get all license amount of the owner
// @Description Get the license amount that the wallet address has purchased
// @Tags License
// @Accept json
// @Produce json
// @Param address path string true "owner address(an ethereum address with prefix '0x')""
// @Success 200 {object} map[string]int "return amount and delegated amount"
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
		delamount, err := database.GetDelegatedLicenseAmountByOwner(owner)
		if err != nil {
			logger.Error(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"amount": amount,
			"delegatedAmount": delamount,
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

// @Summary Get license price
// @Description Get license price, include how many USDT and how many ETH
// @Tags License
// @Accept json
// @Produce json
// @Success 200 {object} LicensePrice "return the price"
// @Failure 500 {object} map[string]string "internal server error"
// @Router /license/price [get]
func GetLicensePrice() gin.HandlerFunc {
	return func(c *gin.Context) {
		licensePrice_usdt := float64(dumper.LICENSE_PRICE_USDT)

		ethusdt, err := getEthPrice()
		if err != nil {
			logger.Error(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		ethAmount := getEthAmount(licensePrice_usdt, ethusdt)

		licensePrice := LicensePrice{
			Usdt: fmt.Sprintf("%.6f", licensePrice_usdt),
			Eth:  fmt.Sprintf("%.6f", ethAmount),
		}

		c.JSON(http.StatusOK, gin.H{
			"price": licensePrice,
		})
	}
}

// @Summary Handle license purchase
// @Description User pay for license, and the server will check the payment, if valid, server will mint license for the user
// @Tags License
// @Accept json
// @Produce json
// @Param  request body MintRequest true "receiver: the buyer; amount: buy how many licenses; value: pay how many wei; txhash: the transaction hash that receiver transfer eth to admin"
// @Success 200 {object} map[string]interface{} "return the transaction hash of mintLicense"
// @Failure 400 {object} map[string]string "request parameter error"
// @Failure 500 {object} map[string]string "internal server error"
// @Router /license/purchase [post]
func HandleLicensePurchase(d *dumper.Dumper) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req MintRequest
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		ethAmount := getEthAmount(float64(dumper.LICENSE_PRICE_USDT), dumper.EthUSD)
		value := ethAmount * float64(req.Amount)
		logger.Debug("license purchase timestamp:", time.Now().Format("2006-01-02 15:04:05"))

		isValid, err := d.PurchaseTxValid(req.TxHash, req.Receiver, value, req.Amount)
		if !isValid {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		history := database.LicensePurchaseHistory{
			TxHash: req.TxHash,
			Payer: req.Receiver,
			Amount: uint16(req.Amount),
			Price: dumper.LICENSE_PRICE_USDT,
			Value: req.Value,
			Done: false,
		}
		err = history.CreateLicensePurchaseHistory()
		if err != nil {
			logger.Debug(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		txHash, err := d.MintNFT(req.Receiver, req.Amount, dumper.LICENSE_PRICE_USDT)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		history.Done = true
		err = history.UpdateLicensePurchaseHistory()
		if err != nil {
			logger.Debug(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "txHash": txHash})
	}
}

func getEthPrice() (float64, error) {
	now := time.Now().Unix()
	if dumper.EthUSD == 0 || dumper.EthUSD_Timestamp+35 < int(now) {
		logger.Debug("Local EthUSD is 0 or timestamp is far behind now, ethusd:", dumper.EthUSD, " ethusd_timestamp:", dumper.EthUSD_Timestamp, " now:", now)
		err := dumper.GetEthPriceFromEtherscan()
		if err != nil {
			logger.Debug(err)
			return 0, err
		}
	}
	return dumper.EthUSD, nil
}

func getEthAmount(usdtAmount float64, ethusdt float64) float64 {
	ethAmount := usdtAmount / ethusdt
	precision := math.Pow(10, 6)
	ethAmountRounded := math.Ceil(ethAmount*precision) / precision
	return ethAmountRounded
}
