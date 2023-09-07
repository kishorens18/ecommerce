package services

import (
	"context"
	"fmt"
	"testing"
)

var ctx context.Context

func TestVerifyPassword(t *testing.T) {

	// hashedPassword,err:= HashPassword("345")
	// fmt.Println(hashedPassword)
	// if err!=nil{
	//  fmt.Println(err)
	// }
	// result := VerifyPassword(hashedPassword, "345")
	// fmt.Println(result)

	// mongoclient, _ := config.ConnectDataBase()
	// collection := mongoclient.Database("Ecommerce").Collection("CustomerProfile")
	// query := bson.M{"customerid": "1456"}
	// var customer models.Customer
	// err2 := collection.FindOne(ctx, query).Decode(&customer)
	// if err2 != nil {
	// 	fmt.Println(err2)
	// }
	hashPassword, _ := HashPassword("1456")
	fmt.Println(hashPassword)
	result := VerifyPassword(hashPassword, "1456")
	fmt.Println(result)
}
