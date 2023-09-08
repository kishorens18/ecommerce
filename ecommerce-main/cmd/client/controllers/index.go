package controllers_test

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	grpcclient "github.com/kishorens18/ecommerce/cmd/client/grpcClient"
	"github.com/kishorens18/ecommerce/config"
	"github.com/kishorens18/ecommerce/constants"
	"github.com/kishorens18/ecommerce/controllers"
	"github.com/kishorens18/ecommerce/models"
	pb "github.com/kishorens18/ecommerce/proto"
	"github.com/kishorens18/ecommerce/services"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	mongoclient *mongo.Client
	ctx         context.Context
	server      *gin.Engine
)

type User struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	CustomerId string `json:"customerid"`
}

func HandlerSignUp(c *gin.Context) {
	var request pb.CustomerDetails

	// Parse incoming JSON
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	grpcClient, _ := grpcclient.GetGrpcClientInstance()
	pass, err2 := services.HashPassword(request.Password)
	if err2 != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err2.Error()})
		return
	}
	request.Password = pass
	response, err := grpcClient.CreateCustomer(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"value": response})
}

func HandlerSignIn(c *gin.Context) {
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	if isValidUser(user) {
		token, err := createToken(user.Email, user.CustomerId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Token creation failed"})
			return
		}
		grpcClient, _ := grpcclient.GetGrpcClientInstance()
		grpcClient.CreateTokens(c.Request.Context(), &pb.Token{Email: user.Email, Token: token, Customerid: user.CustomerId})

		c.JSON(http.StatusOK, gin.H{"token": token})
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
	}

}

func HandlerDeleteCustomer(c *gin.Context) {
	var user models.DeleteRequest
	token := c.GetHeader("Authorization")
	customerID, err := controllers.ExtractCustomerID(token, constants.SecretKey)
	user.CustomerId = customerID
	fmt.Println(user.CustomerId)
	if err != nil {
		fmt.Println("error from token extraction")
		c.JSON(http.StatusBadRequest, gin.H{"message": "User not deleted"})
		panic(err)
	}

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		fmt.Println("error while blinding")
		return
	}

	grpcClient, _ := grpcclient.GetGrpcClientInstance()
	_, err2 := grpcClient.DeleteCustomer(c.Request.Context(), &pb.DeleteDetails{CustomerID: user.CustomerId})
	if err2 != nil {
		fmt.Println("error from controller")
		c.JSON(http.StatusBadGateway, gin.H{"status": "fail", "message": err.Error()})
		c.JSON(http.StatusBadRequest, gin.H{"message": "User not deleted"})
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted"})

}

func HandlerUpdateCustomer(c *gin.Context) {
	var user models.UpdateRequest
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	token := c.GetHeader("Authorization")
	customerID, err := controllers.ExtractCustomerID(token, constants.SecretKey)
	user.CustomerId = customerID

	if err != nil {
		fmt.Println("err in extracting token")
		log.Fatal(err)
	}

	grpcClient, _ := grpcclient.GetGrpcClientInstance()
	updatedUser, err := grpcClient.UpdateCustomer(c.Request.Context(), &pb.UpdateDetails{CustomerId: user.CustomerId,
		Field: user.Field, OldValue: user.OldValue, NewValue: user.NewValue})

	if err != nil {
		fmt.Println("err from customer response :", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User updated", "data": updatedUser})

}

func HandlerGetById(c *gin.Context) {

	token := c.GetHeader("Authorization")

	grpcClient, _ := grpcclient.GetGrpcClientInstance()
	result, err := grpcClient.GetByCustomerId(c.Request.Context(), &pb.GetbyId{Token: token})
	if err != nil {
		fmt.Println("Error:", err.Error()) // Print the error message for debugging
		c.JSON(http.StatusBadGateway, gin.H{"status": "fail", "message": err.Error()})
	}
	c.JSON(http.StatusCreated, gin.H{"status": "success", "data": result})

}

func HandlerResetPassword(c *gin.Context) {

	var user models.UpdatePassword
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	grpcClient, _ := grpcclient.GetGrpcClientInstance()
	response, err := grpcClient.UpdatePassword(c.Request.Context(), &pb.PasswordDetails{Email: user.Email, OldPassword: user.OldPassword, NewPassword: user.NewPassword})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"value": response})

}

func isValidUser(user User) bool {
	// Simulated user validation (replace with your actual validation logic)
	mongoclient, _ := config.ConnectDataBase()
	collection := mongoclient.Database("Ecommerce").Collection("CustomerProfile")

	query := bson.M{"customerid": user.CustomerId}
	var customer models.Customer
	err := collection.FindOne(ctx, query).Decode(&customer)
	if err != nil {
		return false
	}

	if customer.Email == user.Email {

		hashedPassword := customer.Password
		result := services.VerifyPassword(hashedPassword, user.Password)
		return result
	}
	return false
}

func createToken(email, customerid string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": email, "customerid": customerid,
		"exp": time.Now().Add(time.Hour * 1).Unix(),
	})

	tokenString, err := token.SignedString([]byte(constants.SecretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
