# CDNFix 架构重构与优化记录

## 1. 核心缺陷修复 (P0)

### 1.1 并发数据竞争修复
**问题**：旧版在 `query` 阶段并发调用腾讯云 API 后，会并发读写 `TaskState` JSON 文件，由于缺乏文件锁，会导致典型的 Lost Update（状态覆写丢失）。
**解决**：在 `workflow/workflow.go` 引入了进程级的 `sync.Mutex` (`stateMutex`)，严格包装了读-改-写的完整流程，确保状态落盘的原子性。

### 1.2 `WaitAndMarkTask` 永久死锁修复
**问题**：旧版 `WaitAndMarkTask` 是一个无限 `for` 循环，若腾讯云由于故障一直返回 `"process"`，进程将永远无法退出。
**解决**：通过引入 `context.Context` 和 `time.NewTicker`，为该操作设定了 30 分钟的严格超时阈值（Timeout）。一旦超时自动抛出异常，不再无限挂起。

### 1.3 废弃大文件污染清理
**问题**：由于 `.gitignore` 不完善，根目录曾遗留 14MB 的 `cndfix` 编译二进制文件。
**解决**：已删除对应垃圾文件，并在 `.gitignore` 中显式忽略了 `cdn` 及 `cndfix`。

## 2. 职责解耦与逻辑下沉 (P1)

### 2.1 上帝函数拆分
**问题**：原有的 `cmd/runtime.go` 充斥了多达 350 行的核心业务逻辑，包隔离性名存实亡。
**解决**：删除了 `cmd/runtime.go`。将其核心功能彻底下沉拆分：
- **`workflow/executor.go`**：负责任务下发、配置解析、URL分拣及记录登记。
- **`workflow/query.go`**：负责并发查询与任务状态轮询。
`cmd/` 包重归 CLI 门面层，仅负责参数解析。

### 2.2 消除重复的样板代码
**问题**：原 `RefreshURLs` 与 `RefreshPaths` 分支包含近 40 行完全重复的状态写入（Copy-Paste）逻辑。
**解决**：抽象为内联闭包函数 `recordTask`，实现了统一的任务记录入库逻辑，精简了核心执行函数。

## 3. 高级性能与体验优化 (P2)

### 3.1 废弃无意义的高频 I/O 刷盘
**问题**：`WaitAndMarkTask` 每 10 秒查询一次 API 后，无论任务是否完成，都会把相同的 `pending` 状态连带更新的 `LastCheckedAt` 全量刷入 JSON 文件。若有 50 个任务，每 10 秒会发生 50 次文件系统写锁争抢。
**解决**：移除了无效状态刷盘，只有在 `completed` 或状态实质变更时才会执行 `MarkTask` 写盘动作。大幅降低了无效的磁盘开销。

### 3.2 串行等待升级为并发限流查询
**问题**：`QueryTaskGroups` 对同一个 Site 的任务 ID 列表使用了死循环式的**串行阻塞等待**。如果有多个任务，查询周期呈线性暴增。
**解决**：重构为全并发 Goroutine 组查询，并使用带容量的 Channel 作为 Semaphore，将腾讯云 API 并发查询请求上限优雅地控制在 10 并发，兼顾了极速并发与 API 限流安全。

### 3.3 Unix 管道 (Stdin) 原生支持
**问题**：只能通过 `--urls` 参数或 `--urlfile` 传文件输入，不支持典型的 Shell 管道组合。
**解决**：在 `workflow.ReadURLs` 中加入了字符设备的识别 `(stat.Mode() & os.ModeCharDevice) == 0`，完美支持 `cat urls.txt | cdnfix push` 式的管道流输入。

## 4. 代码洁癖与构建优化 (P3)

- **Viper 依赖解除**：修改了 `cloud/tencent/clent.go`，剥离了对 `viper` 的全局依赖，改为通过结构体严格参数传递。
- **静默异常修复**：`cmd/refresh.go` 及 `cmd/push.go` 针对用户传空 URL 列表的场景，从之前的静默退出改为了明确抛错拦截。
- **冗余物料清理**：清理了废弃的 `main_test.go` 和带有命名拼写错误的 `cloud/commog.go`。
- **构建标准化**：将 `go.mod` 与 `Dockerfile` 中的 golang 环境版本统一拉平至 `1.21`，消除了 Dockerfile 内冗余的一倍构建行为。
