package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"math-ai.com/math-ai/internal/domain/banner"
	"math-ai.com/math-ai/internal/domain/shared/mtime"
	"math-ai.com/math-ai/internal/infrastructure/database"
	"math-ai.com/math-ai/internal/infrastructure/persistence/mysql/models"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/pagination"
)

const (
	bannerTable = "ma_banners"

	bannerColumns = `b.id, b.banner_id, b.title, b.short_text, b.media_type,
		b.media_url_key, b.button_text, b.button_link_url, b.note,
		b.banner_status, b.status,
		b.create_id, b.create_dt, b.modify_id, b.modify_dt`

	// Reads exclude both system-inactive rows and the DELETED business
	// status used by the soft-delete path. INACTIVE banner_status rows
	// still pass here — callers filter them out via BannerStatus when they
	// only want display-ready banners.
	bannerActiveWhere = `b.status = ? AND b.deleted_dt IS NULL
		AND (b.banner_status IS NULL OR b.banner_status != ?)`
)

func bannerActiveArgs() []any {
	return []any{enum.StatusActive, enum.BannerStatusTypeDeleted}
}

type BannerRepository struct {
	db database.Executor
}

func NewBannerRepository(db database.Executor) banner.IRepository {
	return &BannerRepository{db: db}
}

func scanBanner(s database.RowScanner) (*models.BannerModel, error) {
	var m models.BannerModel
	if err := s.Scan(&m.Id, &m.BannerId, &m.Title, &m.ShortText, &m.MediaType,
		&m.MediaURLKey, &m.ButtonText, &m.ButtonLinkURL, &m.Note,
		&m.BannerStatus, &m.Status,
		&m.CreateId, &m.CreateDt, &m.ModifyId, &m.ModifyDt); err != nil {
		return nil, err
	}
	return &m, nil
}

// findOneBy is the single-row read helper. `where` is a package-controlled
// SQL fragment; args supply placeholders. bannerActiveWhere is prepended so
// every read excludes soft-deleted and inactive rows.
func (r *BannerRepository) findOneBy(ctx context.Context, where string, args ...any) (*banner.Banner, error) {
	fullArgs := append(bannerActiveArgs(), args...)
	query := `SELECT ` + bannerColumns + ` FROM ` + bannerTable + ` b WHERE ` +
		bannerActiveWhere + ` AND (` + where + `)`

	m, err := scanBanner(r.db.QueryRow(ctx, query, fullArgs...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("banner repo find (%s): %w", where, err)
	}
	return ModelToDomainBanner(m), nil
}

func (r *BannerRepository) findBareById(ctx context.Context, id int64) (*banner.Banner, error) {
	args := append(bannerActiveArgs(), id)
	query := `SELECT ` + bannerColumns + ` FROM ` + bannerTable + ` b WHERE ` +
		bannerActiveWhere + ` AND b.id = ?`

	m, err := scanBanner(r.db.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("banner repo find bare by id: %w", err)
	}
	return ModelToDomainBanner(m), nil
}

func (r *BannerRepository) FindByBannerId(ctx context.Context, bannerId int64) (*banner.Banner, error) {
	return r.findOneBy(ctx, "b.banner_id = ?", bannerId)
}

func (r *BannerRepository) ListBanners(ctx context.Context, params *banner.ListBannersParams) ([]*banner.Banner, *pagination.Pagination, error) {
	filterWhere, filterArgs := buildBannerListFilterClause(params)

	countArgs := append(bannerActiveArgs(), filterArgs...)
	countQuery := `SELECT COUNT(*) FROM ` + bannerTable + ` b WHERE ` +
		bannerActiveWhere + filterWhere

	var total int64
	if err := r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, nil, fmt.Errorf("banner repo count: %w", err)
	}

	listArgs := append(bannerActiveArgs(), filterArgs...)
	// Newest banners first — banners are time-sensitive display content.
	query := `SELECT ` + bannerColumns + ` FROM ` + bannerTable + ` b WHERE ` +
		bannerActiveWhere + filterWhere +
		` ORDER BY b.create_dt DESC, b.id DESC`

	var pg *pagination.Pagination
	if params == nil || !params.TakeAll {
		page := int64(1)
		limit := int64(20)
		if params != nil {
			page = params.Page
			limit = params.Limit
		}
		pg = pagination.NewPagination(page, limit, total)
		query += ` LIMIT ? OFFSET ?`
		listArgs = append(listArgs, pg.Size, pg.Skip)
	} else {
		pg = pagination.NewPagination(1, total, total)
	}

	rows, err := r.db.Query(ctx, query, listArgs...)
	if err != nil {
		return nil, nil, fmt.Errorf("banner repo list: %w", err)
	}
	defer rows.Close()

	var banners []*banner.Banner
	for rows.Next() {
		m, err := scanBanner(rows)
		if err != nil {
			return nil, nil, fmt.Errorf("banner repo scan row: %w", err)
		}
		banners = append(banners, ModelToDomainBanner(m))
	}
	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("banner repo rows iteration: %w", err)
	}
	return banners, pg, nil
}

// buildBannerListFilterClause appends optional narrowing predicates. Search
// is a case-insensitive substring match against title and short_text;
// MediaType and BannerStatus narrow by exact match. Keeping each clause
// guarded preserves placeholder/arg ordering parity.
func buildBannerListFilterClause(params *banner.ListBannersParams) (string, []any) {
	if params == nil {
		return "", nil
	}
	var (
		clause string
		args   []any
	)
	if params.Search != nil {
		needle := strings.TrimSpace(*params.Search)
		if needle != "" {
			clause += ` AND (b.title LIKE ? OR b.short_text LIKE ?)`
			like := "%" + needle + "%"
			args = append(args, like, like)
		}
	}
	if params.MediaType != nil && strings.TrimSpace(*params.MediaType) != "" {
		clause += ` AND b.media_type = ?`
		args = append(args, strings.TrimSpace(*params.MediaType))
	}
	if params.BannerStatus != nil && strings.TrimSpace(*params.BannerStatus) != "" {
		clause += ` AND b.banner_status = ?`
		args = append(args, strings.TrimSpace(*params.BannerStatus))
	}
	if len(params.BannerIds) > 0 {
		placeholders := make([]string, len(params.BannerIds))
		for i, id := range params.BannerIds {
			placeholders[i] = "?"
			args = append(args, id)
		}
		clause += ` AND b.banner_id IN (` + strings.Join(placeholders, ",") + `)`
	}
	return clause, args
}

func (r *BannerRepository) Create(ctx context.Context, b *banner.Banner) (*banner.Banner, error) {
	query := `
		INSERT INTO ` + bannerTable + `
			(banner_id, title, short_text, media_type, media_url_key,
			 button_text, button_link_url, note, banner_status,
			 create_id, create_dt, modify_dt)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.Exec(ctx, query,
		b.BannerId(), b.Title(), b.ShortText(), b.MediaType(), b.MediaURLKey(),
		b.ButtonText(), b.ButtonLinkURL(), b.Note(), b.BannerStatus(),
		b.CreateId(), mtime.Now().Time, mtime.Now().Time)
	if err != nil {
		return nil, fmt.Errorf("banner repo create: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("banner repo last insert id: %w", err)
	}
	return r.findBareById(ctx, id)
}

// Update applies a partial update. Nullable columns use COALESCE(?, col)
// so a nil pointer leaves them unchanged. The two NOT NULL string columns
// (media_type, media_url_key) use a non-empty guard: an empty value means
// "leave alone", so callers pass the current value only when changing it.
func (r *BannerRepository) Update(ctx context.Context, b *banner.Banner) error {
	var mediaTypeArg any
	if b.MediaType() != "" {
		mediaTypeArg = b.MediaType()
	}
	var mediaURLKeyArg any
	if b.MediaURLKey() != "" {
		mediaURLKeyArg = b.MediaURLKey()
	}

	query := `
		UPDATE ` + bannerTable + `
		SET title           = COALESCE(?, title),
			short_text      = COALESCE(?, short_text),
			media_type      = COALESCE(?, media_type),
			media_url_key   = COALESCE(?, media_url_key),
			button_text     = COALESCE(?, button_text),
			button_link_url = COALESCE(?, button_link_url),
			note            = COALESCE(?, note),
			banner_status   = COALESCE(?, banner_status),
			modify_id       = COALESCE(?, modify_id),
			modify_dt       = ?
		WHERE banner_id = ?
	`
	if _, err := r.db.Exec(ctx, query,
		b.Title(), b.ShortText(), mediaTypeArg, mediaURLKeyArg,
		b.ButtonText(), b.ButtonLinkURL(), b.Note(), b.BannerStatus(),
		b.ModifyId(), mtime.Now().Time, b.BannerId()); err != nil {
		return fmt.Errorf("banner repo update: %w", err)
	}
	return nil
}

func (r *BannerRepository) SoftDeleteByBannerId(ctx context.Context, bannerId int64) error {
	query := `
		UPDATE ` + bannerTable + `
		SET banner_status = ?,
			status        = ?,
			deleted_dt    = ?,
			modify_dt     = ?
		WHERE banner_id = ?
	`
	now := mtime.Now().Time
	if _, err := r.db.Exec(ctx, query,
		enum.BannerStatusTypeDeleted, enum.StatusInactive, now, now, bannerId); err != nil {
		return fmt.Errorf("banner repo soft delete: %w", err)
	}
	return nil
}

func (r *BannerRepository) ForceDeleteByBannerId(ctx context.Context, bannerId int64) error {
	query := `
		DELETE FROM ` + bannerTable + `
		WHERE banner_id = ?
	`
	if _, err := r.db.Exec(ctx, query, bannerId); err != nil {
		return fmt.Errorf("banner repo force delete: %w", err)
	}
	return nil
}

func ModelToDomainBanner(m *models.BannerModel) *banner.Banner {
	b := banner.NewBanner()
	b.SetId(m.Id)
	b.SetBannerId(m.BannerId)
	b.SetTitle(m.Title)
	b.SetShortText(m.ShortText)
	b.SetMediaType(m.MediaType)
	b.SetMediaURLKey(m.MediaURLKey)
	b.SetButtonText(m.ButtonText)
	b.SetButtonLinkURL(m.ButtonLinkURL)
	b.SetNote(m.Note)
	b.SetBannerStatus(m.BannerStatus)
	b.SetStatus(m.Status)
	b.SetCreateId(m.CreateId)
	b.SetCreateDt(mtime.MathTime{Time: m.CreateDt})
	b.SetModifyId(m.ModifyId)
	b.SetModifyDt(mtime.MathTime{Time: m.ModifyDt})
	return b
}
