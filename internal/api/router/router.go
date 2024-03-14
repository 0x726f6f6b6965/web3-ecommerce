package router

import (
	"github.com/0x726f6f6b6965/web3-ecommerce/internal/api"
	"github.com/0x726f6f6b6965/web3-ecommerce/internal/api/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(server *gin.Engine, admin string) {
	RegisterAuthRouter(server.Group("/auth/"))
	RegisterUserRouter(server.Group("/user/"))
	RegisterProductRouter(server.Group("/product/"))
	RegisterOrderRouter(server.Group("/order/"))
	RegisterAdminRouter(server.Group("/admin/"), admin)
	RegisterPaymentRouter(server.Group("/payment/"))
}
func RegisterAuthRouter(group *gin.RouterGroup) {
	group.POST("/register", api.UserApi.Register)
	group.POST("/token", api.UserApi.GetToken)
}
func RegisterUserRouter(group *gin.RouterGroup) {
	group.Use(middleware.UserAuthorization())
	group.GET("/info", api.UserApi.GetUser)
	group.PATCH("/info", api.UserApi.UpdateUser)
}

func RegisterProductRouter(group *gin.RouterGroup) {
	group.GET("/list", api.ProductApi.GetProductList)
	group.GET(":productId", api.ProductApi.GetProduct)
}
func RegisterOrderRouter(group *gin.RouterGroup) {
	group.Use(middleware.UserAuthorization())
	group.POST("/create", api.OrderApi.CreateOrder)
	group.GET("/list", api.OrderApi.GetOrders)
	group.GET("/:orderId", api.OrderApi.GetOrder)
	group.GET("/cancel/:orderId", api.OrderApi.CancelOrder)
}

func RegisterPaymentRouter(group *gin.RouterGroup) {
	group.Use(middleware.UserAuthorization())
	group.POST("/pay", api.PaymentApi.Pay)
}

func RegisterAdminRouter(group *gin.RouterGroup, admin string) {
	group.Use(middleware.AdminAuthorization(admin))
	group.POST("/product/create", api.ProductApi.CreateProduct)
	group.PATCH("/product/:productId", api.ProductApi.UpdateProduct)
	group.PATCH("/order/:orderId", api.OrderApi.UpdateOrderStatus)
}
