package routes

import (
	"github.com/gin-gonic/gin"
	controllers_test "github.com/kishorens18/ecommerce/cmd/client/controllers"
)

func AppRoutes(r *gin.Engine) {
	r.POST("/signup", controllers_test.HandlerSignUp)
	r.POST("/signin", controllers_test.HandlerSignIn)
	r.POST("/deletecustomer", controllers_test.HandlerDeleteCustomer)
	r.POST("/updatecustomer", controllers_test.HandlerUpdateCustomer)
	r.GET("/getbyid", controllers_test.HandlerGetById)
	r.POST("/resetpassword", controllers_test.HandlerResetPassword)
}
