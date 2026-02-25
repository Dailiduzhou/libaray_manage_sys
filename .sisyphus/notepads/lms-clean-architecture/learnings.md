# Learnings - Request Model Analysis

## Date: 2026-02-25

### Task: Analyze existing request models in backend/models/requests.go

## Findings

### Existing Request Models (5 total):
1. **RegisterRequest** - username, password
2. **LoginRequest** - username, password  
3. **CreateBookRequest** - title, author, summary, cover (multipart.FileHeader), initial_stock
4. **UpdateBookRequest** - id, title, author, summary, cover (multipart.FileHeader), stock, total_stock
5. **FindBookRequest** - id

### Service Interface Expectations vs Available DTOs:

| Service Method | Parameters | Has Request Model? |
|---------------|------------|-------------------|
| CreateBook | title, author, summary, coverPath, initialStock | ✅ CreateBookRequest |
| GetBooks | title, author, summary | ❌ MISSING |
| UpdateBook | id, title, author, summary, coverPath, stock, totalStock | ✅ UpdateBookRequest |
| DeleteBook | id | ✅ FindBookRequest |
| GetBookByID | id | ✅ FindBookRequest |
| Register | username, password | ✅ RegisterRequest |
| Login | username, password | ✅ LoginRequest |
| GetUserByUsername | username | ❌ MISSING |
| BorrowBook | userID, bookID | ❌ MISSING |
| ReturnBook | userID, bookID | ❌ MISSING |
| GetUserRecords | userID | ❌ MISSING |
| GetRecordsByUserID | userID | ❌ MISSING |
| GetAllRecords | none | ✅ (no params needed) |

### Conclusion
**ADDITIONAL DTOs ARE REQUIRED** - 5 new request models needed:

1. **GetBooksRequest** - query params for title, author, summary (for filtering)
2. **FindUserRequest** - username param for GetUserByUsername
3. **BorrowBookRequest** - userID, bookID for BorrowBook
4. **ReturnBookRequest** - userID, bookID for ReturnBook
5. **UserRecordsRequest** - userID for GetUserRecords/GetRecordsByUserID

### Recommendation
Create new DTOs in requests.go for the missing service layer operations, especially for:
- User lookup operations
- Borrow/Return book operations  
- Book query/filter operations



---

## Model Analysis (backend/models/models.go)

### Task: Check if models need adjustments for Repository/Service layers

### Repository Interface Requirements vs Model Fields:

| Model | Repository Methods | Required Fields | Status |
|-------|-------------------|-----------------|--------|
| Book | Create, GetByID, Update, Delete, FindAll, FindByTitle, FindByAuthor | ID, Title, Author, Summary, CoverPath, InitialStock, Stock, TotalStock, CreatedAt, UpdatedAt | ✅ Complete |
| User | Create, GetByID, GetByUsername, FindAll | ID, Username, Password, Role, CreatedAt, UpdatedAt | ✅ Complete |
| BorrowRecord | Create, GetByID, GetByUserID, GetByBookID, GetByUserAndBook, FindAll | ID, UserID, BookID, BorrowDate, ReturnDate, Status, User (relation), Book (relation), DeletedAt (soft delete) | ✅ Complete |

### Findings:
- **Book model**: Has all fields for CRUD + title/author search - ✅ No changes needed
- **User model**: Has all fields for user operations - ✅ No changes needed  
- **BorrowRecord model**: Has all fields for borrow operations + soft delete + relations - ✅ No changes needed

### Note:
- Minor typo in line 36: `gorm:"defualt:0"` should be `gorm:"default:0"` (typo in "default")
- This is a cosmetic issue and doesn't affect functionality

### Conclusion
**NO ADJUSTMENTS NEEDED** - All three models (Book, User, BorrowRecord) are complete and ready for Repository/Service layers.