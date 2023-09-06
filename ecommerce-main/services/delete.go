package services

import (
	"github.com/kishorens18/ecommerce/models"
	"go.mongodb.org/mongo-driver/bson"
)

func (p *CustomerService) DeleteCustomer(user *models.DeleteRequest) {
	// Check if the customer ID is provided
	if user.CustomerId == "" {
		return
	}

	// Create a filter to find the customer by ID
	filter := bson.M{"customerid": user.CustomerId}

	// Delete the customer document
	result, err := p.ProfileCollection.DeleteOne(p.ctx, filter)
	if err != nil {
		return
	}

	// Check if a document was deleted
	if result.DeletedCount == 0 {
		return
	}
}
