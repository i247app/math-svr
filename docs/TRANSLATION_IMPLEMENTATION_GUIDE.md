# Translation Table Implementation Guide

## Overview

This guide explains how to use the translation table pattern implemented for multi-language support in the Math-AI application.

## Architecture

### Database Schema

```
Main Tables (language-agnostic):
├── grades
├── semesters
├── chapters
└── lessons

Translation Tables (language-specific):
├── grade_translations
├── semester_translations
├── chapter_translations
└── lesson_translations
```

### Key Concepts

1. **One Entity = One Row**: Each grade, semester, chapter, or lesson exists once in the main table
2. **Multiple Translations**: Each entity can have multiple translation rows (one per language)
3. **Language Context**: Language is determined from HTTP headers and passed through context
4. **Fallback Support**: If a translation doesn't exist, fall back to default language (EN)

## Step-by-Step Usage

### 1. Language Middleware

Add language detection middleware to extract language from request headers:

```go
// internal/handlers/http/middleware/language.go
package middleware

import (
	"net/http"
	"math-ai.com/math-ai/internal/shared/utils/language"
)

func LanguageMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get language from Accept-Language header or custom header
		lang := r.Header.Get("Accept-Language")
		if lang == "" {
			lang = r.Header.Get("X-Language")
		}

		// Normalize language code
		if lang == "" {
			lang = "EN"
		} else {
			// Extract primary language (e.g., "en-US" -> "EN")
			if len(lang) >= 2 {
				lang = strings.ToUpper(lang[:2])
			}
		}

		// Set language in context
		ctx := language.SetLanguage(r.Context(), lang)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
```

### 2. Repository Pattern

Repository methods should join with translation tables using language from context:

```go
// Example: Get grade with translation
func (r *GradeRepository) GetByIDWithTranslation(ctx context.Context, tx *sql.Tx, id string) (*domain.Grade, error) {
	lang := language.GetLanguage(ctx)

	query := `
		SELECT
			g.id, g.icon_url, g.status, g.display_order,
			g.create_id, g.create_dt, g.modify_id, g.modify_dt, g.deleted_dt,
			COALESCE(gt_user.label, gt_default.label) as label,
			COALESCE(gt_user.description, gt_default.description) as description
		FROM grades g
		LEFT JOIN grade_translations gt_user
			ON g.id = gt_user.grade_id AND gt_user.language = ?
		LEFT JOIN grade_translations gt_default
			ON g.id = gt_default.grade_id AND gt_default.language = 'EN'
		WHERE g.id = ? AND g.deleted_dt IS NULL
	`

	var model models.GradeModel
	var label, description string

	err := r.db.QueryRowContext(ctx, query, lang, id).Scan(
		&model.ID, &model.IconURL, &model.Status, &model.DisplayOrder,
		&model.CreateID, &model.CreateDT, &model.ModifyID, &model.ModifyDT, &model.DeletedDT,
		&label, &description,
	)

	if err != nil {
		return nil, err
	}

	grade := domain.BuildGradeDomainFromModel(&model)
	grade.SetLabel(label)
	grade.SetDescription(description)

	return grade, nil
}

// List all grades with translations
func (r *GradeRepository) ListWithTranslations(ctx context.Context, tx *sql.Tx) ([]*domain.Grade, error) {
	lang := language.GetLanguage(ctx)

	query := `
		SELECT
			g.id, g.icon_url, g.status, g.display_order,
			g.create_id, g.create_dt, g.modify_id, g.modify_dt, g.deleted_dt,
			COALESCE(gt_user.label, gt_default.label) as label,
			COALESCE(gt_user.description, gt_default.description) as description
		FROM grades g
		LEFT JOIN grade_translations gt_user
			ON g.id = gt_user.grade_id AND gt_user.language = ?
		LEFT JOIN grade_translations gt_default
			ON g.id = gt_default.grade_id AND gt_default.language = 'EN'
		WHERE g.deleted_dt IS NULL
		ORDER BY g.display_order
	`

	rows, err := r.db.QueryContext(ctx, query, lang)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var grades []*domain.Grade
	for rows.Next() {
		var model models.GradeModel
		var label, description string

		err := rows.Scan(
			&model.ID, &model.IconURL, &model.Status, &model.DisplayOrder,
			&model.CreateID, &model.CreateDT, &model.ModifyID, &model.ModifyDT, &model.DeletedDT,
			&label, &description,
		)
		if err != nil {
			return nil, err
		}

		grade := domain.BuildGradeDomainFromModel(&model)
		grade.SetLabel(label)
		grade.SetDescription(description)

		grades = append(grades, grade)
	}

	return grades, nil
}
```

### 3. Creating/Updating with Translations

When creating or updating entities, you need to handle both main table and translation table:

```go
// Create grade with translation
func (r *GradeRepository) CreateWithTranslation(ctx context.Context, tx *sql.Tx, grade *domain.Grade, language string) error {
	// Insert into main table
	query := `
		INSERT INTO grades (id, icon_url, status, display_order, create_id, create_dt, modify_id, modify_dt)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := tx.ExecContext(ctx, query,
		grade.ID(), grade.IconURL(), grade.Status(), grade.DisplayOrder(),
		grade.CreateID(), grade.CreatedAt(), grade.ModifyID(), grade.ModifiedAt(),
	)
	if err != nil {
		return err
	}

	// Insert into translation table
	translationQuery := `
		INSERT INTO grade_translations (id, grade_id, language, label, description)
		VALUES (UUID(), ?, ?, ?, ?)
	`
	_, err = tx.ExecContext(ctx, translationQuery,
		grade.ID(), language, grade.Label(), grade.Description(),
	)

	return err
}

// Update translation for existing grade
func (r *GradeRepository) UpdateTranslation(ctx context.Context, tx *sql.Tx, gradeID, language, label, description string) error {
	query := `
		INSERT INTO grade_translations (id, grade_id, language, label, description)
		VALUES (UUID(), ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			label = VALUES(label),
			description = VALUES(description)
	`
	_, err := tx.ExecContext(ctx, query, gradeID, language, label, description)
	return err
}
```

### 4. Service Layer

Service layer uses repositories transparently:

```go
func (s *GradeService) GetGrade(ctx context.Context, id string) (*dto.GradeResponse, error) {
	// Language is already in context from middleware
	grade, err := s.repo.GetByIDWithTranslation(ctx, nil, id)
	if err != nil {
		return nil, err
	}

	return s.responseBuilder.BuildGradeResponse(ctx, grade), nil
}

func (s *GradeService) ListGrades(ctx context.Context) ([]*dto.GradeResponse, error) {
	grades, err := s.repo.ListWithTranslations(ctx, nil)
	if err != nil {
		return nil, err
	}

	var responses []*dto.GradeResponse
	for _, grade := range grades {
		responses = append(responses, s.responseBuilder.BuildGradeResponse(ctx, grade))
	}

	return responses, nil
}

func (s *GradeService) CreateGrade(ctx context.Context, req *dto.CreateGradeRequest) (*dto.GradeResponse, error) {
	lang := language.GetLanguage(ctx)

	grade := domain.NewGradeDomain()
	grade.GenerateID()
	grade.SetLabel(req.Label)
	grade.SetDescription(req.Description)
	grade.SetStatus(req.Status)
	grade.SetDisplayOrder(req.DisplayOrder)

	err := s.repo.DoTransaction(ctx, func(tx *sql.Tx) error {
		return s.repo.CreateWithTranslation(ctx, tx, grade, lang)
	})

	if err != nil {
		return nil, err
	}

	return s.responseBuilder.BuildGradeResponse(ctx, grade), nil
}
```

### 5. Controller Layer

Controllers don't need to change - they work transparently:

```go
func (c *GradeController) GetGrade(w http.ResponseWriter, r *http.Request) {
	// Language is already in context from middleware
	id := chi.URLParam(r, "id")

	grade, err := c.gradeService.GetGrade(r.Context(), id)
	if err != nil {
		// Handle error
		return
	}

	response.JSON(w, http.StatusOK, grade)
}
```

## Adding a New Language

To add a new language (e.g., French):

1. **No schema changes needed!** Just insert new translation rows:

```sql
-- Add French translations for all grades
INSERT INTO grade_translations (id, grade_id, language, label, description)
SELECT UUID(), id, 'FR',
       'Classe 1',  -- French label
       'Première année de l\'enseignement primaire'  -- French description
FROM grades
WHERE id = 'd46c8252-06a7-4d6e-8f24-3525278214ae';
```

2. **That's it!** The application will automatically use French translations when `Accept-Language: fr` header is sent.

## Migration Steps

### Before Running Migrations

**IMPORTANT**: The migrations will:
1. Create translation tables
2. Migrate existing data to translation tables
3. Remove language columns from main tables

Make sure you have a database backup before running!

### Run Migrations

```bash
# Run all three migrations in order
make migrate-up
```

### Verify Migration

```sql
-- Check that data was migrated
SELECT COUNT(*) FROM grade_translations;  -- Should have rows
SELECT COUNT(*) FROM semester_translations;  -- Should have rows
SELECT COUNT(*) FROM chapter_translations;  -- Should have rows
SELECT COUNT(*) FROM lesson_translations;  -- Should have rows

-- Check that columns were removed
DESCRIBE grades;  -- Should NOT have label, description columns
DESCRIBE semesters;  -- Should NOT have name, description, languague columns
```

### Rollback (if needed)

```bash
# Rollback migrations in reverse order
make migrate-down
```

## Testing

### Test with Different Languages

```bash
# Get grade in English
curl -H "Accept-Language: en" http://localhost:8080/grades/123

# Get grade in Vietnamese
curl -H "Accept-Language: vn" http://localhost:8080/grades/123

# Get grade in French (if translations exist)
curl -H "Accept-Language: fr" http://localhost:8080/grades/123

# Fallback to default if language not found
curl -H "Accept-Language: es" http://localhost:8080/grades/123  # Returns EN
```

## Best Practices

1. **Always use context**: Pass language through context, never as explicit parameters
2. **Default to EN**: Always have English translations as fallback
3. **Use COALESCE**: Join both user language and default language for fallback support
4. **Transaction for creation**: When creating entities, insert both main table and translation in same transaction
5. **Indexes**: Translation tables have indexes on (entity_id, language) for performance

## Common Queries

### Get all available languages for an entity

```sql
SELECT DISTINCT language
FROM grade_translations
WHERE grade_id = '123';
```

### Find entities missing translations

```sql
SELECT g.id, g.status
FROM grades g
LEFT JOIN grade_translations gt ON g.id = gt.grade_id AND gt.language = 'VN'
WHERE gt.id IS NULL;
```

### Bulk add missing translations

```sql
INSERT INTO grade_translations (id, grade_id, language, label, description)
SELECT UUID(), g.id, 'FR',
       CONCAT('Grade ', g.display_order),  -- Generate default label
       'French description needed'         -- Placeholder
FROM grades g
LEFT JOIN grade_translations gt ON g.id = gt.grade_id AND gt.language = 'FR'
WHERE gt.id IS NULL;
```

## Troubleshooting

### Issue: Getting NULL for label/description
- **Cause**: No translation exists for requested language AND no EN fallback
- **Solution**: Ensure all entities have at least EN translation

### Issue: Wrong language returned
- **Cause**: Language not properly set in context
- **Solution**: Check middleware is applied and language header is sent

### Issue: Query performance slow
- **Cause**: Missing indexes on translation tables
- **Solution**: Ensure indexes exist on (entity_id, language) columns

## Performance Considerations

- Translation tables use indexes on (entity_id, language) - fast lookups
- LEFT JOIN with COALESCE is efficient for fallback logic
- Consider caching frequently accessed translations
- For bulk operations, use batch queries with IN clauses

## Future Enhancements

- Translation management API endpoints
- Translation versioning/history
- Automatic translation suggestions
- Translation completeness dashboard
