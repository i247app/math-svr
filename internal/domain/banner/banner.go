package banner

import (
	"math-ai.com/math-ai/internal/domain/shared/mtime"
)

// Banner models ma_banners. Every optional column is a pointer so the
// domain stays faithful to schema nullability. mediaURLKey is the S3
// object key for IMAGE/VIDEO banners; the presigned URL is produced at
// the module/service edge, never stored on the entity. buttonLinkURL is
// an absolute client-supplied URL and is passed through as-is (not
// presigned).
type Banner struct {
	bannerId      int64
	title         *string
	shortText     *string
	mediaType     string
	mediaURLKey   string
	buttonText    *string
	buttonLinkURL *string
	rptFlg        *string
	kwords        *string
	note          *string
	bannerStatus  *string
	status        string
	createId      *int64
	createDt      mtime.MathTime
	modifyId      *int64
	modifyDt      mtime.MathTime
}

func NewBanner() *Banner {
	return &Banner{}
}

func (b *Banner) BannerId() int64              { return b.bannerId }
func (b *Banner) SetBannerId(id int64)         { b.bannerId = id }
func (b *Banner) Title() *string               { return b.title }
func (b *Banner) SetTitle(t *string)           { b.title = t }
func (b *Banner) ShortText() *string           { return b.shortText }
func (b *Banner) SetShortText(t *string)       { b.shortText = t }
func (b *Banner) MediaType() string            { return b.mediaType }
func (b *Banner) SetMediaType(t string)        { b.mediaType = t }
func (b *Banner) MediaURLKey() string          { return b.mediaURLKey }
func (b *Banner) SetMediaURLKey(k string)      { b.mediaURLKey = k }
func (b *Banner) ButtonText() *string          { return b.buttonText }
func (b *Banner) SetButtonText(t *string)      { b.buttonText = t }
func (b *Banner) ButtonLinkURL() *string       { return b.buttonLinkURL }
func (b *Banner) SetButtonLinkURL(u *string)   { b.buttonLinkURL = u }
func (b *Banner) Note() *string                { return b.note }
func (b *Banner) SetNote(n *string)            { b.note = n }
func (b *Banner) RptFlg() *string              { return b.rptFlg }
func (b *Banner) SetRptFlg(v *string)          { b.rptFlg = v }
func (b *Banner) Kwords() *string              { return b.kwords }
func (b *Banner) SetKwords(v *string)          { b.kwords = v }
func (b *Banner) BannerStatus() *string        { return b.bannerStatus }
func (b *Banner) SetBannerStatus(v *string)    { b.bannerStatus = v }
func (b *Banner) Status() string               { return b.status }
func (b *Banner) SetStatus(v string)           { b.status = v }
func (b *Banner) CreateId() *int64             { return b.createId }
func (b *Banner) SetCreateId(id *int64)        { b.createId = id }
func (b *Banner) CreateDt() mtime.MathTime     { return b.createDt }
func (b *Banner) SetCreateDt(t mtime.MathTime) { b.createDt = t }
func (b *Banner) ModifyId() *int64             { return b.modifyId }
func (b *Banner) SetModifyId(id *int64)        { b.modifyId = id }
func (b *Banner) ModifyDt() mtime.MathTime     { return b.modifyDt }
func (b *Banner) SetModifyDt(t mtime.MathTime) { b.modifyDt = t }
