# 定时任务服务 (Cron Task Service)

这是一个基于 go-zero 框架构建的通用定时任务服务，支持灵活的任务注册和管理。

## 功能特性

- 🕒 支持标准 cron 表达式和秒级精度
- 🔧 通用的任务接口，易于扩展
- 📊 任务执行结果记录和查询
- ⚡ 并发控制和超时管理
- 🛡️ 错误处理和日志记录
- 🔄 动态任务注册和取消

## 目录结构

```
apps/task/cron/
├── etc/                    # 配置文件
│   └── cron-local.yaml
├── internal/               # 内部实现
│   ├── config/            # 配置结构
│   ├── handler/           # 处理器
│   ├── logic/             # 业务逻辑
│   ├── svc/               # 服务上下文
│   └── types/             # 类型定义
├── tasks/                 # 任务实现
│   ├── example_task.go    # 示例任务
│   ├── data_cleanup_task.go # 数据清理任务
│   ├── stats_task.go      # 统计任务
│   └── registry.go        # 任务注册器
├── cron.go                # 主入口文件
└── README.md              # 说明文档
```

## 如何添加新任务

### 1. 创建任务实现

在 `tasks/` 目录下创建新的任务文件，实现 `types.Task` 接口：

```go
package tasks

import (
    "context"
    "time"
    "github.com/iceymoss/go-hichat-api/apps/task/cron/internal/svc"
)

type YourCustomTask struct {
    svc *svc.ServiceContext
}

func NewYourCustomTask(svc *svc.ServiceContext) *YourCustomTask {
    return &YourCustomTask{svc: svc}
}

func (t *YourCustomTask) GetName() string {
    return "your_custom_task"
}

func (t *YourCustomTask) GetSpec() string {
    return "0 0 */2 * * *" // 每2小时执行一次
}

func (t *YourCustomTask) GetDescription() string {
    return "你的自定义任务描述"
}

func (t *YourCustomTask) GetTimeout() time.Duration {
    return 5 * time.Minute
}

func (t *YourCustomTask) Execute(ctx context.Context) error {
    // 在这里实现你的业务逻辑
    return nil
}
```

### 2. 注册任务

在 `tasks/registry.go` 文件中注册新任务：

```go
func RegisterAllTasks(taskManager types.TaskManager, svc *svc.ServiceContext) {
    registry := NewTaskRegistry()

    // 现有任务...
    registry.RegisterTask(NewExampleTask(svc))
    registry.RegisterTask(NewDataCleanupTask(svc))
    registry.RegisterTask(NewStatsTask(svc))
    
    // 添加你的新任务
    registry.RegisterTask(NewYourCustomTask(svc))

    // 注册所有任务
    for _, task := range registry.GetAllTasks() {
        if err := taskManager.RegisterTask(task); err != nil {
            panic(err)
        }
    }
}
```

## Cron 表达式说明

本服务支持标准的 cron 表达式，并可选支持秒级精度：

### 标准格式 (5位)
```
* * * * *
│ │ │ │ │
│ │ │ │ └── 星期几 (0-7, 0和7都表示星期日)
│ │ │ └──── 月份 (1-12)
│ │ └────── 日期 (1-31)
│ └──────── 小时 (0-23)
└────────── 分钟 (0-59)
```

### 秒级精度格式 (6位)
```
* * * * * *
│ │ │ │ │ │
│ │ │ │ │ └── 星期几 (0-7, 0和7都表示星期日)
│ │ │ │ └──── 月份 (1-12)
│ │ │ └────── 日期 (1-31)
│ │ └──────── 小时 (0-23)
│ └────────── 分钟 (0-59)
└──────────── 秒 (0-59)
```

### 常用示例

- `0 * * * * *` - 每分钟执行
- `0 0 * * * *` - 每小时执行
- `0 0 0 * * *` - 每天午夜执行
- `0 0 0 * * 0` - 每周日午夜执行
- `0 0 0 1 * *` - 每月1号午夜执行

## 配置说明

在 `etc/cron-local.yaml` 中配置服务参数：

```yaml
Cron:
  WithSeconds: true        # 是否启用秒级精度
  MaxConcurrency: 10      # 任务并发限制
  TaskTimeout: 300        # 任务超时时间(秒)
```

## 运行服务

```bash
# 开发环境运行
go run apps/task/cron/cron.go -f apps/task/cron/etc/cron-local.yaml

# 或者编译后运行
go build -o cron-task apps/task/cron/cron.go
./cron-task -f apps/task/cron/etc/cron-local.yaml
```

## 任务管理

### 查看任务状态
可以通过日志查看任务的执行状态和结果。

### 任务执行结果
每个任务的执行结果都会被记录，包括：
- 任务名称
- 开始时间
- 结束时间
- 执行时长
- 执行状态（成功/失败）
- 错误信息（如果有）

## 注意事项

1. **任务幂等性**: 确保任务可以安全地重复执行
2. **错误处理**: 在任务中妥善处理错误，避免任务崩溃
3. **资源管理**: 注意数据库连接、文件句柄等资源的正确释放
4. **超时设置**: 根据任务复杂度合理设置超时时间
5. **日志记录**: 在任务中添加适当的日志记录，便于调试和监控

## 扩展功能

未来可以考虑添加以下功能：

- 任务执行历史查询 API
- 任务动态启停控制
- 任务执行统计和监控
- 任务依赖关系管理
- 分布式任务调度
