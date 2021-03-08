package policy

import (
	"context"
	"fmt"
	commonhttp "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/notification/policy/dao"
	"github.com/goharbor/harbor/src/pkg/notifier/model"
	"net/http"
	"time"
)

// Manager manages the notification policies
type Manager interface {
	// Create new policy
	Create(ctx context.Context, policy *models.NotificationPolicy) (int64, error)
	// List the policies, returns the policy list and error
	List(ctx context.Context, query *q.Query) ([]*models.NotificationPolicy, error)
	// Get policy with specified ID
	Get(ctx context.Context, id int64) (*models.NotificationPolicy, error)
	// GetByNameAndProjectID get policy by the name and projectID
	GetByNameAndProjectID(ctx context.Context, name string, projectID int64) (*models.NotificationPolicy, error)
	// Update the specified policy
	Update(ctx context.Context, policy *models.NotificationPolicy) error
	// Delete the specified policy
	Delete(ctx context.Context, policyID int64) error
	// Test the specified policy
	Test(ctx context.Context, policy *models.NotificationPolicy) error
	// GetRelatedPolices get event type related policies in project
	GetRelatedPolices(ctx context.Context, projectID int64, eventType string) ([]*models.NotificationPolicy, error)
}

// DefaultManager ...
type DefaultManager struct {
	dao dao.DAO
}

// NewDefaultManger ...
func NewDefaultManger() *DefaultManager {
	return &DefaultManager{
		dao: dao.New(),
	}
}

// Create notification policy
func (m *DefaultManager) Create(ctx context.Context, policy *models.NotificationPolicy) (int64, error) {
	t := time.Now()
	policy.CreationTime = t
	policy.UpdateTime = t

	err := policy.ConvertToDBModel()
	if err != nil {
		return 0, err
	}
	return m.dao.Create(ctx, policy)
}

// List the notification policies, returns the policy list and error
func (m *DefaultManager) List(ctx context.Context, query *q.Query) ([]*models.NotificationPolicy, error) {
	policies := []*models.NotificationPolicy{}
	persisPolicies, err := m.dao.List(ctx, query)
	if err != nil {
		return nil, err
	}

	for _, policy := range persisPolicies {
		err := policy.ConvertFromDBModel()
		if err != nil {
			return nil, err
		}
		policies = append(policies, policy)
	}

	return policies, nil
}

// Get notification policy with specified ID
func (m *DefaultManager) Get(ctx context.Context, id int64) (*models.NotificationPolicy, error) {
	policy, err := m.dao.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if policy == nil {
		return nil, nil
	}
	if err := policy.ConvertFromDBModel(); err != nil {
		return nil, err
	}
	return policy, err
}

// GetByNameAndProjectID notification policy by the name and projectID
func (m *DefaultManager) GetByNameAndProjectID(ctx context.Context, name string, projectID int64) (*models.NotificationPolicy, error) {
	query := q.New(q.KeyWords{"name": name, "project_id": projectID})
	policies, err := m.dao.List(ctx, query)
	if err != nil {
		return nil, err
	}
	if len(policies) == 0 {
		return nil, errors.New(nil).WithCode(errors.NotFoundCode).WithMessage("no notification policy found")
	}
	policy := policies[0]
	if err := policy.ConvertFromDBModel(); err != nil {
		return nil, err
	}
	return policy, err
}

// Update the specified notification policy
func (m *DefaultManager) Update(ctx context.Context, policy *models.NotificationPolicy) error {
	policy.UpdateTime = time.Now()
	err := policy.ConvertToDBModel()
	if err != nil {
		return err
	}
	return m.dao.Update(ctx, policy)
}

// Delete the specified notification policy
func (m *DefaultManager) Delete(ctx context.Context, policyID int64) error {
	return m.dao.Delete(ctx, policyID)
}

// Test the specified notification policy, just test for network connection without request body
func (m *DefaultManager) Test(policy *models.NotificationPolicy) error {
	for _, target := range policy.Targets {
		switch target.Type {
		case model.NotifyTypeHTTP, model.NotifyTypeSlack:
			return m.policyHTTPTest(target.Address, target.SkipCertVerify)
		default:
			return fmt.Errorf("invalid policy target type: %s", target.Type)
		}
	}
	return nil
}

func (m *DefaultManager) policyHTTPTest(address string, skipCertVerify bool) error {
	req, err := http.NewRequest(http.MethodPost, address, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	client := http.Client{
		Transport: commonhttp.GetHTTPTransportByInsecure(skipCertVerify),
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	log.Debugf("policy test success with address %s, skip cert verify :%v", address, skipCertVerify)

	return nil
}

// GetRelatedPolices get policies including event type in project
func (m *DefaultManager) GetRelatedPolices(ctx context.Context, projectID int64, eventType string) ([]*models.NotificationPolicy, error) {
	policies, err := m.dao.List(ctx, q.New(q.KeyWords{"project_id": projectID}))
	if err != nil {
		return nil, fmt.Errorf("failed to get notification policies with projectID %d: %v", projectID, err)
	}

	var result []*models.NotificationPolicy

	for _, ply := range policies {
		if !ply.Enabled {
			continue
		}
		for _, t := range ply.EventTypes {
			if t != eventType {
				continue
			}
			result = append(result, ply)
		}
	}
	return result, nil
}
