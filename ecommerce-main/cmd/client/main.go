package main

import (
	

	"github.com/gin-gonic/gin"
	grpcclient "github.com/kishorens18/ecommerce/cmd/client/grpcClient"
	"github.com/kishorens18/ecommerce/cmd/client/routes"
)

func main() {
	_,conn:=grpcclient.GetGrpcClientInstance();
	defer conn.Close()
	
	r := gin.Default()
	routes.AppRoutes(r)
	r.Run(":8080")

}

// import (
// 	"context"
// 	"fmt"
// 	"log"
// 	"net/http"
// 	"time"

// 	"github.com/dgrijalva/jwt-go"
// 	"github.com/gin-gonic/gin"
// 	"github.com/kishorens18/ecommerce/config"
// 	"github.com/kishorens18/ecommerce/constants"
// 	"github.com/kishorens18/ecommerce/controllers"
// 	"github.com/kishorens18/ecommerce/models"
// 	pb "github.com/kishorens18/ecommerce/proto"
// 	"github.com/kishorens18/ecommerce/services"
// 	"google.golang.org/grpc"

// 	"go.mongodb.org/mongo-driver/bson"
// 	"go.mongodb.org/mongo-driver/mongo"
// )

// var (
// 	mongoclient *mongo.Client
// 	ctx         context.Context
// 	server      *gin.Engine
// )

// type User struct {
// 	Email      string `json:"email"`
// 	Password   string `json:"password"`
// 	CustomerId string `json:"customerid"`
// }

// func main() {
// 	r := gin.Default()
// 	conn, err := grpc.Dial("localhost:5002", grpc.WithInsecure())
// 	if err != nil {
// 		log.Fatalf("Failed to connect: %v", err)
// 	}
// 	defer conn.Close()

// 	client := pb.NewCustomerServiceClient(conn)
// 	r.POST("/signup", func(c *gin.Context) {
// 		var request pb.CustomerDetails

// 		// Parse incoming JSON
// 		if err := c.ShouldBindJSON(&request); err != nil {
// 			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 			return
// 		}

// 		pass, err2 := services.HashPassword(request.Password)
// 		if err2 != nil {
// 			c.JSON(http.StatusBadRequest, gin.H{"error": err2.Error()})
// 			return
// 		}
// 		request.Password = pass

// 		// Call the gRPC service
// 		response, err := client.CreateCustomer(c.Request.Context(), &request)
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 			return
// 		}

// 		c.JSON(http.StatusOK, gin.H{"value": response})
// 	})

// 	r.POST("/signin", func(c *gin.Context) {
// 		var user User
// 		if err := c.ShouldBindJSON(&user); err != nil {
// 			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
// 			return
// 		}
// 		if isValidUser(user) {
// 			token, err := createToken(user.Email, user.CustomerId)
// 			if err != nil {
// 				c.JSON(http.StatusInternalServerError, gin.H{"error": "Token creation failed"})
// 				return
// 			}
// 			client.CreateTokens(c.Request.Context(), &pb.Token{Email: user.Email, Token: token, Customerid: user.CustomerId})

// 			c.JSON(http.StatusOK, gin.H{"token": token})
// 		} else {
// 			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
// 		}
// 	})

// 	r.POST("/deletecustomer", func(c *gin.Context) {
// 		var user models.DeleteRequest
// 		token := c.GetHeader("Authorization")
// 		customerID, err := controllers.ExtractCustomerID(token, constants.SecretKey)
// 		user.CustomerId = customerID
// 		fmt.Println(user.CustomerId)
// 		if err != nil {
// 			fmt.Println("error from token extraction")
// 			c.JSON(http.StatusBadRequest, gin.H{"message": "User not deleted"})
// 			panic(err)
// 		}

// 		if err := c.ShouldBindJSON(&user); err != nil {
// 			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 			fmt.Println("error while blinding")
// 			return
// 		}
// 		_, err2 := client.DeleteCustomer(c.Request.Context(), &pb.DeleteDetails{CustomerID: user.CustomerId})
// 		if err2 != nil {
// 			fmt.Println("error from controller")
// 			c.JSON(http.StatusBadGateway, gin.H{"status": "fail", "message": err.Error()})
// 			c.JSON(http.StatusBadRequest, gin.H{"message": "User not deleted"})
// 		}

// 		c.JSON(http.StatusOK, gin.H{"message": "User deleted"})

// 	})
// 	r.POST("/updatecustomer", func(c *gin.Context) {
// 		var user models.UpdateRequest
// 		if err := c.ShouldBindJSON(&user); err != nil {
// 			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 			return
// 		}
// 		token := c.GetHeader("Authorization")
// 		customerID, err := controllers.ExtractCustomerID(token, constants.SecretKey)
// 		user.CustomerId = customerID

// 		if err != nil {
// 			fmt.Println("err in extracting token")
// 			log.Fatal(err)
// 		}
// 		updatedUser, err := client.UpdateCustomer(c.Request.Context(), &pb.UpdateDetails{CustomerId: user.CustomerId,
// 			Field: user.Field, OldValue: user.OldValue, NewValue: user.NewValue})

// 		if err != nil {
// 			fmt.Println("err from customer response :", err)
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
// 			return
// 		}

// 		c.JSON(http.StatusOK, gin.H{"message": "User updated", "data": updatedUser})

// 	})

// 	r.GET("/getbyid", func(c *gin.Context) {
// 		token := c.GetHeader("Authorization")
// 		result, err := client.GetByCustomerId(c.Request.Context(), &pb.GetbyId{Token: token})
// 		if err != nil {
// 			fmt.Println("Error:", err.Error()) // Print the error message for debugging
// 			c.JSON(http.StatusBadGateway, gin.H{"status": "fail", "message": err.Error()})
// 		}
// 		c.JSON(http.StatusCreated, gin.H{"status": "success", "data": result})
// 	})

// 	r.POST("/resetpassword", func(c *gin.Context) {
// 		var user models.UpdatePassword
// 		if err := c.ShouldBindJSON(&user); err != nil {
// 			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
// 			return
// 		}
// 		response, err := client.UpdatePassword(c.Request.Context(), &pb.PasswordDetails{Email: user.Email, OldPassword: user.OldPassword, NewPassword: user.NewPassword})
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 			return
// 		}

// 		c.JSON(http.StatusOK, gin.H{"value": response})
// 	})

// 	r.Run(":8080")
// }

// func isValidUser(user User) bool {
// 	// Simulated user validation (replace with your actual validation logic)
// 	mongoclient, _ := config.ConnectDataBase()
// 	collection := mongoclient.Database("Ecommerce").Collection("CustomerProfile")

// 	query := bson.M{"customerid": user.CustomerId}
// 	var customer models.Customer
// 	err := collection.FindOne(ctx, query).Decode(&customer)
// 	if err != nil {
// 		return false
// 	}

// 	if customer.Email == user.Email {

// 		hashedPassword := customer.Password
// 		result := services.VerifyPassword(hashedPassword, user.Password)
// 		return result
// 	}
// 	return false
// }

// func createToken(email, customerid string) (string, error) {
// 	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
// 		"email": email, "customerid": customerid,
// 		"exp": time.Now().Add(time.Hour * 1).Unix(),
// 	})

// 	tokenString, err := token.SignedString([]byte(constants.SecretKey))
// 	if err != nil {
// 		return "", err
// 	}

// 	return tokenString, nil
// }
