package policy

import (
	"github.com/goharbor/harbor/src/common/models"
)

// Manager manages the notification policies
type Manager interface {
	// Create new rule
	Create(*models.NotificationPolicy) (int64, error)
	// List the policies, returns the rule list and error
	List(int64) ([]*models.NotificationPolicy, error)
	// Get rule with specified ID
	Get(int64) (*models.NotificationPolicy, error)
	// GetByNameAndProjectID get rule by the name and projectID
	GetByNameAndProjectID(string, int64) (*models.NotificationPolicy, error)
	// Update the specified rule
	Update(*models.NotificationPolicy) error
	// Delete the specified rule
	Delete(int64) error
	// Test the specified rule
	Test(*models.NotificationPolicy) error
	// GetRelatedPolices get event type related policies in project
	GetRelatedPolices(int64, string) ([]*models.NotificationPolicy, error)
}
