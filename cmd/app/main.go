package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/0x726f6f6b6965/web3-ecommerce/internal/api"
	"github.com/0x726f6f6b6965/web3-ecommerce/internal/api/router"
	"github.com/0x726f6f6b6965/web3-ecommerce/internal/config"
	"github.com/0x726f6f6b6965/web3-ecommerce/internal/helper"
	"github.com/0x726f6f6b6965/web3-ecommerce/internal/storage"
	"github.com/0x726f6f6b6965/web3-ecommerce/pkg/contract"
	"github.com/0x726f6f6b6965/web3-ecommerce/pkg/erc20"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

func main() {
	godotenv.Load()
	path := os.Getenv("CONFIG")
	cfg := new(config.AppConfig)
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatal("read yaml error", err)
		return
	}
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		log.Fatal("unmarshal yaml error", err)
		return
	}
	if cfg.IsDevEnv() {
		storage.NewDevLocalClient(cfg.DB.Table, cfg.DB.Host, cfg.DB.Port)
	} else {
		if err := storage.NewDynamoClient(context.Background(), cfg.DB.Region, cfg.DB.Table); err != nil {
			log.Fatalf(fmt.Sprintf("Failed to create dynamo client: %s", err))
		}
	}

	secret, err := os.ReadFile(os.Getenv(cfg.Secret))
	if err != nil {
		log.Fatal("read jwt secret error", err)
		return
	}
	helper.JwtSecretKey = secret

	owner, err := os.ReadFile(os.Getenv(cfg.Owner))
	if err != nil {
		log.Fatal("read jwt owner error", err)
		return
	}

	prot := cfg.HttpPort
	client, err := ethclient.Dial(cfg.EthUrl)
	if err != nil {
		log.Fatalf(fmt.Sprintf("Failed to connect ethereum: %s", err))
	}
	token, err := contract.CreateContract(cfg.Token.FilePath, cfg.Token.Address)
	if err != nil {
		log.Fatalf(fmt.Sprintf("Failed to create contract: %s", err))
	}
	chainId, err := client.ChainID(context.Background())
	if err != nil {
		log.Fatalf(fmt.Sprintf("Failed to get chain id: %s", err))
	}
	ercService := erc20.NewERC20Service(client, token, chainId, cfg.Token.Decimals)
	api.NewProductApi(time.Minute * 10)
	api.NewOrderApi()
	api.NewPaymentApi(ercService, client)
	api.NewUserApi(client)
	cfg.HttpPort = prot
	if err := startServer(cfg, string(owner)); err != nil {
		log.Fatalf(fmt.Sprintf("Failed to start server: %s", err))
	}
}

func startServer(cfg *config.AppConfig, owner string) error {
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.HttpPort),
		Handler: initEngine(cfg, owner),
	}
	ctx, cancel := context.WithCancel(context.Background())

	go listenToSystemSignals(cancel)

	go func() {
		<-ctx.Done()
		if err := server.Shutdown(context.Background()); err != nil {
			log.Fatalf(fmt.Sprintf("Failed to shutdown server: %s", err))
		}
	}()
	log.Println("Server started success")
	err := server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		log.Println("Server was shutdown gracefully")
		return nil
	}
	return err
}

func initEngine(cfg *config.AppConfig, owner string) *gin.Engine {
	gin.SetMode(func() string {
		if cfg.IsDevEnv() {
			return gin.DebugMode
		}
		return gin.ReleaseMode
	}())
	engine := gin.New()
	engine.Use(cors.Default())
	engine.Use(gin.CustomRecovery(func(c *gin.Context, err interface{}) {
		c.AbortWithStatusJSON(http.StatusOK, gin.H{
			"code": 500,
			"msg":  "Service internal exception!",
		})
	}))
	router.RegisterRoutes(engine, owner)
	return engine
}

func listenToSystemSignals(cancel context.CancelFunc) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	for {
		select {
		case <-signalChan:
			cancel()
			return
		}
	}
}
