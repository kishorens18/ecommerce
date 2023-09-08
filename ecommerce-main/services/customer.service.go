package services

import (
	"context"
	"fmt"
	"log"

	"github.com/kishorens18/ecommerce/config"
	"github.com/kishorens18/ecommerce/interfaces"
	"github.com/kishorens18/ecommerce/models"
	ecommerce "github.com/kishorens18/ecommerce/proto"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CustomerService struct {
	ProfileCollection *mongo.Collection
	tokenCollection   *mongo.Collection
	ctx               context.Context
}

// HashPassword hashes a given password using bcrypt.
func HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

// VerifyPassword compares a hashed password with a plain password.
func VerifyPassword(hashedPassword, plainPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	return err == nil
}

func VerifyPasswordForResetPassword(hashedPassword, plainPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	return err == nil
}

// InitCustomerService initializes a new CustomerService instance.
func InitCustomerService(collection, tokenCollection *mongo.Collection, ctx context.Context) interfaces.ICustomer {
	return &CustomerService{collection, tokenCollection, ctx}
}

// CreateCustomer creates a new customer and stores it in the database.
func (p *CustomerService) CreateCustomer(user *models.Customer) (*models.CustomerDBResponse, error) {

	res, err := p.ProfileCollection.InsertOne(p.ctx, &user)
	if err != nil {
		return nil, err
	}
	mongoclient, _ := config.ConnectDataBase()
	collection := mongoclient.Database("Ecommerce").Collection("CustomerProfile")
	query := bson.M{"customerid": user.CustomerId}
	var customer models.Customer
	err2 := collection.FindOne(p.ctx, query).Decode(&customer)
	if err != nil {
		return nil, err2
	}

	var newUser models.CustomerDBResponse
	query2 := bson.M{"_id": res.InsertedID}
	err = p.ProfileCollection.FindOne(p.ctx, query2).Decode(&newUser)
	if err != nil {
		return nil, err
	}
	return &newUser, nil
}

// UpdatePassword updates the password for a customer.
func (p *CustomerService) UpdatePassword(user *models.UpdatePassword) (*models.CustomerDBResponse, error) {
	if user.OldPassword == user.NewPassword {
		return nil, nil
	}

	query := bson.M{"email": user.Email}
	var customer models.Customer
	err := p.ProfileCollection.FindOne(p.ctx, query).Decode(&customer)
	if err != nil {
		return nil, err
	}
	if !VerifyPasswordForResetPassword(customer.Password, user.OldPassword) {
		fmt.Println("errror in verifying")
		return nil, nil
	}

	user.NewPassword, _ = HashPassword(user.NewPassword)
	filter := bson.M{"email": user.Email}
	update := bson.M{"$set": bson.M{"password": user.NewPassword}}

	_, err = p.ProfileCollection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Fatal(err)
	}

	response := models.CustomerDBResponse{
		Customer_id: customer.CustomerId,
	}
	return &response, nil
}

// CustomerLogin performs customer login and returns the customer ID.
func (p *CustomerService) CustomerLogin(email, password string) (*models.CustomerDBResponse, error) {
	query := bson.M{"email": email, "password": password}
	var customer models.Customer
	err := p.ProfileCollection.FindOne(p.ctx, query).Decode(&customer)
	if err != nil {
		return nil, err
	}

	responseCustomer := &models.CustomerDBResponse{
		Customer_id: customer.CustomerId,
	}
	return responseCustomer, nil
}

// CreateTokens creates tokens for a user.
func (p *CustomerService) CreateTokens(user *models.Token) (*ecommerce.Empty, error) {
	res, err := p.tokenCollection.InsertOne(p.ctx, &user)
	if err != nil {
		return nil, err
	}

	var newUser models.Token
	query := bson.M{"_id": res.InsertedID}
	err = p.ProfileCollection.FindOne(p.ctx, query).Decode(&newUser)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (p *CustomerService) UpdateCustomer(user *models.UpdateRequest) (*models.CustomerDBResponse, error) {
	var updatedUser models.CustomerDBResponse
	if user.Field == "country" || user.Field == "street1" || user.Field == "street2" || user.Field == "city" || user.Field == "state" || user.Field == "zip" {

		filter := bson.D{
			{Key: "customerid", Value: user.CustomerId},
			{Key: "address." + user.Field, Value: user.OldValue},
		}
		update := bson.D{
			{Key: "$set", Value: bson.D{
				{Key: "address.$." + user.Field, Value: user.NewValue},
			}},
		}
		options := options.Update()

		result, err := p.ProfileCollection.UpdateOne(p.ctx, filter, update, options)
		if err != nil {
			fmt.Println("error while updating")
			return nil, err
		}

		if result.MatchedCount == 0 {
			// No documents matched the filter criteria, so return an error
			return nil, mongo.ErrNoDocuments
		}

		filter3 := bson.D{{Key: "customerid", Value: user.CustomerId}}
		// Fetch the updated user document

		err2 := p.ProfileCollection.FindOne(p.ctx, filter3).Decode(&updatedUser)
		if err2 != nil {
			fmt.Println("Error decoding document:", err2)
			return nil, err2
		}
		return &updatedUser, nil

	} else {
		filter2 := bson.D{
			{Key: "customerid", Value: user.CustomerId},
			{Key: user.Field, Value: user.OldValue},
		}
		update2 := bson.D{
			{Key: "$set", Value: bson.D{
				{Key: user.Field, Value: user.NewValue},
			}},
		}
		options2 := options.Update()

		result, err := p.ProfileCollection.UpdateOne(p.ctx, filter2, update2, options2)
		if err != nil {
			fmt.Println("error while updating")
			return nil, err
		}

		if result.MatchedCount == 0 {
			// No documents matched the filter criteria, so return an error
			return nil, mongo.ErrNoDocuments
		}

		filter := bson.D{{Key: "customerid", Value: user.CustomerId}}
		// Fetch the updated user document

		err2 := p.ProfileCollection.FindOne(p.ctx, filter).Decode(&updatedUser)
		if err2 != nil {
			fmt.Println("Error decoding document:", err2)
			return nil, err2
		}
		return &updatedUser, nil
	}
	return &updatedUser, nil
}
func (p *CustomerService) DeleteCustomer(user *models.DeleteRequest) error {
	// Check if the customer ID is provided
	if user.CustomerId == "" {
		fmt.Println("request from controller is empty")
		return status.Error(codes.InvalidArgument, "Invalid Customer ID")
	}

	// Create a filter to find the customer by ID
	filter := bson.M{"customerid": user.CustomerId}

	// Delete the customer document
	result, err := p.ProfileCollection.DeleteOne(p.ctx, filter)
	if err != nil {
		fmt.Println("error while deleting in db")
		return status.Error(codes.InvalidArgument, "Invalid Customer ID")
	}

	// Check if a document was deleted
	if result.DeletedCount == 0 {
		return status.Error(codes.InvalidArgument, "Invalid Customer ID")
	}
	return nil
}

func (p *CustomerService) GetByCustomerId(id string) (*models.Customer, error) {
	filter := bson.D{{Key: "customerid", Value: id}}
	var customer *models.Customer
	res := p.ProfileCollection.FindOne(p.ctx, filter)
	err := res.Decode(&customer)
	if err != nil {
		return nil, err
	}
	return customer, nil

}
