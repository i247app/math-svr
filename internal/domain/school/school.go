package school

import (
	mtime "math-ai.com/math-ai/internal/domain/shared/time"
)

// School models ma_schools. Address fields (district, province) are
// nullable and represented as *string so the domain stays faithful to
// the column nullability. image_key is the S3 key; the presigned URL is
// produced at the module/service edge, not stored on the entity.
type School struct {
	id           int64
	schoolId     int64
	name         string
	description  *string
	imageKey     *string
	district     *string
	province     *string
	note         *string
	schoolStatus *string
	status       string
	createId     *int64
	createDt     mtime.MathTime
	modifyId     *int64
	modifyDt     mtime.MathTime
}

func NewSchool() *School {
	return &School{}
}

func (s *School) Id() int64                 { return s.id }
func (s *School) SetId(id int64)            { s.id = id }
func (s *School) SchoolId() int64           { return s.schoolId }
func (s *School) SetSchoolId(id int64)      { s.schoolId = id }
func (s *School) Name() string              { return s.name }
func (s *School) SetName(n string)          { s.name = n }
func (s *School) Description() *string      { return s.description }
func (s *School) SetDescription(d *string)  { s.description = d }
func (s *School) ImageKey() *string         { return s.imageKey }
func (s *School) SetImageKey(k *string)     { s.imageKey = k }
func (s *School) District() *string         { return s.district }
func (s *School) SetDistrict(d *string)     { s.district = d }
func (s *School) Province() *string         { return s.province }
func (s *School) SetProvince(p *string)     { s.province = p }
func (s *School) Note() *string             { return s.note }
func (s *School) SetNote(n *string)         { s.note = n }
func (s *School) SchoolStatus() *string     { return s.schoolStatus }
func (s *School) SetSchoolStatus(v *string) { s.schoolStatus = v }
func (s *School) Status() string            { return s.status }
func (s *School) SetStatus(v string)        { s.status = v }
func (s *School) CreateId() *int64          { return s.createId }
func (s *School) SetCreateId(id *int64)     { s.createId = id }
func (s *School) CreateDt() mtime.MathTime  { return s.createDt }
func (s *School) SetCreateDt(t mtime.MathTime) {
	s.createDt = t
}
func (s *School) ModifyId() *int64         { return s.modifyId }
func (s *School) SetModifyId(id *int64)    { s.modifyId = id }
func (s *School) ModifyDt() mtime.MathTime { return s.modifyDt }
func (s *School) SetModifyDt(t mtime.MathTime) {
	s.modifyDt = t
}
