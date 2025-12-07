# Migration Plan: Move to Translation Table Pattern

## Phase 1: Create Translation Tables

### 1. Create grade_translations table
```sql
CREATE TABLE grade_translations (
    id CHAR(36) NOT NULL,
    grade_id CHAR(36) NOT NULL,
    language VARCHAR(10) NOT NULL,
    label VARCHAR(128) NOT NULL,
    description VARCHAR(255) NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY unique_grade_language (grade_id, language),
    FOREIGN KEY (grade_id) REFERENCES grades(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
```

### 2. Create semester_translations table
```sql
CREATE TABLE semester_translations (
    id CHAR(36) NOT NULL,
    semester_id CHAR(36) NOT NULL,
    language VARCHAR(10) NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY unique_semester_language (semester_id, language),
    FOREIGN KEY (semester_id) REFERENCES semesters(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
```

### 3. Create chapter_translations table
```sql
CREATE TABLE chapter_translations (
    id CHAR(36) NOT NULL,
    chapter_id CHAR(36) NOT NULL,
    language VARCHAR(10) NOT NULL,
    title VARCHAR(200) NOT NULL,
    description TEXT DEFAULT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY unique_chapter_language (chapter_id, language),
    FOREIGN KEY (chapter_id) REFERENCES chapters(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
```

### 4. Create lesson_translations table
```sql
CREATE TABLE lesson_translations (
    id CHAR(36) NOT NULL,
    lesson_id CHAR(36) NOT NULL,
    language VARCHAR(10) NOT NULL,
    title VARCHAR(200) NOT NULL,
    content TEXT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY unique_lesson_language (lesson_id, language),
    FOREIGN KEY (lesson_id) REFERENCES lessons(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
```

## Phase 2: Migrate Existing Data (if you have current chapters/lessons)

### Migrate chapters
```sql
-- Insert into chapter_translations from existing chapters
INSERT INTO chapter_translations (id, chapter_id, language, title, description)
SELECT UUID(), id, languague, title, description
FROM chapters;
```

### Migrate lessons
```sql
-- Insert into lesson_translations from existing lessons
INSERT INTO lesson_translations (id, lesson_id, language, title, content)
SELECT UUID(), id, languague, title, content
FROM lessons;
```

### Migrate semesters
```sql
-- Insert into semester_translations from existing semesters
INSERT INTO semester_translations (id, semester_id, language, name, description)
SELECT UUID(), id, languague, name, description
FROM semesters;
```

## Phase 3: Update Schema

### Remove language columns from main tables
```sql
-- Remove from chapters (after migration)
ALTER TABLE chapters DROP COLUMN languague;
ALTER TABLE chapters DROP COLUMN title;
ALTER TABLE chapters DROP COLUMN description;

-- Remove from lessons
ALTER TABLE lessons DROP COLUMN languague;
ALTER TABLE lessons DROP COLUMN title;
ALTER TABLE lessons DROP COLUMN content;

-- Remove from semesters
ALTER TABLE semesters DROP COLUMN languague;
ALTER TABLE semesters DROP COLUMN name;
ALTER TABLE semesters DROP COLUMN description;
```

## Phase 4: Application Layer Changes

### Go Service Layer Example
```go
type GradeResponse struct {
    ID           string    `json:"id"`
    IconURL      *string   `json:"icon_url"`
    Status       string    `json:"status"`
    DisplayOrder int       `json:"display_order"`
    Label        string    `json:"label"`        // from translation
    Description  string    `json:"description"`  // from translation
    Language     string    `json:"language"`
}

// Repository method
func (r *GradeRepository) GetByIDWithTranslation(ctx context.Context, id string, language string) (*domain.Grade, error) {
    query := `
        SELECT g.id, g.icon_url, g.status, g.display_order,
               COALESCE(gt_user.label, gt_default.label) as label,
               COALESCE(gt_user.description, gt_default.description) as description,
               COALESCE(gt_user.language, gt_default.language) as language
        FROM grades g
        LEFT JOIN grade_translations gt_user ON g.id = gt_user.grade_id AND gt_user.language = ?
        LEFT JOIN grade_translations gt_default ON g.id = gt_default.grade_id AND gt_default.language = 'EN'
        WHERE g.id = ? AND g.deleted_dt IS NULL
    `

    // Execute query...
}

// Service method
func (s *GradeService) GetGrade(ctx context.Context, id string) (*dto.GradeResponse, error) {
    // Get language from context (from middleware or header)
    language := s.getLanguageFromContext(ctx) // e.g., "VN", "EN", "FR"

    grade, err := s.repo.GetByIDWithTranslation(ctx, id, language)
    if err != nil {
        return nil, err
    }

    return s.responseBuilder.BuildGradeResponse(ctx, grade), nil
}
```

## Benefits Summary

✅ **Single Source of Truth**: One row per grade, regardless of languages
✅ **Easy to Add Languages**: Just insert new translation rows
✅ **No Duplicate Metadata**: Icons, status, display_order stored once
✅ **Clean Foreign Keys**: chapters.grade_id always points to the same grade
✅ **Fallback Support**: Can default to English if translation missing
✅ **Better Performance**: Smaller main tables with indexed translations
✅ **Maintainability**: Update metadata once, applies to all languages

## Rollback Strategy

If needed, you can revert by:
1. Re-adding language columns to main tables
2. Migrating data back from translation tables
3. Dropping translation tables
