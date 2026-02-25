# 图书馆管理系统 Clean Architecture 重构计划

## TL;DR

> **快速摘要**: 将基于 Gin 的图书馆管理系统从紧耦合架构重构为 Clean Architecture，包含 Repository + Service 层，实现依赖注入以支持纯内存单元测试。

> **交付物**:
> - `repositories/interfaces.go` - 数据持久化接口定义
> - `repositories/book_repo.go`, `user_repo.go`, `borrow_repo.go` - GORM 实现
> - `services/interfaces.go` - 业务逻辑接口定义
> - `services/book_service.go`, `user_service.go`, `borrow_service.go` - 服务实现
> - `controllers/controller.go` - 修改为依赖注入的 Handler
> - `main.go` - 依赖注入 wiring
> - 测试示例: Handler 单元测试 + Repository 集成测试

> **预估工作量**: Large
> **并行执行**: YES - 4 waves
> **关键路径**: 接口定义 → GORM 实现 → Service 实现 → Handler 改造 → Wiring

---

## Context

### 原始需求
用户要求将一个基于 Gin 框架的图书馆管理系统 RESTful API 从紧耦合重构为 Clean Architecture，实现：
1. 提取核心领域模型 (已有 Book, User, BorrowRecord)
2. 创建 Repository 接口负责数据持久化
3. 创建 Service 接口负责业务逻辑
4. 实现依赖注入 - Handler 依赖接口而非 `*gorm.DB`
5. 添加 go:generate mockgen 标记

### 访谈摘要
**关键讨论**:
- **架构选择**: Repository + Service 层
- **测试策略**: Handler 层用 gomock + httptest 单元测试; Repository 层用 testcontainers-go 集成测试
- **目录结构**: 在 backend 目录创建 repositories/, services/
- **模块路径**: `github.com/Dailiduzhou/library_manage_sys`

### Metis 审查
**识别的边界问题** (已解决):
- 文件上传逻辑保持 在 Handler 层 (utils.SaveImages 依赖 gin.Context)
- 事务管理由 Service 层传入 `*gorm.DB` 参数
- Session 处理保持在 Handler 层，只传 userID 到 Service

**需要锁定**:
- API 端点路径不变
- HTTP 方法不变
- JSON 字段名不变
- 错误消息不变
- HTTP 状态码不变

---

## Work Objectives

### 核心目标
将 `controllers/controller.go` 中的 12 个 Handler 函数从直接依赖 `config.DB` 重构为依赖 Repository/Service 接口，实现可测试性。

### 具体交付物
1. **Repository 接口** (repositories/interfaces.go)
   - `BookRepository` 接口
   - `UserRepository` 接口
   - `BorrowRepository` 接口
   - 每个接口上方添加 `//go:generate mockgen ...` 注释

2. **Repository 实现** (repositories/*.go)
   - `GormBookRepository`
   - `GormUserRepository`
   - `GormBorrowRepository`

3. **Service 接口** (services/interfaces.go)
   - `BookService` 接口
   - `UserService` 接口
   - `BorrowService` 接口
   - 每个接口上方添加 `//go:generate mockgen ...` 注释

4. **Service 实现** (services/*.go)
   - `BookServiceImpl`
   - `UserServiceImpl`
   - `BorrowServiceImpl`

5. **修改 Handler** (controllers/controller.go)
   - 修改为接收接口而非 `*gorm.DB`
   - 添加构造函数 `NewXXXHandler`

6. **依赖注入 Wiring** (main.go)
   - 创建接口实例
   - 注入到 Handlers

7. **测试示例**
   - Handler 单元测试示例 (gomock + httptest)
   - Repository 集成测试示例 (testcontainers-go)

### 定义完成
- [ ] `go build ./...` 成功
- [ ] `go generate ./...` 生成 mock 文件
- [ ] 现有 API 端点功能不变

### Must Have
- 所有 12 个 Handler 函数必须支持依赖注入
- 必须保留现有的 API 契约 (端点、响应格式、错误消息)
- 必须添加 go:generate 注释

### Must NOT Have (Guardrails)
- 不修改任何 API 端点路径
- 不修改 HTTP 方法映射
- 不修改 JSON 字段名
- 不修改错误消息字符串
- 不修改 HTTP 状态码
- 不修改数据库 schema

---

## Verification Strategy

### 测试决策
- **基础设施存在**: YES
- **自动化测试**: Tests-after (先实现，再添加测试)
- **测试框架**: gomock + httptest + testcontainers-go

### QA 策略
每个任务必须包含 Agent-Executable QA Scenarios:
- Backend: 使用 `go build` 验证编译
- Integration: 使用 `go test` 运行测试

---

## Execution Strategy

### 并行执行 Waves

```
Wave 1 (立即开始 - 基础接口):
├── T1: 创建 Repository 接口 + go:generate 注释
├── T2: 创建 Service 接口 + go:generate 注释  
├── T3: 创建 models/requests.go 请求模型 (如有需要)
└── T4: 更新 models/models.go (如有需要)

Wave 2 (T1-T4 后 - Repository 实现):
├── T5: 实现 GormBookRepository
├── T6: 实现 GormUserRepository
├── T7: 实现 GormBorrowRepository
└── T8: 创建 repositories/repositories.go 统一导出

Wave 3 (T5-T8 后 - Service 实现):
├── T9: 实现 BookServiceImpl
├── T10: 实现 UserServiceImpl
├── T11: 实现 BorrowServiceImpl
└── T12: 创建 services/services.go 统一导出

Wave 4 (T9-T12 后 - Handler 改造):
├── T13: 修改 BookHandler 依赖注入
├── T14: 修改 UserHandler 依赖注入
├── T15: 修改 BorrowHandler 依赖注入
├── T16: 更新 controllers/controller.go 整体结构
└── T17: 更新 main.go 依赖注入 Wiring

Wave 5 (最后 - 测试 + 验证):
├── T18: 创建 Handler 单元测试示例
├── T19: 创建 Repository 集成测试示例
├── T20: 运行 go generate 验证 mock 生成
├── T21: 运行 go build 验证编译
└── T22: 运行 go test 验证测试通过
```

### 依赖矩阵
- **T1-T4**: — — 5-8
- **T5-T8**: 1 — 9-12
- **T9-T12**: 2, 5-8 — 13-17
- **T13-T17**: 9-12 — 18-22
- **T18-T22**: 13-17 — Final

---

## TODOs

- [ ] T1. 创建 Repository 接口定义 (BookRepository, UserRepository, BorrowRepository)

  **What to do**:
  - 在 `backend/repositories/interfaces.go` 创建文件
  - 定义 `BookRepository` 接口: Create, GetByID, Update, Delete, FindAll, FindByTitle, FindByAuthor
  - 定义 `UserRepository` 接口: Create, GetByID, GetByUsername, FindAll
  - 定义 `BorrowRepository` 接口: Create, GetByID, GetByUserID, GetByBookID, GetByUserAndBook, FindAll
  - 每个接口上方添加 `//go:generate mockgen -source=$GOFILE -destination=mocks/mock_$GOFILE -package=mocks`

  **Must NOT do**:
  - 不包含 GORM 特定类型在接口方法签名中 (使用 models.Xxx 而非 *gorm.DB 作为返回类型)

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: 需要设计清晰的接口抽象，需要对现有业务逻辑有深入理解
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with T2, T3, T4)
  - **Blocks**: T5-T8 (Repository 实现)
  - **Blocked By**: None

  **References**:
  - `backend/controllers/controller.go` - 分析现有 DB 操作模式
  - `backend/models/models.go` - 现有模型定义

  **QA Scenarios**:
  - [ ] `go build -n ./repositories` 验证语法正确
  - [ ] 检查接口方法签名与 Handler 中的使用匹配

---

- [ ] T2. 创建 Service 接口定义

  **What to do**:
  - 在 `backend/services/interfaces.go` 创建文件
  - 定义 `BookService` 接口: CreateBook, GetBooks, UpdateBook, DeleteBook, GetBookByID
  - 定义 `UserService` 接口: Register, Login (密码验证), GetUserByUsername
  - 定义 `BorrowService` 接口: BorrowBook, ReturnBook, GetUserRecords, GetAllRecords, GetRecordsByUserID
  - 每个接口上方添加 `//go:generate mockgen ...` 注释
  - Service 层处理业务逻辑 (库存检查、事务管理、错误转换)

  **Must NOT do**:
  - 不在 Service 接口中使用 gin.Context 类型
  - 不包含 HTTP 响应相关逻辑

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: 需要封装业务逻辑，需要理解现有 Handler 中的业务规则
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with T1, T3, T4)
  - **Blocks**: T9-T12 (Service 实现)
  - **Blocked By**: None

  **References**:
  - `backend/controllers/controller.go` - 分析业务逻辑 (库存检查、事务、错误处理)

  **QA Scenarios**:
  - [ ] `go build -n ./services` 验证语法正确
  - [ ] 检查 Service 方法参数不包含 gin.Context

---

- [ ] T3. 分析现有请求模型，确认是否需要新增

  **What to do**:
  - 读取 `backend/models/requests.go` 了解现有请求模型
  - 检查 Handler 中的 ShouldBindJSON/ShouldBind 调用
  - 确认是否需要新增请求模型 (如 Service 专用 DTO)

  **Must NOT do**:
  - 不修改现有请求模型结构

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: 分析现有代码结构，决策是否需要新增
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with T1, T2, T4)
  - **Blocks**: None
  - **Blocked By**: None

  **References**:
  - `backend/models/requests.go`
  - `backend/controllers/controller.go` - ShouldBind 调用

  **QA Scenarios**:
  - [ ] 确认现有请求模型是否满足 Service 层需求

---

- [ ] T4. 检查 models/models.go 是否有需要调整

  **What to do**:
  - 读取 `backend/models/models.go`
  - 检查是否需要添加 Service 层专用的模型字段或方法
  - 确认现有模型是否满足 Repository 层需求

  **Must NOT do**:
  - 不修改现有字段定义
  - 不修改 JSON 标签

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: 检查现有模型是否完整
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with T1, T2, T3)
  - **Blocks**: T5-T8
  - **Blocked By**: None

  **QA Scenarios**:
  - [ ] 确认 models 包可正常导入

---

- [ ] T5. 实现 GormBookRepository

  **What to do**:
  - 在 `backend/repositories/book_repo.go` 创建文件
  - 实现 `BookRepository` 接口的所有方法
  - 方法签名使用 `*gorm.DB` 作为第一个参数 (支持事务)
  - 返回 `*models.Book` 或 `[]*models.Book`

  **Must NOT do**:
  - 不在 Repository 层处理业务逻辑 (如库存检查)
  - 只做纯数据操作

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: 需要实现完整的 GORM 操作，包含复杂查询
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with T6, T7, T8)
  - **Blocks**: T9-T12
  - **Blocked By**: T1

  **References**:
  - `backend/controllers/controller.go` - 现有 Book 的 DB 操作
  - `backend/models/models.go` - Book 模型定义

  **Acceptance Criteria**:
  - [ ] 实现 Create, GetByID, Update, Delete, FindAll, FindByTitle, FindByAuthor 方法
  - [ ] `go build ./repositories` 成功

  **QA Scenarios**:
  - [ ] Scenario: 验证 GormBookRepository 实现了 BookRepository 接口
    Tool: Bash
    Steps: `cd backend && go build ./repositories`
    Expected Result: 编译成功，无错误
    Evidence: N/A (编译验证)

---

- [ ] T6. 实现 GormUserRepository

  **What to do**:
  - 在 `backend/repositories/user_repo.go` 创建文件
  - 实现 `UserRepository` 接口
  - 方法签名使用 `*gorm.DB` 作为第一个参数

  **Must NOT do**:
  - 不在 Repository 层处理密码验证

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: 实现用户数据持久化
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with T5, T7, T8)
  - **Blocks**: T9-T12
  - **Blocked By**: T1

  **References**:
  - `backend/controllers/controller.go` - Register, Login 中的 DB 操作

  **QA Scenarios**:
  - [ ] `go build ./repositories` 成功

---

- [ ] T7. 实现 GormBorrowRepository

  **What to do**:
  - 在 `backend/repositories/borrow_repo.go` 创建文件
  - 实现 `BorrowRepository` 接口
  - 方法签名使用 `*gorm.DB` 作为第一个参数
  - 支持 FOR UPDATE 锁查询

  **Must NOT do**:
  - 不在 Repository 层处理库存更新

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: 实现借阅记录持久化，包含复杂的事务查询
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with T5, T6, T8)
  - **Blocks**: T9-T12
  - **Blocked By**: T1

  **References**:
  - `backend/controllers/controller.go` - BorrowBook, ReturnBook 中的 DB 操作

  **QA Scenarios**:
  - [ ] `go build ./repositories` 成功

---

- [ ] T8. 创建 repositories/repositories.go 统一导出

  **What to do**:
  - 在 `backend/repositories/repositories.go` 创建文件
  - 导出所有 Repository 构造函数
  - `NewGormBookRepository(db *gorm.DB) BookRepository`
  - `NewGormUserRepository(db *gorm.DB) UserRepository`
  - `NewGormBorrowRepository(db *gorm.DB) BorrowRepository`

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: 简单的导出函数
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with T5, T6, T7)
  - **Blocks**: T9-T12, T17
  - **Blocked By**: T1

  **QA Scenarios**:
  - [ ] `go build ./repositories` 成功

---

- [ ] T9. 实现 BookServiceImpl

  **What to do**:
  - 在 `backend/services/book_service.go` 创建文件
  - 实现 `BookService` 接口
  - 依赖 `BookRepository` 接口
  - 包含 CreateBook (检查重复), GetBooks (条件查询), UpdateBook (事务), DeleteBook (检查库存)

  **Must NOT do**:
  - 不处理 HTTP 请求/响应
  - 不依赖 gin.Context

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: 封装图书业务逻辑，需要处理库存检查和事务
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with T10, T11, T12)
  - **Blocks**: T13-T16
  - **Blocked By**: T2, T5-T8

  **References**:
  - `backend/controllers/controller.go` - CreateBook, GetBooks, UpdateBook, DeleteBooks 业务逻辑

  **QA Scenarios**:
  - [ ] `go build ./services` 成功

---

- [ ] T10. 实现 UserServiceImpl

  **What to do**:
  - 在 `backend/services/user_service.go` 创建文件
  - 实现 `UserService` 接口
  - 依赖 `UserRepository` 接口
  - 包含 Register (密码加密), Login (密码验证), GetUserByUsername

  **Must NOT do**:
  - 不处理 Session (Session 是 Handler 职责)

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: 封装用户业务逻辑，包含密码处理
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with T9, T11, T12)
  - **Blocks**: T13-T16
  - **Blocked By**: T2, T5-T8

  **References**:
  - `backend/controllers/controller.go` - Register, Login 业务逻辑
  - `backend/utils/utils.go` - HashPassword, ComparePassword

  **QA Scenarios**:
  - [ ] `go build ./services` 成功

---

- [ ] T11. 实现 BorrowServiceImpl

  **What to do**:
  - 在 `backend/services/borrow_service.go` 创建文件
  - 实现 `BorrowService` 接口
  - 依赖 `BorrowRepository`, `BookRepository` 接口
  - 包含 BorrowBook (库存检查、事务), ReturnBook (库存更新、事务), GetUserRecords, GetAllRecords
  - 使用 SELECT ... FOR UPDATE 处理并发

  **Must NOT do**:
  - 不处理 HTTP 请求/响应

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: 封装借阅业务逻辑，包含复杂的库存和事务管理
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with T9, T10, T12)
  - **Blocks**: T13-T16
  - **Blocked By**: T2, T5-T8

  **References**:
  - `backend/controllers/controller.go` - BorrowBook, ReturnBook 业务逻辑

  **QA Scenarios**:
  - [ ] `go build ./services` 成功

---

- [ ] T12. 创建 services/services.go 统一导出

  **What to do**:
  - 在 `backend/services/services.go` 创建文件
  - 导出所有 Service 构造函数
  - `NewBookService(repo BookRepository) BookService`
  - `NewUserService(repo UserRepository) UserService`
  - `NewBorrowService(borrowRepo BorrowRepository, bookRepo BookRepository) BorrowService`

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: 简单的导出函数
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with T9, T10, T11)
  - **Blocks**: T13-T16
  - **Blocked By**: T2, T5-T8

  **QA Scenarios**:
  - [ ] `go build ./services` 成功

---

- [ ] T13. 修改 BookHandler 依赖注入

  **What to do**:
  - 修改 `backend/controllers/controller.go`
  - 创建 `BookHandler` 结构体，包含 `BookService` 和 `BookRepository` 字段
  - 添加构造函数 `NewBookHandler(bookService BookService, bookRepo BookRepository) *BookHandler`
  - 修改 CreateBook, GetBooks, UpdateBook, DeleteBooks 函数为 BookHandler 方法
  - 文件上传逻辑 (utils.SaveImages) 保留在 Handler 层

  **Must NOT do**:
  - 不改变 API 响应格式
  - 不改变 HTTP 状态码
  - 不改变错误消息

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: 修改现有 Handler 适配依赖注入，需要小心保持 API 兼容
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4 (with T14, T15, T16, T17)
  - **Blocks**: T18-T22
  - **Blocked By**: T9-T12

  **References**:
  - `backend/controllers/controller.go` - 现有 Book Handler 实现
  - `backend/services/interfaces.go` - Service 接口定义
  - `backend/repositories/interfaces.go` - Repository 接口定义

  **QA Scenarios**:
  - [ ] `go build ./controllers` 成功
  - [ ] API 端点行为不变

---

- [ ] T14. 修改 UserHandler 依赖注入

  **What to do**:
  - 修改 `backend/controllers/controller.go`
  - 创建 `UserHandler` 结构体，包含 `UserService` 字段
  - 添加构造函数 `NewUserHandler(userService UserService) *UserHandler`
  - 修改 Register, Login, Logout 函数为 UserHandler 方法
  - Session 管理保留在 Handler 层 (使用 sessions.Default(c))

  **Must NOT do**:
  - 不改变 API 响应格式
  - 不改变 Session 相关行为

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: 修改现有 Handler 适配依赖注入
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4 (with T13, T15, T16, T17)
  - **Blocks**: T18-T22
  - **Blocked By**: T9-T12

  **QA Scenarios**:
  - [ ] `go build ./controllers` 成功

---

- [ ] T15. 修改 BorrowHandler 依赖注入

  **What to do**:
  - 修改 `backend/controllers/controller.go`
  - 创建 `BorrowHandler` 结构体，包含 `BorrowService`, `BookRepository` 字段
  - 添加构造函数 `NewBorrowHandler(borrowService BorrowService) *BorrowHandler`
  - 修改 BorrowBook, ReturnBook, BorrowRecords, GetAllBorrowRecords, BorrowRecordsByID 为方法

  **Must NOT do**:
  - 不改变 API 响应格式

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: 修改现有 Handler 适配依赖注入
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4 (with T13, T14, T16, T17)
  - **Blocks**: T18-T22
  - **Blocked By**: T9-T12

  **QA Scenarios**:
  - [ ] `go build ./controllers` 成功

---

- [ ] T16. 重构 controllers/controller.go 整体结构

  **What to do**:
  - 将所有全局函数迁移到对应的 Handler 结构体方法
  - 移除对 `config.DB` 的直接引用
  - 确保所有 Handler 方法签名正确
  - 清理未使用的导入

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: 整理代码结构
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4 (with T13, T14, T15, T17)
  - **Blocks**: T18-T22
  - **Blocked By**: T9-T12

  **QA Scenarios**:
  - [ ] `go build ./controllers` 成功
  - [ ] `go vet ./controllers` 无警告

---

- [ ] T17. 更新 main.go 依赖注入 Wiring

  **What to do**:
  - 修改 `backend/main.go`
  - 导入 repositories 和 services 包
  - 创建 Repository 实例: `bookRepo := repositories.NewGormBookRepository(config.DB)`
  - 创建 Service 实例: `bookService := services.NewBookService(bookRepo)`
  - 创建 Handler 实例: `bookHandler := controllers.NewBookHandler(bookService, bookRepo)`
  - 修改 routes 注册使用 handler 实例方法

  **Must NOT do**:
  - 不改变路由路径
  - 不改变中间件逻辑

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: 依赖注入 wiring，需要理解整体架构
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4 (with T13, T14, T15, T16)
  - **Blocks**: T18-T22
  - **Blocked By**: T9-T12

  **References**:
  - `backend/routes/book.go` - 路由注册
  - `backend/routes/user.go` - 路由注册

  **QA Scenarios**:
  - [ ] `go build ./...` 成功

---

- [ ] T18. 创建 Handler 单元测试示例

  **What to do**:
  - 在 `backend/controllers/book_handler_test.go` 创建测试文件
  - 使用 gomock 生成 mock
  - 使用 httptest 模拟 HTTP 请求
  - 测试 GetBooks 端点: 参数校验、查询逻辑、响应格式

  **Must NOT do**:
  - 不使用真实数据库

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: 编写单元测试示例，需要熟悉 gomock 和 httptest
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 5 (with T19, T20, T21, T22)
  - **Blocked By**: T13-T17

  **QA Scenarios**:
  - [ ] `go test ./controllers -v -run TestGetBooks` 通过

---

- [ ] T19. 创建 Repository 集成测试示例

  **What to do**:
  - 在 `backend/repositories/integration_test.go` 创建测试文件 (或 separate file)
  - 使用 testcontainers-go 启动 MySQL 容器
  - 测试 GormBookRepository 的实际数据库操作
  - 包含 Create, GetByID, Delete 测试用例

  **Must NOT do**:
  - 不使用 SQLite 内存数据库

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: 编写集成测试，需要熟悉 testcontainers-go
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 5 (with T18, T20, T21, T22)
  - **Blocked By**: T13-T17

  **QA Scenarios**:
  - [ ] `go test ./repositories -v -run TestIntegration` 通过 (需要 Docker)

---

- [ ] T20. 运行 go generate 验证 mock 生成

  **What to do**:
  - 在 backend 目录运行 `go generate ./...`
  - 验证 `mocks/mock_*.go` 文件生成成功

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: 验证工具配置
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 5 (with T18, T19, T21, T22)
  - **Blocked By**: T13-T17

  **QA Scenarios**:
  - [ ] `go generate ./...` 成功
  - [ ] `backend/mocks/` 目录包含生成的 mock 文件

---

- [ ] T21. 运行 go build 验证编译

  **What to do**:
  - 运行 `go build -o /dev/null ./...` 验证整个项目编译
  - 确保没有编译错误

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: 验证编译
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 5 (with T18, T19, T20, T22)
  - **Blocked By**: T13-T17

  **QA Scenarios**:
  - [ ] `go build ./...` 成功

---

- [ ] T22. 运行 go test 验证测试通过

  **What to do**:
  - 运行 `go test ./...` 验证测试通过
  - 包括单元测试和集成测试

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: 验证测试
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 5 (with T18, T19, T20, T21)
  - **Blocked By**: T13-T17

  **QA Scenarios**:
  - [ ] `go test ./...` 通过

---

## Final Verification Wave

- [ ] F1. **API 合约审计** — 验证所有端点行为不变
  读取 plan，对比实际 API 响应格式、状态码、错误消息

- [ ] F2. **代码质量审查** — `go vet`, `go fmt`
  运行代码质量检查工具

- [ ] F3. **编译验证** — 完整编译
  `go build ./...`

- [ ] F4. **依赖注入验证** — 确保无全局 DB 引用在 Handler
  检查 controllers/ 不再直接使用 config.DB

---

## Commit Strategy

- Wave 1: `feat(clean-arch): add repository interfaces`
- Wave 2: `feat(clean-arch): implement gorm repositories`
- Wave 3: `feat(clean-arch): implement services`
- Wave 4: `refactor(controllers): add dependency injection`
- Wave 5: `test: add unit and integration tests`

---

## Success Criteria

### 验证命令
```bash
cd backend
go build ./...           # 编译成功
go generate ./...       # 生成 mock 文件
go test ./...           # 测试通过
go vet ./...            # 无警告
```

### 最终检查清单
- [ ] 所有 Handler 不再直接依赖 `config.DB`
- [ ] 所有 Handler 依赖接口 (Repository/Service)
- [ ] go:generate 注释在所有接口文件
- [ ] Mock 文件可正常生成
- [ ] API 行为完全不变
