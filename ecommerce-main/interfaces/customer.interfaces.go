package interfaces

import (
	"github.com/kishorens18/ecommerce/models"
	ecommerce "github.com/kishorens18/ecommerce/proto"
)

type ICustomer interface {
	CreateCustomer(customer *models.Customer) (*models.CustomerDBResponse, error)
	CreateTokens(token *models.Token) (*ecommerce.Empty, error)
	UpdatePassword(Password *models.UpdatePassword) (*models.CustomerDBResponse, error)
	UpdateCustomer(cus *models.UpdateRequest) (*models.CustomerDBResponse, error)
	DeleteCustomer(cus *models.DeleteRequest) error
	GetByCustomerId(res string) (*models.Customer, error)
}
