package school

import (
	"math-ai.com/math-ai/internal/domain/shared/audit"
)

// School models ma_schools. Address fields (district, province) are
// nullable and represented as *string so the domain stays faithful to
// the column nullability. image_key is the S3 key; the presigned URL is
// produced at the module/service edge, not stored on the entity.
type School struct {
	audit.Audit

	schoolId     int64
	name         string
	description  *string
	imageKey     *string
	district     *string
	province     *string
	schoolStatus *string
}

func NewSchool() *School {
	return &School{}
}

func (s *School) SchoolId() int64      { return s.schoolId }
func (s *School) SetSchoolId(id int64) { s.schoolId = id }

func (s *School) Name() string     { return s.name }
func (s *School) SetName(n string) { s.name = n }

func (s *School) Description() *string     { return s.description }
func (s *School) SetDescription(d *string) { s.description = d }

func (s *School) ImageKey() *string     { return s.imageKey }
func (s *School) SetImageKey(k *string) { s.imageKey = k }

func (s *School) District() *string     { return s.district }
func (s *School) SetDistrict(d *string) { s.district = d }

func (s *School) Province() *string     { return s.province }
func (s *School) SetProvince(p *string) { s.province = p }

func (s *School) SchoolStatus() *string     { return s.schoolStatus }
func (s *School) SetSchoolStatus(v *string) { s.schoolStatus = v }
