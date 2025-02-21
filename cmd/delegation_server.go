package cmd

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
	"github.com/urfave/cli/v2"

	"github.com/Me-Nodeslist/database/database"
	"github.com/Me-Nodeslist/database/dumper"
)

var DelegationServerCmd = &cli.Command{
	Name:  "server",
	Usage: "node-delegation server",
	Subcommands: []*cli.Command{
		serverRunCmd,
	},
}

var serverRunCmd = &cli.Command{
	Name:  "run",
	Usage: "run node-delegation server",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "endpoint",
			Aliases: []string{"e"},
			Usage:   "input your endpoint",
			Value:   ":8082",
		},
		&cli.StringFlag{
			Name:  "chain",
			Usage: "input chain name, e.g.(dev)",
			Value: "product",
		},
		&cli.StringFlag{
			Name:  "licenseNFT",
			Usage: "input licenseNFT contract address",
			Value: "",
		},
		&cli.StringFlag{
			Name:  "delMEMO",
			Usage: "input delMEMO contract address",
			Value: "",
		},
		&cli.StringFlag{
			Name:  "settlement",
			Usage: "input settlement contract address",
			Value: "",
		},
		&cli.StringFlag{
			Name:  "delegation",
			Usage: "input delegation contract address",
			Value: "",
		},
	},
	Action: func(ctx *cli.Context) error {
		endPoint := ctx.String("endpoint")
		chain := ctx.String("chain")

		licenseNFT := ctx.String("licenseNFT")
		delMEMO := ctx.String("delMEMO")
		settlement := ctx.String("settlement")
		delegation := ctx.String("delegation")

		addrs := &dumper.ContractAddress{
			LicenseNFT: common.HexToAddress(licenseNFT),
			DelMEMO:    common.HexToAddress(delMEMO),
			Settlement: common.HexToAddress(settlement),
			Delegation: common.HexToAddress(delegation),
		}

		cctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		err := database.InitDatabase("~/.nodedelegation-"+chain)
		if err != nil {
			return err
		}

		dumper, err := dumper.NewDumper(chain, addrs)
		if err != nil {
			return err
		}

		err = dumper.Dump()
		if err != nil {
			return err
		}
		go dumper.SubscribeEvents(cctx)

		srv, err := NewLightServer(endPoint)
		if err != nil {
			log.Fatalf("new store node server: %s\n", err)
		}

		go func() {
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("listen: %s\n", err)
			}
		}()

		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		log.Println("Shutting down server...")

		if err := srv.Shutdown(cctx); err != nil {
			log.Fatal("Server forced to shutdown: ", err)
		}

		log.Println("Server exiting")

		return nil
	},
}

func NewLightServer(endpoint string) (*http.Server, error) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.MaxMultipartMemory = 8 << 20 // 8 MiB
	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Welcome Node-Delegation Server!")
	})

	return &http.Server{
		Addr:    endpoint,
		Handler: router,
	}, nil
}
