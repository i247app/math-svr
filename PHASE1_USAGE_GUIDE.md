# Phase 1 Usage Guide - Repository Refactoring

This guide shows you how to use the new Phase 1 improvements without any external libraries.

## What's New in Phase 1?

### 1. **BaseRepository** - Common repository operations
### 2. **QueryHelper** - SQL query building utilities
### 3. **Query Constants** - Centralized SQL query management
### 4. **Pagination Params** - Standardized pagination parameters

---

## 1. BaseRepository

The `BaseRepository` provides common operations that eliminate code duplication across repositories.

### How to Use

**Embed BaseRepository in your repository:**

```go
type userRepository struct {
    BaseRepository  // Embed to inherit methods
}

func NewUserRepository(db db.IDatabase) di.IUserRepository {
    return &userRepository{
        BaseRepository: NewBaseRepository(db),
    }
}
```

### Available Methods

#### A. PaginatedList - Automated Pagination

**Before (80+ lines):**
```go
func (r *userRepository) List(...) {
    var queryBuilder strings.Builder
    var countBuilder strings.Builder
    args := []interface{}{}
    countArgs := []interface{}{}

    // Build main query
    queryBuilder.WriteString(`SELECT ...`)

    // Build count query separately (DUPLICATION!)
    countBuilder.WriteString(`SELECT COUNT(*) ...`)

    // Add search to both queries
    if search != "" {
        queryBuilder.WriteString(` AND name LIKE ?`)
        args = append(args, "%"+search+"%")

        countBuilder.WriteString(` AND name LIKE ?`)
        countArgs = append(countArgs, "%"+search+"%")
    }

    // Execute count
    var total int64
    countRow := r.db.QueryRow(ctx, nil, countBuilder.String(), countArgs...)
    countRow.Scan(&total)

    // Build pagination
    paginationObj := pagination.NewPagination(...)

    // Add ORDER BY
    if orderBy != "" {
        queryBuilder.WriteString(` ORDER BY ` + orderBy)
        if orderDesc {
            queryBuilder.WriteString(` DESC`)
        }
    }

    // Add LIMIT OFFSET
    queryBuilder.WriteString(` LIMIT ? OFFSET ?`)
    args = append(args, limit, offset)

    // Finally execute
    rows, err := r.db.Query(ctx, nil, queryBuilder.String(), args...)
    //...
}
```

**After (20-30 lines):**
```go
func (r *userRepository) List(ctx context.Context, params di.ListUsersParams) ([]*domain.User, *pagination.Pagination, error) {
    // Base query
    baseQuery := queries.UserListQuery
    countQuery := queries.UserListCountQuery
    args := []interface{}{}

    // Add search if needed
    if params.Search != "" {
        baseQuery, countQuery, args = queries.UserQueries{}.BuildListQueryWithSearch(params.Search)
    }

    // Build pagination params
    paginationParams := pagination.Params{
        Page:      params.Page,
        Limit:     params.Limit,
        OrderBy:   params.OrderBy,
        OrderDesc: params.OrderDesc,
        TakeAll:   params.TakeAll,
    }

    // Execute with automatic count and pagination
    query, queryArgs, paginationObj, err := r.PaginatedList(
        ctx,
        baseQuery,
        args,
        countQuery,
        args,
        paginationParams,
    )
    if err != nil {
        return nil, nil, err
    }

    // Execute final query
    rows, err := r.db.Query(ctx, nil, query, queryArgs...)
    if err != nil {
        return nil, nil, err
    }
    defer rows.Close()

    // Scan results (same as before)
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
```

**Benefits:**
- ✅ Automatic count query execution
- ✅ No duplication of WHERE conditions
- ✅ Automatic ORDER BY and LIMIT/OFFSET
- ✅ 60% less code

---

#### B. FlexibleUpdate - Map-Based Updates

**Before (rigid, many if-blocks):**
```go
func (r *userRepository) Update(ctx context.Context, user *domain.User) (int64, error) {
    var queryBuilder strings.Builder
    args := []interface{}{}

    queryBuilder.WriteString("UPDATE ma_users SET ")
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

    if user.Email() != "" {
        updates = append(updates, "email = ?")
        args = append(args, user.Email())
    }

    // ... 10 more if-blocks

    updates = append(updates, "modify_dt = ?")
    args = append(args, mathtime.Now())

    queryBuilder.WriteString(strings.Join(updates, ", "))
    queryBuilder.WriteString(" WHERE id = ? AND deleted_dt IS NULL")
    args = append(args, user.ID())

    result, err := r.db.Exec(ctx, nil, queryBuilder.String(), args...)
    return result.RowsAffected()
}
```

**After (flexible, map-based):**
```go
// New method with map-based updates
func (r *userRepository) UpdateFields(
    ctx context.Context,
    tx *sql.Tx,
    id string,
    updates map[string]interface{},
) (int64, error) {

    // Whitelist of updateable fields (security!)
    allowedFields := map[string]bool{
        "name":       true,
        "phone":      true,
        "email":      true,
        "dob":        true,
        "role_id":    true,
        "avatar_key": true,
        "status":     true,
    }

    // Use BaseRepository method
    return r.FlexibleUpdate(ctx, tx, "ma_users", id, updates, allowedFields)
}

// Usage in service:
func (s *userService) UpdateProfile(ctx context.Context, uid string, dto UpdateProfileDTO) error {
    updates := make(map[string]interface{})

    // Only update what's provided
    if dto.Name != nil {
        updates["name"] = *dto.Name
    }
    if dto.Email != nil {
        updates["email"] = *dto.Email
    }
    if dto.Phone != nil {
        updates["phone"] = *dto.Phone
    }

    // Can set to NULL
    if dto.ClearAvatar {
        updates["avatar_key"] = nil
    }

    _, err := s.userRepo.UpdateFields(ctx, nil, uid, updates)
    return err
}
```

**Benefits:**
- ✅ No if-blocks for each field
- ✅ Can update any combination of fields
- ✅ Can set fields to NULL
- ✅ Field whitelist for security
- ✅ Adding new updateable field = just add to whitelist

---

#### C. BatchInsert - Efficient Bulk Inserts

**Before (N queries for N records):**
```go
func (s *userService) CreateMultipleUsers(ctx context.Context, users []*domain.User) error {
    for _, user := range users {
        _, err := s.userRepo.Create(ctx, nil, user)
        if err != nil {
            return err
        }
    }
    return nil
}
```

**After (1 query for N records):**
```go
func (r *userRepository) BatchCreate(
    ctx context.Context,
    tx *sql.Tx,
    users []*domain.User,
) error {

    columns := []string{"id", "name", "phone", "email", "avatar_key", "dob", "role_id", "status", "create_dt", "modify_dt"}

    values := make([][]interface{}, 0, len(users))
    for _, user := range users {
        row := []interface{}{
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
        }
        values = append(values, row)
    }

    _, err := r.BatchInsert(ctx, tx, "ma_users", columns, values)
    return err
}
```

**Benefits:**
- ✅ 10x-100x faster for bulk inserts
- ✅ Single transaction
- ✅ Reduced network overhead

---

#### D. SoftDelete & HardDelete - Standardized Deletion

**Before:**
```go
func (r *userRepository) Delete(ctx context.Context, tx *sql.Tx, uid string) error {
    query := `UPDATE ma_users SET deleted_dt = ?, modify_dt = ? WHERE id = ?`
    _, err := r.db.Exec(ctx, tx, query, mathtime.Now(), mathtime.Now(), uid)
    return err
}

func (r *userRepository) ForceDelete(ctx context.Context, tx *sql.Tx, uid string) error {
    query := `DELETE FROM ma_users WHERE id = ?`
    _, err := r.db.Exec(ctx, tx, query, uid)
    return err
}
```

**After:**
```go
func (r *userRepository) Delete(ctx context.Context, tx *sql.Tx, uid string) error {
    _, err := r.SoftDelete(ctx, tx, "ma_users", uid)
    return err
}

func (r *userRepository) ForceDelete(ctx context.Context, tx *sql.Tx, uid string) error {
    _, err := r.HardDelete(ctx, tx, "ma_users", uid)
    return err
}
```

**Benefits:**
- ✅ Consistent deletion logic across all repositories
- ✅ Less boilerplate

---

## 2. QueryHelper

Utility functions for building SQL fragments.

### Available Methods

#### A. BuildWhereClause

```go
conditions := map[string]interface{}{
    "status": "active",
    "role_id": "admin",
}

whereClause, args := r.Helper().BuildWhereClause(conditions)
// Returns: "WHERE status = ? AND role_id = ?", []interface{}{"active", "admin"}
```

#### B. BuildLikeConditions

```go
fields := []string{"name", "email", "phone"}
searchTerm := "john"

condition, args := r.Helper().BuildLikeConditions(fields, searchTerm)
// Returns: "(name LIKE ? OR email LIKE ? OR phone LIKE ?)", []interface{}{"%john%", "%john%", "%john%"}
```

#### C. BuildUpdateSet

```go
updates := map[string]interface{}{
    "name": "John Doe",
    "email": "john@example.com",
}

setClause, args := r.Helper().BuildUpdateSet(updates)
// Returns: "SET name = ?, email = ?", []interface{}{"John Doe", "john@example.com"}
```

#### D. BuildInCondition

```go
ids := []interface{}{"id1", "id2", "id3"}

condition, args := r.Helper().BuildInCondition("user_id", ids)
// Returns: "user_id IN (?, ?, ?)", []interface{}{"id1", "id2", "id3"}
```

#### E. BuildOrderBy

```go
orderClause := r.Helper().BuildOrderBy("created_at", true)
// Returns: "ORDER BY created_at DESC"
```

#### F. BuildPagination

```go
paginationClause := r.Helper().BuildPagination(10, 20)
// Returns: "LIMIT 10 OFFSET 20"
```

---

## 3. Query Constants

Centralize your SQL queries for better management.

### Structure

Create a `queries` package with query constants for each repository:

```
internal/driven-adapter/persistence/queries/
├── user_queries.go
├── grade_queries.go
├── permission_queries.go
└── role_queries.go
```

### Example

**queries/user_queries.go:**
```go
package queries

const UserSelectColumns = `u.id, u.name, u.email, u.status`

const UserBaseSelect = `SELECT ` + UserSelectColumns + `
    FROM ma_users u
    WHERE u.deleted_dt IS NULL`

const UserFindByID = UserBaseSelect + ` AND u.id = ?`
const UserFindByEmail = UserBaseSelect + ` AND u.email = ?`
```

**Usage in repository:**
```go
func (r *userRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
    row := r.db.QueryRow(ctx, nil, queries.UserFindByID, id)
    return r.scanUser(row)
}
```

**Benefits:**
- ✅ All queries in one place
- ✅ Easy to review and optimize
- ✅ Reusable query fragments
- ✅ Version control friendly

---

## 4. Complete Example: Refactoring User Repository

Here's a complete before/after example:

### List Method - Before
```go
func (r *userRepository) List(ctx context.Context, params di.ListUsersParams) ([]*domain.User, *pagination.Pagination, error) {
    var queryBuilder strings.Builder
    args := []interface{}{}

    queryBuilder.WriteString(`
        SELECT u.id, u.name, u.phone, u.email, u.avatar_key, u.dob,
               u.role_id, u.status,
               u.create_id, u.create_dt, u.modify_id, u.modify_dt,
               r.name as role_name
        FROM ma_users u
        LEFT JOIN roles r ON u.role_id = r.id AND r.deleted_dt IS NULL
        WHERE u.deleted_dt IS NULL
    `)

    if params.Search != "" {
        queryBuilder.WriteString(` AND (u.name LIKE ? OR u.email LIKE ?)`)
        searchTerm := "%" + params.Search + "%"
        args = append(args, searchTerm, searchTerm)
    }

    countQuery := "SELECT COUNT(*) FROM ma_users u WHERE u.deleted_dt IS NULL"
    countArgs := []interface{}{}
    if params.Search != "" {
        countQuery += ` AND (u.name LIKE ? OR u.email LIKE ?)`
        searchTerm := "%" + params.Search + "%"
        countArgs = append(countArgs, searchTerm, searchTerm)
    }

    var total int64
    countRow := r.db.QueryRow(ctx, nil, countQuery, countArgs...)
    if err := countRow.Scan(&total); err != nil {
        return nil, nil, fmt.Errorf("failed to count users: %v", err)
    }

    paginationResult := pagination.NewPagination(params.Page, params.Limit, total)
    if params.TakeAll {
        paginationResult.Size = total
        paginationResult.Skip = 0
        paginationResult.Page = 1
        paginationResult.TotalPages = 1
    }

    if params.OrderBy != "" {
        queryBuilder.WriteString(fmt.Sprintf(" ORDER BY u.%s", params.OrderBy))
        if params.OrderDesc {
            queryBuilder.WriteString(" DESC")
        }
    }

    if !params.TakeAll {
        queryBuilder.WriteString(` LIMIT ? OFFSET ?`)
        args = append(args, paginationResult.Size, paginationResult.Skip)
    }

    rows, err := r.db.Query(ctx, nil, queryBuilder.String(), args...)
    if err != nil {
        return nil, nil, fmt.Errorf("failed to list users: %v", err)
    }
    defer rows.Close()

    var users []*domain.User
    for rows.Next() {
        var u models.UserModel
        if err := rows.Scan(
            &u.ID, &u.Name, &u.Phone, &u.Email, &u.AvatarKey, &u.Dob,
            &u.RoleID, &u.Status,
            &u.CreateID, &u.CreateDT, &u.ModifyID, &u.ModifyDT,
            &u.Role,
        ); err != nil {
            return nil, nil, fmt.Errorf("scan error: %v", err)
        }
        users = append(users, domain.BuildUserDomainFromModel(&u))
    }

    return users, paginationResult, nil
}
```

### List Method - After
```go
func (r *userRepository) List(ctx context.Context, params di.ListUsersParams) ([]*domain.User, *pagination.Pagination, error) {
    // Get base queries
    baseQuery := queries.UserListQuery
    countQuery := queries.UserListCountQuery
    args := []interface{}{}

    // Add search if provided
    if params.Search != "" {
        baseQuery, countQuery, args = queries.UserQueries{}.BuildListQueryWithSearch(params.Search)
    }

    // Build pagination params
    paginationParams := pagination.Params{
        Page:      params.Page,
        Limit:     params.Limit,
        OrderBy:   params.OrderBy,
        OrderDesc: params.OrderDesc,
        TakeAll:   params.TakeAll,
    }

    // Execute with automatic pagination
    query, queryArgs, paginationObj, err := r.PaginatedList(
        ctx,
        baseQuery,
        args,
        countQuery,
        args,
        paginationParams,
    )
    if err != nil {
        return nil, nil, err
    }

    // Execute query
    rows, err := r.db.Query(ctx, nil, query, queryArgs...)
    if err != nil {
        return nil, nil, err
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

// Helper method to scan user (reusable)
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
```

**Improvements:**
- ✅ 40% less code
- ✅ No query duplication
- ✅ Easier to read and maintain
- ✅ Reusable scanUser method
- ✅ Query constants for better organization

---

## Migration Checklist

When refactoring a repository to use Phase 1:

- [ ] 1. Embed `BaseRepository` in your repository struct
- [ ] 2. Create query constants in `queries/` package
- [ ] 3. Replace pagination logic with `PaginatedList`
- [ ] 4. Add `UpdateFields` method using `FlexibleUpdate`
- [ ] 5. Replace delete methods with `SoftDelete`/`HardDelete`
- [ ] 6. Extract scan logic into helper method
- [ ] 7. Test all methods
- [ ] 8. Update service layer to use new methods

---

## Next Steps

After Phase 1 is stable:

- **Phase 2**: Split Read and Write operations (CQRS preparation)
- **Phase 3**: Optimize read models with denormalization
- **Phase 4**: Add event sourcing (if needed)

---

## Questions?

Refer to:
- `internal/driven-adapter/persistence/base_repository.go` - BaseRepository implementation
- `internal/shared/db/query_builder.go` - QueryHelper implementation
- `internal/driven-adapter/persistence/queries/user_queries.go` - Query constants example
- `REPOSITORY_REFACTORING_PLAN.md` - Full refactoring plan
