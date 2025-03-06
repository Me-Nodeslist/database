package server

import (
	"log"
	"net/http"

	"github.com/Me-Nodeslist/database/docs"
	"github.com/Me-Nodeslist/database/logs"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Router struct {
	*gin.Engine
}

var logger = logs.Logger("server")

func NewServer(endpoint string) (*http.Server, error) {
	log.Println("Begin listen and server...")
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// allow all cross-origin requests
	router.Use(cors.Default())

	router.MaxMultipartMemory = 8 << 20 // 8 MiB
	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Welcome Node-Delegation Server!")
	})

	docs.SwaggerInfo.BasePath = "/v1"
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r := Router{
		router,
	}
	r.registerLicenseRouter()
	r.registerNodeRouter()
	r.registerRewardRouter()

	return &http.Server{
		Addr:    endpoint,
		Handler: router,
	}, nil
}

func (r Router) registerLicenseRouter() {
	r.GET("/license/amount", GetLicenseAmount()) // all and delegated
	r.GET("/license/amount/owner/:address", GetLicenseAmountOfOwner())
	r.GET("/license/info/owner/:address", GetLicenseInfosOfOwner()) // page
	r.GET("/license/price", GetLicensePrice())
}

func (r Router) registerNodeRouter() {
	r.GET("/node/amount", GetNodeAmount())
	r.GET("/node/info", GetNodeInfos()) // page
	r.GET("/node/info/owner/:address", GetNodeInfoOfOwner())
	r.GET("/node/info/recipient/:address", GetNodeInfosOfRecipient())            // page
	r.GET("/node/info/delegation/:address", GetNodeInfosOfdelegation()) // page
}

func (r Router) registerRewardRouter() {
	r.GET("/reward/info/:address", GetRewardInfo())
	r.GET("/reward/redeem/info/:address", GetRedeemInfo())
}
