# 重构总结（分层解耦版）

本文仅记录当前项目在分层解耦上的结果，聚焦依赖方向、职责边界与组装方式。

## 1. 总体依赖方向

当前依赖方向已统一为：

`Controller -> Service 接口 -> Repository 接口 -> GORM 实现`

关键原则：
- 上层只依赖抽象（interface），不依赖具体 ORM 或全局 DB。
- 事务边界由 `Transactor` 统一管理，业务层不直接操作 `*gorm.DB`。
- 数据读写下沉到 Repository；业务规则留在 Service。

---

## 2. Controller 层解耦

现状：
- Controller 通过构造注入 Service，不直接使用 `config.DB`。
- Controller 仅处理 HTTP 协议职责：参数绑定、鉴权上下文、响应编码。
- 业务分支（库存校验、借还规则、存在性判断）不在 Controller 实现。

结果：
- Web 框架与业务逻辑分离，Controller 可替换性提升。
- Handler 测试可通过 mock service 独立进行。

---

## 3. Service 层解耦

现状：
- `BookService`、`BorrowService`、`UserService` 均采用仓储接口注入。
- Service 已移除对 `config.DB` 与 `*gorm.DB` 的直接依赖。
- Service 不再依赖 `gorm.ErrRecordNotFound`；改为仓储返回 `nil` 语义再映射为领域错误。

职责边界：
- Service 负责业务规则：
  - 图书唯一性与库存语义校验
  - 借书/还书状态流转
  - 领域错误转换
- Service 不负责 SQL/ORM 细节。

结果：
- 业务逻辑可在不启动数据库的情况下进行单元测试。
- ORM 更换对 Service 的影响被最小化。

---

## 4. Repository 层解耦

现状：
- `BookRepository` / `BorrowRepository` / `UserRepository` 均为 ORM 无关接口。
- 接口方法签名不暴露 `*gorm.DB`。
- GORM 细节收敛在 `gorm*Repository` 实现中。
- 构造函数统一为：
  - `NewGormBookRepository(db *gorm.DB) BookRepository`
  - `NewGormBorrowRepository(db *gorm.DB) BorrowRepository`
  - `NewGormUserRepository(db *gorm.DB) UserRepository`

补充：
- 已删除历史聚合工厂文件，避免旧命名/旧签名造成编译与 IDE 误报。
- repository mocks 已按新接口重生成。

结果：
- Repository 接口成为稳定边界；ORM 变化只影响实现层。

---

## 5. 事务层（Transactor）解耦

现状：
- 引入 `Transactor` + `TxRepositories` 抽象：
  - `WithinTransaction(ctx, fn)` 统一提交/回滚语义
  - 事务内通过 `repos.Books()` / `repos.Borrows()` 获取同一事务上下文下的仓储
- Borrow 相关复合操作（锁行、扣减/回补库存、更新借阅记录）通过该抽象执行。

结果：
- Service 持有“事务能力”而非 ORM 对象。
- 事务一致性与业务流程绑定，但不泄露底层实现。

---

## 6. 依赖组装（Composition Root）

现状：
- 在 `main.go` 完成一次性组装：
  - DB 初始化
  - GORM Repository 实例化
  - Service 注入 Repository/Transactor
  - Controller 注入 Service

结果：
- 依赖创建集中，运行时关系清晰。
- 替换实现（如内存仓储、其他 ORM）只需调整组装层。

---

## 7. 当前解耦结论

结论：
- Repository 与 Service 已完成接口级解耦。
- Service 与 ORM / 全局 DB 已解耦。
- Controller 与数据访问细节已解耦。
- 事务管理已抽象为独立能力并接入业务流程。

整体上，项目已形成“分层清晰、依赖单向、组装集中”的结构，可在不破坏上层代码的前提下演进底层持久化实现。
