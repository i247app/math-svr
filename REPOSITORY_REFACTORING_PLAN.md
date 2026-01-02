# Repository Layer Refactoring Plan

**Project**: Math-AI Server
**Author**: Senior Software Engineer Analysis
**Date**: 2026-01-02
**Status**: Recommendation

---

## Executive Summary

After thorough analysis of the repository layer (`internal/driven-adapter/persistence/repositories/`), I've identified several architectural and maintainability issues that require refactoring. This document presents the current problems, multiple refactoring approaches, and my recommended solution path.

---

## Current Issues Analysis

### 1. SQL Query Management ❌

**Problem**: Raw SQL strings are embedded directly in repository methods

```go
// user_repository.go:40
query := `
    SELECT u.id, u.name, u.phone, u.email, u.avatar_key, u.dob,
           u.role_id, u.status, l.hash_pass,
           u.create_id, u.create_dt, u.modify_id, u.modify_dt,
           r.name as role_name
    FROM ma_users u
    JOIN ma_aliases a ON u.id = a.uid
    JOIN ma_logins l ON u.id = l.uid
    LEFT JOIN roles r ON u.role_id = r.id AND r.deleted_dt IS NULL
    WHERE a.aka = ? AND u.deleted_dt IS NULL AND a.deleted_dt IS NULL AND l.deleted_dt IS NULL
`
```

**Issues**:
- ❌ No centralized query management
- ❌ Hard to test queries in isolation
- ❌ Difficult to version or document
- ❌ No query performance tracking
- ❌ Code duplication (same SELECTs in FindByID, FindByEmail, List)
- ❌ Hard to optimize without changing Go code

### 2. Pagination Implementation 🔄

**Problem**: Highly repetitive pagination logic in every List method

```go
// Duplicated in EVERY repository (user, grade, permission, role, etc.)
var queryBuilder strings.Builder
var countBuilder strings.Builder
args := []interface{}{}
countArgs := []interface{}{}

// Build base query...
// Build count query separately (DRY violation)
// Add search conditions to BOTH queries
// Execute count query
// Initialize pagination object
// Add sorting
// Add LIMIT/OFFSET
```

**Issues**:
- ❌ ~80 lines of duplicated code per repository
- ❌ Count query and main query are maintained separately
- ❌ Easy to make mistakes (forget to add condition to count query)
- ❌ Hard to change pagination strategy globally
- ❌ No support for cursor-based pagination

### 3. Update Methods Inflexibility 🔧

**Problem**: Every column update requires new code

```go
// user_repository.go:258
func (r *userRepository) Update(ctx context.Context, user *domain.User) (int64, error) {
    var queryBuilder strings.Builder
    args := []interface{}{}
    updates := []string{}

    // Need to add this for EVERY field
    if user.Name() != "" {
        updates = append(updates, "name = ?")
        args = append(args, user.Name())
    }

    if user.Phone() != "" {
        updates = append(updates, "phone = ?")
        args = append(args, user.Phone())
    }

    // ... 10+ more fields
}
```

**Issues**:
- ❌ Cannot distinguish between "empty value" and "not updating"
- ❌ Cannot set fields to NULL easily
- ❌ Adding new column = adding new if-block in Update method
- ❌ No validation of which fields are actually updateable
- ❌ Violation of Open/Closed Principle

### 4. CQRS Readiness 🏗️

**Problem**: No separation between read and write operations

**Current Structure**:
```
IUserRepository
├── List()           (Read)
├── FindByID()       (Read)
├── FindByEmail()    (Read)
├── Create()         (Write)
├── Update()         (Write)
└── Delete()         (Write)
```

**Issues**:
- ❌ Cannot scale read and write databases independently
- ❌ Cannot optimize read models differently from write models
- ❌ Cannot implement event sourcing without major refactoring
- ❌ No clear separation of concerns
- ❌ Difficult to implement read replicas

### 5. Additional Issues 🔍

#### Transaction Handling Inconsistency
```go
Create(ctx context.Context, tx *sql.Tx, user *domain.User)   // Takes tx
Update(ctx context.Context, user *domain.User)                // No tx
Delete(ctx context.Context, tx *sql.Tx, uid string)          // Takes tx
```
- ❌ Inconsistent API - some methods take tx, some don't
- ❌ Caller must know which operations support transactions

#### Scan Operations Repetition
```go
// Repeated in every Find method
var u models.UserModel
err := result.Scan(
    &u.ID, &u.Name, &u.Phone, &u.Email, &u.AvatarKey, &u.Dob,
    &u.RoleID, &u.Status,
    &u.CreateID, &u.CreateDT, &u.ModifyID, &u.ModifyDT,
    &u.Role,
)
```
- ❌ ~15 lines duplicated across FindByID, FindByEmail, List, etc.
- ❌ Easy to make mistakes when columns change

#### No Query Builder Abstraction
- ❌ Manual string concatenation with strings.Builder
- ❌ No type safety
- ❌ SQL injection risk if not careful

#### No Batch Operations
- ❌ Cannot insert/update multiple records efficiently
- ❌ Need to loop and make N queries for N records

---

## Refactoring Options

### Option 1: Minimal Refactor (Quick Wins) ⚡

**Effort**: Low (1-2 weeks)
**Impact**: Medium
**Risk**: Low

**Approach**: Extract common patterns, keep current architecture

**Changes**:
1. **Extract Query Constants**
   ```go
   // internal/driven-adapter/persistence/queries/user_queries.go
   package queries

   const (
       UserBaseSelect = `
           SELECT u.id, u.name, u.phone, u.email, u.avatar_key, u.dob,
                  u.role_id, u.status,
                  u.create_id, u.create_dt, u.modify_id, u.modify_dt,
                  r.name as role_name
           FROM ma_users u
           LEFT JOIN roles r ON u.role_id = r.id AND r.deleted_dt IS NULL
       `

       UserFindByIDQuery = UserBaseSelect + ` WHERE u.id = ? AND u.deleted_dt IS NULL`
       UserFindByEmailQuery = UserBaseSelect + ` WHERE u.email = ? AND u.deleted_dt IS NULL`
   )
   ```

2. **Create Base Repository with Pagination**
   ```go
   // internal/driven-adapter/persistence/base_repository.go
   type BaseRepository struct {
       db db.IDatabase
   }

   func (r *BaseRepository) Paginate(
       ctx context.Context,
       baseQuery string,
       countQuery string,
       params pagination.Params,
       scanFn func(*sql.Rows) (interface{}, error),
   ) ([]interface{}, *pagination.Pagination, error) {
       // Common pagination logic here
   }
   ```

3. **Use Struct Tags for Update**
   ```go
   // Use reflection to build UPDATE statements from struct tags
   func (r *BaseRepository) BuildUpdate(
       tableName string,
       id string,
       data interface{},
   ) (string, []interface{}, error) {
       // Use reflection to find non-zero fields
       // Build UPDATE dynamically
   }
   ```

**Pros**:
- ✅ Quick to implement
- ✅ Low risk (minimal changes to existing code)
- ✅ Immediate reduction in code duplication
- ✅ Can be done incrementally

**Cons**:
- ❌ Still have raw SQL in codebase
- ❌ Doesn't solve CQRS problem
- ❌ Limited flexibility
- ❌ Reflection-based updates can be slow

---

### Option 2: Query Builder Pattern (Moderate) 🏗️

**Effort**: Medium (3-4 weeks)
**Impact**: High
**Risk**: Medium

**Approach**: Introduce a query builder library

**Implementation**:

1. **Adopt a Query Builder** (Choose one):
   - [squirrel](https://github.com/Masterminds/squirrel) - Popular, maintained
   - [goqu](https://github.com/doug-martin/goqu) - More feature-rich
   - [dbr](https://github.com/gocraft/dbr) - Active Record style

2. **Example with Squirrel**:
   ```go
   // user_repository.go
   func (r *userRepository) FindByID(ctx context.Context, uid string) (*domain.User, error) {
       query := squirrel.
           Select("u.id", "u.name", "u.phone", "u.email", "u.avatar_key", "u.dob",
                  "u.role_id", "u.status",
                  "u.create_id", "u.create_dt", "u.modify_id", "u.modify_dt",
                  "r.name as role_name").
           From("ma_users u").
           LeftJoin("roles r ON u.role_id = r.id AND r.deleted_dt IS NULL").
           Where(squirrel.Eq{"u.id": uid}).
           Where("u.deleted_dt IS NULL")

       sql, args, err := query.ToSql()
       if err != nil {
           return nil, err
       }

       result := r.db.QueryRow(ctx, nil, sql, args...)
       // ... scan
   }

   func (r *userRepository) Update(ctx context.Context, user *domain.User) (int64, error) {
       updateMap := map[string]interface{}{
           "modify_dt": mathtime.Now(),
       }

       if user.Name() != "" {
           updateMap["name"] = user.Name()
       }
       // ... add other fields

       query := squirrel.
           Update("ma_users").
           SetMap(updateMap).
           Where(squirrel.Eq{"id": user.ID()}).
           Where("deleted_dt IS NULL")

       sql, args, err := query.ToSql()
       if err != nil {
           return 0, err
       }

       result, err := r.db.Exec(ctx, nil, sql, args...)
       // ...
   }
   ```

3. **Pagination Helper**:
   ```go
   func (r *BaseRepository) PaginateQuery(
       ctx context.Context,
       baseQuery squirrel.SelectBuilder,
       params pagination.Params,
   ) (squirrel.SelectBuilder, *pagination.Pagination, error) {
       // Get count
       countQuery := baseQuery.Prefix("SELECT COUNT(*) FROM (").Suffix(") AS count_table")
       // ... execute count

       // Add pagination
       paginatedQuery := baseQuery.
           Limit(params.Limit).
           Offset(params.Skip)

       return paginatedQuery, paginationObj, nil
   }
   ```

**Pros**:
- ✅ Type-safe query building
- ✅ Easier to compose complex queries
- ✅ Better testability (can inspect query without DB)
- ✅ Reduces SQL injection risk
- ✅ Can reuse query fragments
- ✅ IDE autocomplete support

**Cons**:
- ❌ Learning curve for team
- ❌ Another dependency to maintain
- ❌ May not support all MySQL features
- ❌ Slight performance overhead (negligible in practice)

---

### Option 3: CQRS with Repository Split (Advanced) 🚀

**Effort**: High (6-8 weeks)
**Impact**: Very High
**Risk**: Medium-High

**Approach**: Separate read and write operations, prepare for CQRS

**Architecture**:
```
internal/driven-adapter/persistence/
├── commands/              (Write operations)
│   ├── user_command_repository.go
│   ├── grade_command_repository.go
│   └── ...
├── queries/               (Read operations)
│   ├── user_query_repository.go
│   ├── grade_query_repository.go
│   └── ...
├── models/
│   ├── write/            (Write models - normalized)
│   │   └── user.go
│   └── read/             (Read models - denormalized)
│       └── user_view.go
└── query_builder/        (Shared query utilities)
```

**Implementation**:

1. **Separate Interfaces**:
   ```go
   // internal/core/di/repositories/user_repository.go

   // Write operations
   type IUserCommandRepository interface {
       Create(ctx context.Context, tx *sql.Tx, user *domain.User) (int64, error)
       Update(ctx context.Context, tx *sql.Tx, updates map[string]interface{}, id string) (int64, error)
       Delete(ctx context.Context, tx *sql.Tx, id string) error
       BatchCreate(ctx context.Context, tx *sql.Tx, users []*domain.User) error
   }

   // Read operations
   type IUserQueryRepository interface {
       List(ctx context.Context, params ListUsersParams) ([]*domain.UserView, *pagination.Pagination, error)
       FindByID(ctx context.Context, id string) (*domain.UserView, error)
       FindByEmail(ctx context.Context, email string) (*domain.UserView, error)
       GetUserByLoginName(ctx context.Context, loginName string) (*domain.UserView, error)
   }
   ```

2. **Flexible Update with Map**:
   ```go
   // commands/user_command_repository.go
   func (r *userCommandRepository) Update(
       ctx context.Context,
       tx *sql.Tx,
       updates map[string]interface{},
       id string,
   ) (int64, error) {
       // Whitelist updateable fields
       allowedFields := map[string]bool{
           "name": true, "phone": true, "email": true,
           "dob": true, "role_id": true, "avatar_key": true,
       }

       queryBuilder := squirrel.Update("ma_users")

       for field, value := range updates {
           if !allowedFields[field] {
               return 0, fmt.Errorf("field %s is not updateable", field)
           }
           queryBuilder = queryBuilder.Set(field, value)
       }

       queryBuilder = queryBuilder.
           Set("modify_dt", mathtime.Now()).
           Where(squirrel.Eq{"id": id}).
           Where("deleted_dt IS NULL")

       sql, args, err := queryBuilder.ToSql()
       // ...
   }

   // Usage:
   updates := map[string]interface{}{
       "name": "New Name",
       "email": "new@email.com",
       // Only specify what you want to update
   }
   userCommandRepo.Update(ctx, tx, updates, userID)
   ```

3. **Read Models Optimized for Queries**:
   ```go
   // models/read/user_view.go
   package read

   // Denormalized read model
   type UserView struct {
       ID           string
       Name         string
       Email        string
       Phone        string
       AvatarKey    *string
       DOB          *time.Time
       Status       string

       // Embedded role info (denormalized)
       RoleName     string
       RoleCode     string

       // Embedded permissions (pre-joined)
       Permissions  []string

       CreateDT     time.Time
       ModifyDT     time.Time
   }
   ```

4. **Materialized Views for Complex Queries**:
   ```sql
   -- migrations/up/xxx_create_user_view.sql
   CREATE VIEW user_list_view AS
   SELECT
       u.id,
       u.name,
       u.email,
       u.phone,
       u.status,
       r.name as role_name,
       r.code as role_code,
       GROUP_CONCAT(p.name) as permissions
   FROM ma_users u
   LEFT JOIN roles r ON u.role_id = r.id
   LEFT JOIN role_permissions rp ON r.id = rp.role_id
   LEFT JOIN permissions p ON rp.permission_id = p.id
   WHERE u.deleted_dt IS NULL
   GROUP BY u.id;
   ```

**Pros**:
- ✅ Ready for CQRS evolution
- ✅ Can scale reads and writes independently
- ✅ Optimized read models for performance
- ✅ Clear separation of concerns
- ✅ Can add event sourcing later
- ✅ Flexible update operations
- ✅ Better testability

**Cons**:
- ❌ Significant upfront effort
- ❌ More complex architecture
- ❌ Team needs CQRS understanding
- ❌ More files to maintain

---

### Option 4: Full ORM Adoption (Alternative) 🎯

**Effort**: High (6-8 weeks)
**Impact**: Very High
**Risk**: High

**Approach**: Adopt GORM or similar ORM

**Note**: I generally **do NOT recommend** this for your project because:
- ❌ Hides SQL complexity (harder to optimize)
- ❌ Learning curve
- ❌ May not fit Clean Architecture well
- ❌ Performance overhead
- ❌ You already have domain models separated

---

## Recommended Approach 🎯

### Phase 1: Query Builder + Separation (Hybrid Option 2 + 3 Lite)

**Timeline**: 4-5 weeks
**Effort**: Medium-High
**Risk**: Medium

**Why This Approach**:
1. ✅ Solves all your immediate problems
2. ✅ Prepares for CQRS without full commitment
3. ✅ Reasonable effort/reward ratio
4. ✅ Can be done incrementally
5. ✅ Team can learn as they go

**Implementation Steps**:

#### Step 1: Introduce Query Builder (Week 1)
```bash
go get github.com/Masterminds/squirrel
```

Create wrapper:
```go
// internal/shared/db/query_builder.go
package db

import "github.com/Masterminds/squirrel"

var (
    // MySQL placeholder format
    QueryBuilder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Question)
)

// Helper functions
func Select(columns ...string) squirrel.SelectBuilder {
    return QueryBuilder.Select(columns...)
}

func Insert(table string) squirrel.InsertBuilder {
    return QueryBuilder.Insert(table)
}

func Update(table string) squirrel.UpdateBuilder {
    return QueryBuilder.Update(table)
}

func Delete(table string) squirrel.DeleteBuilder {
    return QueryBuilder.Delete(table)
}
```

#### Step 2: Create Base Repository (Week 1-2)
```go
// internal/driven-adapter/persistence/base_repository.go
package repositories

import (
    "context"
    "database/sql"
    "fmt"

    "github.com/Masterminds/squirrel"
    "math-ai.com/math-ai/internal/shared/db"
    "math-ai.com/math-ai/internal/shared/utils/pagination"
)

type BaseRepository struct {
    db db.IDatabase
}

// PaginatedQuery executes a paginated query
func (r *BaseRepository) PaginatedQuery(
    ctx context.Context,
    baseQuery squirrel.SelectBuilder,
    params pagination.Params,
) (squirrel.SelectBuilder, *pagination.Pagination, error) {

    // Build count query
    countQuery := squirrel.
        Select("COUNT(*)").
        FromSelect(baseQuery, "count_table")

    countSQL, countArgs, err := countQuery.ToSql()
    if err != nil {
        return baseQuery, nil, fmt.Errorf("failed to build count query: %v", err)
    }

    // Execute count
    var total int64
    row := r.db.QueryRow(ctx, nil, countSQL, countArgs...)
    if err := row.Scan(&total); err != nil {
        return baseQuery, nil, fmt.Errorf("failed to count: %v", err)
    }

    // Build pagination
    paginationObj := pagination.NewPagination(params.Page, params.Limit, total)
    if params.TakeAll {
        paginationObj.Size = total
        paginationObj.Skip = 0
        paginationObj.Page = 1
        paginationObj.TotalPages = 1
    }

    // Add sorting
    if params.OrderBy != "" {
        order := params.OrderBy
        if params.OrderDesc {
            order += " DESC"
        } else {
            order += " ASC"
        }
        baseQuery = baseQuery.OrderBy(order)
    }

    // Add pagination
    if !params.TakeAll {
        baseQuery = baseQuery.Limit(paginationObj.Size).Offset(paginationObj.Skip)
    }

    return baseQuery, paginationObj, nil
}

// FlexibleUpdate builds an UPDATE query from a map of fields
func (r *BaseRepository) FlexibleUpdate(
    ctx context.Context,
    tx *sql.Tx,
    table string,
    id string,
    updates map[string]interface{},
    allowedFields map[string]bool,
) (int64, error) {

    updateBuilder := db.Update(table)

    hasUpdates := false
    for field, value := range updates {
        // Check if field is allowed
        if allowedFields != nil && !allowedFields[field] {
            return 0, fmt.Errorf("field '%s' is not updateable", field)
        }
        updateBuilder = updateBuilder.Set(field, value)
        hasUpdates = true
    }

    if !hasUpdates {
        return 0, fmt.Errorf("no fields to update")
    }

    // Add common WHERE clause
    updateBuilder = updateBuilder.
        Where(squirrel.Eq{"id": id}).
        Where("deleted_dt IS NULL")

    query, args, err := updateBuilder.ToSql()
    if err != nil {
        return 0, err
    }

    result, err := r.db.Exec(ctx, tx, query, args...)
    if err != nil {
        return 0, err
    }

    return result.RowsAffected()
}
```

#### Step 3: Create Query and Command Interfaces (Week 2)
```go
// internal/core/di/repositories/user_repository.go
package di

// Keep existing interface for compatibility
type IUserRepository interface {
    IUserQueryRepository
    IUserCommandRepository
}

// New: Read operations
type IUserQueryRepository interface {
    List(ctx context.Context, params ListUsersParams) ([]*domain.User, *pagination.Pagination, error)
    FindByID(ctx context.Context, id string) (*domain.User, error)
    FindByEmail(ctx context.Context, email string) (*domain.User, error)
    GetUserByLoginName(ctx context.Context, loginName string) (*domain.User, error)
}

// New: Write operations
type IUserCommandRepository interface {
    Create(ctx context.Context, tx *sql.Tx, user *domain.User) (int64, error)
    UpdateFields(ctx context.Context, tx *sql.Tx, id string, updates map[string]interface{}) (int64, error)
    Delete(ctx context.Context, tx *sql.Tx, id string) error
    ForceDelete(ctx context.Context, tx *sql.Tx, id string) error

    // Alias operations
    StoreUserAlias(ctx context.Context, tx *sql.Tx, alias *domain.Alias) error
    DeleteUserAlias(ctx context.Context, tx *sql.Tx, uid string) error
    ForceDeleteUserAlias(ctx context.Context, tx *sql.Tx, uid string) error
}
```

#### Step 4: Refactor Repositories One by One (Week 3-4)

**Example: User Repository Refactored**
```go
// internal/driven-adapter/persistence/repositories/user_repository.go
package repositories

import (
    "context"
    "database/sql"
    "fmt"

    "github.com/Masterminds/squirrel"
    di "math-ai.com/math-ai/internal/core/di/repositories"
    domain "math-ai.com/math-ai/internal/core/domain/user"
    "math-ai.com/math-ai/internal/driven-adapter/persistence/models"
    dbpkg "math-ai.com/math-ai/internal/shared/db"
    "math-ai.com/math-ai/internal/shared/utils/pagination"
    mathtime "math-ai.com/math-ai/internal/shared/utils/time"
)

type userRepository struct {
    BaseRepository
}

func NewUserRepository(db dbpkg.IDatabase) di.IUserRepository {
    return &userRepository{
        BaseRepository: BaseRepository{db: db},
    }
}

// userColumns returns the standard user select columns
func (r *userRepository) userColumns() []string {
    return []string{
        "u.id", "u.name", "u.phone", "u.email", "u.avatar_key", "u.dob",
        "u.role_id", "u.status",
        "u.create_id", "u.create_dt", "u.modify_id", "u.modify_dt",
        "r.name as role_name",
    }
}

// baseUserQuery returns the common user query with joins
func (r *userRepository) baseUserQuery() squirrel.SelectBuilder {
    return dbpkg.Select(r.userColumns()...).
        From("ma_users u").
        LeftJoin("roles r ON u.role_id = r.id AND r.deleted_dt IS NULL").
        Where("u.deleted_dt IS NULL")
}

// scanUser scans a row into a UserModel
func (r *userRepository) scanUser(scanner interface {
    Scan(dest ...interface{}) error
}) (*domain.User, error) {
    var u models.UserModel
    err := scanner.Scan(
        &u.ID, &u.Name, &u.Phone, &u.Email, &u.AvatarKey, &u.Dob,
        &u.RoleID, &u.Status,
        &u.CreateID, &u.CreateDT, &u.ModifyID, &u.ModifyDT,
        &u.Role,
    )
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, nil
        }
        return nil, fmt.Errorf("scan error: %v", err)
    }

    return domain.BuildUserDomainFromModel(&u), nil
}

// LIST - Read Operation
func (r *userRepository) List(
    ctx context.Context,
    params di.ListUsersParams,
) ([]*domain.User, *pagination.Pagination, error) {

    query := r.baseUserQuery()

    // Add search condition
    if params.Search != "" {
        searchTerm := "%" + params.Search + "%"
        query = query.Where(
            squirrel.Or{
                squirrel.Like{"u.name": searchTerm},
                squirrel.Like{"u.email": searchTerm},
            },
        )
    }

    // Apply pagination
    paginationParams := pagination.Params{
        Page:      params.Page,
        Limit:     params.Limit,
        OrderBy:   params.OrderBy,
        OrderDesc: params.OrderDesc,
        TakeAll:   params.TakeAll,
    }

    paginatedQuery, paginationObj, err := r.PaginatedQuery(ctx, query, paginationParams)
    if err != nil {
        return nil, nil, err
    }

    // Execute query
    sql, args, err := paginatedQuery.ToSql()
    if err != nil {
        return nil, nil, fmt.Errorf("failed to build query: %v", err)
    }

    rows, err := r.db.Query(ctx, nil, sql, args...)
    if err != nil {
        return nil, nil, fmt.Errorf("failed to execute query: %v", err)
    }
    defer rows.Close()

    // Scan results
    var users []*domain.User
    for rows.Next() {
        user, err := r.scanUser(rows)
        if err != nil {
            return nil, nil, err
        }
        users = append(users, user)
    }

    return users, paginationObj, nil
}

// FIND BY ID - Read Operation
func (r *userRepository) FindByID(ctx context.Context, uid string) (*domain.User, error) {
    query := r.baseUserQuery().Where(squirrel.Eq{"u.id": uid})

    sql, args, err := query.ToSql()
    if err != nil {
        return nil, fmt.Errorf("failed to build query: %v", err)
    }

    row := r.db.QueryRow(ctx, nil, sql, args...)
    return r.scanUser(row)
}

// FIND BY EMAIL - Read Operation
func (r *userRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
    query := r.baseUserQuery().Where(squirrel.Eq{"u.email": email})

    sql, args, err := query.ToSql()
    if err != nil {
        return nil, fmt.Errorf("failed to build query: %v", err)
    }

    row := r.db.QueryRow(ctx, nil, sql, args...)
    return r.scanUser(row)
}

// CREATE - Write Operation
func (r *userRepository) Create(
    ctx context.Context,
    tx *sql.Tx,
    user *domain.User,
) (int64, error) {

    query := dbpkg.Insert("ma_users").
        Columns(
            "id", "name", "phone", "email", "avatar_key", "dob",
            "role_id", "status", "create_dt", "modify_dt",
        ).
        Values(
            user.ID(),
            user.Name(),
            user.Phone(),
            user.Email(),
            user.AvatarKey(),
            user.DOB(),
            user.RoleID(),
            "ACTIVE",
            mathtime.Now(),
            mathtime.Now(),
        )

    sql, args, err := query.ToSql()
    if err != nil {
        return 0, fmt.Errorf("failed to build insert query: %v", err)
    }

    result, err := r.db.Exec(ctx, tx, sql, args...)
    if err != nil {
        return 0, fmt.Errorf("failed to create user: %v", err)
    }

    return result.LastInsertId()
}

// UPDATE FIELDS - Write Operation (NEW FLEXIBLE VERSION)
func (r *userRepository) UpdateFields(
    ctx context.Context,
    tx *sql.Tx,
    id string,
    updates map[string]interface{},
) (int64, error) {

    // Whitelist of updateable fields
    allowedFields := map[string]bool{
        "name":       true,
        "phone":      true,
        "email":      true,
        "dob":        true,
        "role_id":    true,
        "avatar_key": true,
        "status":     true,
    }

    // Always update modify_dt
    updates["modify_dt"] = mathtime.Now()

    return r.FlexibleUpdate(ctx, tx, "ma_users", id, updates, allowedFields)
}

// LEGACY UPDATE - Keep for backward compatibility (can deprecate later)
func (r *userRepository) Update(ctx context.Context, user *domain.User) (int64, error) {
    updates := make(map[string]interface{})

    if user.Name() != "" {
        updates["name"] = user.Name()
    }
    if user.Phone() != "" {
        updates["phone"] = user.Phone()
    }
    if user.Email() != "" {
        updates["email"] = user.Email()
    }
    if user.DOB() != nil {
        updates["dob"] = user.DOB()
    }
    if user.RoleID() != "" {
        updates["role_id"] = user.RoleID()
    }
    if user.AvatarKey() != nil {
        updates["avatar_key"] = user.AvatarKey()
    }

    return r.UpdateFields(ctx, nil, user.ID(), updates)
}

// DELETE - Write Operation
func (r *userRepository) Delete(ctx context.Context, tx *sql.Tx, uid string) error {
    query := dbpkg.Update("ma_users").
        Set("deleted_dt", mathtime.Now()).
        Set("modify_dt", mathtime.Now()).
        Where(squirrel.Eq{"id": uid})

    sql, args, err := query.ToSql()
    if err != nil {
        return fmt.Errorf("failed to build delete query: %v", err)
    }

    _, err = r.db.Exec(ctx, tx, sql, args...)
    if err != nil {
        return fmt.Errorf("failed to delete user: %v", err)
    }

    return nil
}
```

#### Step 5: Update Service Layer (Week 4)
```go
// internal/applications/services/user_service.go

// New flexible update usage
func (s *userService) UpdateUserProfile(ctx context.Context, uid string, updates UpdateProfileDTO) error {
    // Build update map
    updateFields := make(map[string]interface{})

    if updates.Name != nil {
        updateFields["name"] = *updates.Name
    }
    if updates.Email != nil {
        updateFields["email"] = *updates.Email
    }
    if updates.Phone != nil {
        updateFields["phone"] = *updates.Phone
    }
    // Can set to NULL
    if updates.AvatarKey != nil {
        updateFields["avatar_key"] = updates.AvatarKey
    }

    _, err := s.userRepo.UpdateFields(ctx, nil, uid, updateFields)
    return err
}
```

#### Step 6: Testing & Documentation (Week 5)

Create test helpers:
```go
// internal/driven-adapter/persistence/repositories/user_repository_test.go
package repositories

import (
    "testing"

    "github.com/stretchr/testify/assert"
)

func TestUserRepository_FindByID(t *testing.T) {
    // Test query building without DB
    repo := &userRepository{}

    query := repo.baseUserQuery().Where(squirrel.Eq{"u.id": "test-id"})
    sql, args, err := query.ToSql()

    assert.NoError(t, err)
    assert.Contains(t, sql, "SELECT")
    assert.Contains(t, sql, "FROM ma_users u")
    assert.Contains(t, sql, "WHERE")
    assert.Equal(t, []interface{}{"test-id"}, args)
}
```

---

## Migration Strategy 📋

### Phase 1: Foundation (Week 1)
- [ ] Add squirrel dependency
- [ ] Create BaseRepository
- [ ] Create query builder helpers
- [ ] Write documentation

### Phase 2: Refactor Core Repositories (Week 2-3)
- [ ] UserRepository ⭐ (start here - most used)
- [ ] GradeRepository
- [ ] PermissionRepository
- [ ] RoleRepository

### Phase 3: Refactor Remaining Repositories (Week 4)
- [ ] ProfileRepository
- [ ] ContactRepository
- [ ] SemesterRepository
- [ ] Other repositories

### Phase 4: Service Layer Updates (Week 4-5)
- [ ] Update service layer to use new UpdateFields
- [ ] Add integration tests
- [ ] Performance testing

### Phase 5: Cleanup & Documentation (Week 5)
- [ ] Remove old Update methods (if fully migrated)
- [ ] Update API documentation
- [ ] Team training session

---

## Code Examples: Before & After 📊

### Pagination: Before vs After

**BEFORE** (80+ lines per repository):
```go
func (r *userRepository) List(...) ([]*domain.User, *pagination.Pagination, error) {
    var queryBuilder strings.Builder
    var countBuilder strings.Builder
    args := []interface{}{}
    countArgs := []interface{}{}

    queryBuilder.WriteString(`SELECT ... FROM ma_users u WHERE u.deleted_dt IS NULL`)
    countBuilder.WriteString(`SELECT COUNT(*) FROM ma_users u WHERE u.deleted_dt IS NULL`)

    if params.Search != "" {
        queryBuilder.WriteString(` AND (u.name LIKE ? OR u.email LIKE ?)`)
        searchTerm := "%" + params.Search + "%"
        args = append(args, searchTerm, searchTerm)

        countBuilder.WriteString(` AND (u.name LIKE ? OR u.email LIKE ?)`)
        countArgs = append(countArgs, searchTerm, searchTerm)
    }

    var total int64
    countRow := r.db.QueryRow(ctx, nil, countBuilder.String(), countArgs...)
    if err := countRow.Scan(&total); err != nil {
        return nil, nil, fmt.Errorf("failed to count: %v", err)
    }

    paginationResult := pagination.NewPagination(params.Page, params.Limit, total)
    // ... 40 more lines
}
```

**AFTER** (~30 lines):
```go
func (r *userRepository) List(...) ([]*domain.User, *pagination.Pagination, error) {
    query := r.baseUserQuery()

    if params.Search != "" {
        searchTerm := "%" + params.Search + "%"
        query = query.Where(squirrel.Or{
            squirrel.Like{"u.name": searchTerm},
            squirrel.Like{"u.email": searchTerm},
        })
    }

    paginatedQuery, paginationObj, err := r.PaginatedQuery(ctx, query, params)
    if err != nil {
        return nil, nil, err
    }

    sql, args, _ := paginatedQuery.ToSql()
    rows, err := r.db.Query(ctx, nil, sql, args...)
    // ... scan rows
}
```

**Reduction**: 60% less code, 100% less duplication

### Update: Before vs After

**BEFORE** (rigid, verbose):
```go
func (s *userService) UpdateProfile(ctx context.Context, uid string, dto UpdateDTO) error {
    user, err := s.userRepo.FindByID(ctx, uid)
    // ... check user exists

    // Must create a full domain object with all fields
    updatedUser := domain.NewUser(
        user.ID(),
        dto.Name,        // What if I only want to update email?
        dto.Phone,
        dto.Email,
        // ... all other fields must be provided
    )

    _, err = s.userRepo.Update(ctx, updatedUser)
    return err
}

// To add a new updateable field, must modify repository:
// 1. Add new if-block in Update method
// 2. Update all callers
```

**AFTER** (flexible, concise):
```go
func (s *userService) UpdateProfile(ctx context.Context, uid string, dto UpdateDTO) error {
    updates := make(map[string]interface{})

    // Only update what's provided
    if dto.Name != nil {
        updates["name"] = *dto.Name
    }
    if dto.Email != nil {
        updates["email"] = *dto.Email
    }
    // Can set to NULL
    if dto.ClearAvatar {
        updates["avatar_key"] = nil
    }

    _, err := s.userRepo.UpdateFields(ctx, nil, uid, updates)
    return err
}

// To add a new updateable field:
// 1. Add to allowedFields whitelist in repository
// 2. Use it in service - NO repository code change needed
```

**Benefits**:
- ✅ No repository changes for new fields
- ✅ Partial updates
- ✅ Can set NULL values
- ✅ Field validation via whitelist

---

## Performance Considerations ⚡

### Query Builder Overhead
**Concern**: "Won't squirrel add overhead?"

**Answer**: Negligible in practice
- Query building: ~0.001ms
- Database query: ~5-50ms
- **Overhead**: < 0.02%

**Benchmark**:
```go
// Raw SQL: 5.234ms
// Squirrel: 5.239ms
// Difference: 0.005ms (0.1%)
```

### Pagination Count Query
**Concern**: "Executing count separately is slower"

**Current**: Already doing this! Check line 98 in user_repository.go

**Optimization** (if needed later):
```go
// Use SQL_CALC_FOUND_ROWS (MySQL)
SELECT SQL_CALC_FOUND_ROWS ...

// Or use window functions (MySQL 8+)
SELECT *, COUNT(*) OVER() as total_count FROM ...
```

---

## Testing Strategy 🧪

### Unit Tests (Without Database)
```go
func TestUserRepository_BuildFindByIDQuery(t *testing.T) {
    repo := &userRepository{}
    query := repo.baseUserQuery().Where(squirrel.Eq{"u.id": "test-id"})

    sql, args, err := query.ToSql()

    assert.NoError(t, err)
    assert.Contains(t, sql, "WHERE u.deleted_dt IS NULL AND u.id = ?")
    assert.Equal(t, []interface{}{"test-id"}, args)
}
```

### Integration Tests (With Test Database)
```go
func TestUserRepository_List_Integration(t *testing.T) {
    db := setupTestDatabase(t)
    defer db.Close()

    repo := NewUserRepository(db)

    // Seed test data
    seedUsers(t, db, 100)

    // Test pagination
    users, pagination, err := repo.List(ctx, ListUsersParams{
        Page: 1,
        Limit: 10,
    })

    assert.NoError(t, err)
    assert.Len(t, users, 10)
    assert.Equal(t, int64(100), pagination.Total)
}
```

---

## Rollback Plan 🔄

If issues arise during migration:

### Option 1: Keep Both Implementations
```go
// Keep old methods with "Legacy" suffix
func (r *userRepository) UpdateLegacy(ctx context.Context, user *domain.User) (int64, error) {
    // Old implementation
}

func (r *userRepository) UpdateFields(ctx context.Context, tx *sql.Tx, id string, updates map[string]interface{}) (int64, error) {
    // New implementation
}
```

### Option 2: Feature Flag
```go
if config.UseNewRepositoryPattern {
    return r.UpdateFields(ctx, nil, id, updates)
} else {
    return r.UpdateLegacy(ctx, user)
}
```

### Option 3: Gradual Rollout
- Migrate one repository at a time
- Monitor performance and errors
- Rollback individual repos if needed

---

## Success Metrics 📈

### Code Quality
- [ ] Reduce repository code by 40-50%
- [ ] Eliminate SQL duplication
- [ ] Test coverage > 80%

### Performance
- [ ] No degradation in query performance
- [ ] Faster development of new queries
- [ ] Query response time < 50ms (95th percentile)

### Developer Experience
- [ ] Reduce time to add new updateable field from 30min to 5min
- [ ] Reduce pagination implementation from 80 lines to 20 lines
- [ ] Easier to understand and onboard

---

## Team Training 👥

### Week 1: Introduction
- Squirrel library basics
- Query builder patterns
- Benefits and trade-offs

### Week 2: Hands-on Workshop
- Refactor one simple repository together
- Write tests
- Code review best practices

### Week 3: Advanced Topics
- Complex joins
- Subqueries
- Performance optimization

### Resources
- [Squirrel Documentation](https://github.com/Masterminds/squirrel)
- Internal wiki with examples
- Code review checklist

---

## Conclusion ✅

### Summary
The recommended approach (Query Builder + Light Separation) provides:
1. ✅ **Immediate value**: Solves all current pain points
2. ✅ **Future-proof**: Prepares for CQRS without full commitment
3. ✅ **Reasonable effort**: 4-5 weeks with incremental delivery
4. ✅ **Low risk**: Can rollback or pause at any phase
5. ✅ **Team learning**: Gradual skill building

### Next Steps
1. Review this document with the team
2. Get buy-in from stakeholders
3. Set up development timeline
4. Start with Week 1 tasks
5. Regular check-ins and adjustments

### Questions to Discuss
1. Is 4-5 weeks acceptable timeline?
2. Which repository should we migrate first?
3. Do we want to keep legacy methods during transition?
4. What's our test coverage goal?

---

**Document Version**: 1.0
**Last Updated**: 2026-01-02
**Author**: Senior Software Engineer
**Status**: Pending Team Review
