# Kairos

## 目录结构

```text
kairos/
├── cmd/
│   └── desktop/              # 桌面端入口：main.go、wire 依赖注入、Windows 资源文件
├── internal/
│   ├── config/               # 应用配置加载
│   ├── constant/             # 全局常量
│   ├── ecode/                # 业务错误码定义
│   └── module/               # 业务模块
│       ├── base/             # 公共模块
│       ├── clipboard/        # 剪贴板模块
│       ├── desktop/          # 桌面应用模块（主题、托盘、窗口控制等桌面能力）
│       └── notes/            # 笔记模块
├── pkg/                      # 基础库
│   ├── cache/                # 缓存
│   ├── config/               # 配置解析
│   ├── database/             # 数据库
│   ├── errs/                 # 错误包装
│   ├── id/                   # 雪花 ID
│   └── logger/               # 日志
├── frontend/                 # 前端工程
│   ├── bindings/             # wails3 生成的 Go → TS 绑定（自动生成，勿手改）
│   ├── src/
│   │   ├── components/       # 组件
│   │   ├── composables/      # 组合式函数
│   │   ├── pages/            # 页面
│   │   ├── router/           # 路由
│   │   ├── services/         # 后端方法聚合出口
│   │   ├── stores/           # 全局状态
│   │   ├── styles/           # 全局样式
│   │   ├── theme/            # Naive UI 主题
│   │   ├── main.ts           # 主窗口入口
│   │   └── clipboard.ts      # 剪切板快捷窗口入口
│   ├── index.html            # 主窗口页面
│   └── clipboard.html        # 剪切板快捷窗口页面
├── build/                    # wails3 构建配置、平台脚本与打包资源（windows/nsis）
├── configs/                  # 运行期配置文件
├── logs/                     # 运行日志输出目录
├── assets.go                 # 嵌入文件
├── Taskfile.yml              # 任务入口
└── go.mod
```

## 常用命令

### 安装依赖

```bash
go mod download               # Go 依赖
cd frontend && pnpm install   # 前端依赖
```

### 桌面端任务（项目根目录）

| 命令 | 作用 |
| --- | --- |
| `wails3 task dev` | 开发模式 |
| `wails3 task build` | 构建应用 |
| `wails3 task run` | 运行应用 |
| `wails3 task package` | 打包安装包 |
| `wails3 task release:version VERSION=0.1.1` | 按指定版本发包（`VERSION` 必填） |
| `wails3 task generate:bindings` | 生成绑定 |
| `wails3 task generate:icons` | 生成图标 |

### 前端脚本（`frontend/` 目录）

| 命令 | 作用 |
| --- | --- |
| `pnpm dev` | 启动开发服务器 |
| `pnpm build` | 生产构建 |
| `pnpm build:dev` | 非压缩构建 |
| `pnpm preview` | 预览构建产物 |
| `pnpm check` | 类型检查 + lint |
| `pnpm type-check` | 类型检查 |
| `pnpm lint` / `pnpm lint:fix` | 代码检查 / 自动修复 |

