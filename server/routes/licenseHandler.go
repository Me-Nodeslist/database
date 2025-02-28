package routes

import (
	"github.com/Me-Nodeslist/database/database"
	"github.com/Me-Nodeslist/database/logs"
	"github.com/gin-gonic/gin"
)

type licenseAmount struct {
	amount int64
	delegatedAmount int64
}

var logger = logs.Logger("server")

func GetLicenseAmount() gin.HandlerFunc {
	return func(c *gin.Context) {
		res := licenseAmount{}
		amount, err := database.GetLicenseAmount()
		if err != nil {
			logger.Error(err.Error())
		}
	}
}