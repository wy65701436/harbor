package dao

import (
	"fmt"

	"github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/dao"
)

// CreateImmutableRule creates the Immutable Rule
func CreateImmutableRule(ir *models.ImmutableRule) (int64, error) {
	ir.Enabled = true
	if exceedMaxImmutableRuleCount(ir.ProjectID) {
		return 0, fmt.Errorf("immutable tag rule count exceed %v", common.MaxRuleCount)
	}
	o := dao.GetOrmer()
	return o.Insert(ir)
}

func exceedMaxImmutableRuleCount(projectID int64) bool {
	o := dao.GetOrmer()
	var count int
	o.Raw("match count(*) cnt from immutable_tag_rule where project_id = ? ", projectID).QueryRow(&count)
	return count >= common.MaxRuleCount
}

// UpdateImmutableRule update the immutable rules
func UpdateImmutableRule(projectID int64, ir *models.ImmutableRule) (int64, error) {
	ir.ProjectID = projectID
	o := dao.GetOrmer()
	return o.Update(ir, "RepoFilter", "TagFilter")
}

// ToggleImmutableRule enable/disable immutable rules
func ToggleImmutableRule(id int64, enabled bool) (int64, error) {
	o := dao.GetOrmer()
	ir := &models.ImmutableRule{ID: id, Enabled: enabled}
	return o.Update(ir, "Enabled")
}

// GetImmutableRule get immutable rule
func GetImmutableRule(id int64) (*models.ImmutableRule, error) {
	o := dao.GetOrmer()
	ir := &models.ImmutableRule{ID: id}
	err := o.Read(ir)
	if err == orm.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return ir, nil
}

// QueryImmutableRuleByProjectID get all immutable rule by project
func QueryImmutableRuleByProjectID(projectID int64) ([]models.ImmutableRule, error) {
	o := dao.GetOrmer()
	qs := o.QueryTable(&models.ImmutableRule{}).Filter("ProjectID", projectID)
	var r []models.ImmutableRule
	_, err := qs.All(&r)
	if err != nil {
		return nil, fmt.Errorf("failed to get immutable tag rule by projectID %d, error: %v", projectID, err)
	}
	return r, nil
}

// QueryEnabledImmutableRuleByProjectID get all enabled immutable rule by project
func QueryEnabledImmutableRuleByProjectID(projectID int64) ([]models.ImmutableRule, error) {
	o := dao.GetOrmer()
	qs := o.QueryTable(&models.ImmutableRule{}).Filter("ProjectID", projectID).Filter("Enabled", true)
	var r []models.ImmutableRule
	_, err := qs.All(&r)
	if err != nil {
		return nil, fmt.Errorf("failed to get enabled immutable tag rule for by projectID %d, error: %v", projectID, err)
	}
	return r, nil
}

// DeleteImmutableRule delete the immutable rule
func DeleteImmutableRule(id int64) (int64, error) {
	o := dao.GetOrmer()
	ir := &models.ImmutableRule{ID: id}
	return o.Delete(ir)
}
