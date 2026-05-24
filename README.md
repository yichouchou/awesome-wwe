# Awesome WWE - WWE 剧本智能体

> 基于 Go + LangChain + MiniMax-M2.7 的 WWE 摔角剧本创作智能体

[English](./README.md) | 中文

---

## 项目简介

Awesome WWE 是一个利用 AI 技术辅助创作 WWE 摔角剧本的智能体项目。基于 Go 语言开发，使用 MiniMax-M2.7 作为 LLM，支持完整的角色管理、剧情线追踪、实体关系存储等功能。

## 技术栈

- **语言**: Go 1.22+
- **LLM**: MiniMax-M2.7 (通过 MiniMax API)
- **数据库**: SQLite (go-sqlite3)
- **架构**: 简洁的 Agent 设计，支持 LangChain 风格扩展

## 核心功能

### 🎬 剧情生成
- 根据角色设定和摔角风格自动生成剧本
- 支持 Raw、SmackDown、NXT 等不同节目风格
- 为 WrestleMania、Summerslam、Royal Rumble 等 PPV 设计高光剧情

### 👥 角色管理
- WWE 摔角手角色数据库
- 别名多映射支持
- 角色关系追踪（盟友、对手、恩怨）
- 角色出场历史记录

### 📖 故事线构建
- 多线程叙事支持
- 伏笔与回收系统
- 剧情冲突管理
- PPV 剧情走向规划

## 项目结构

```
awesome-wwe/
├── cmd/
│   └── wwe-agent/           # 主程序入口
│       └── main.go
├── internal/
│   ├── agent/               # AI 智能体
│   │   └── agent.go
│   ├── index/              # 实体/关系管理
│   │   └── index.go
│   └── llm/                # LLM 客户端
│       └── client.go
├── scripts/                 # 工具脚本
├── docs/                    # 文档
└── README.md
```

## 快速开始

### 安装依赖

```bash
# 克隆项目
git clone https://github.com/yichouchou/awesome-wwe.git
cd awesome-wwe

# 安装 Go 依赖
go mod tidy

# 初始化数据库
go run cmd/wwe-agent/main.go --init
```

### 设置环境变量

```bash
# MiniMax API Key (必需)
export MINIMAX_API_KEY="your-api-key"

# 或使用 Claude Code 的配置
export ANTHROPIC_AUTH_TOKEN="your-anthropic-token"

# 可选：自定义 API 地址和模型
export LLM_BASE_URL="https://api.minimaxi.com/anthropic"
export LLM_MODEL_NAME="MiniMax-M2.7"
```

### 使用示例

```bash
# 初始化数据库
go run cmd/wwe-agent/main.go --init

# 添加角色
go run cmd/wwe-agent/main.go --add-entity "stone_cold:角色:Stone Cold Steve Austin:核心:WWE标志性人物，傲慢粗犷"

go run cmd/wwe-agent/main.go --add-entity "the_rock:角色:The Rock:核心:WWE黄金时代代表，魅力四射"

# 添加对手关系
go run cmd/wwe-agent/main.go --add-rel "stone_cold:the_rock:对手:WWE黄金时代经典对决:50"

# 查询角色
go run cmd/wwe-agent/main.go --query "Austin"

# 列出所有角色
go run cmd/wwe-agent/main.go --list

# 查看统计
go run cmd/wwe-agent/main.go --stats

# 生成剧本
go run cmd/wwe-agent/main.go --generate "Raw第100期 Stone Cold vs The Rock 冠军争夺战"
```

## 数据库结构

### entities (实体表)

| 字段 | 类型 | 说明 |
|-----|------|------|
| id | TEXT | 唯一标识 (如 stone_cold) |
| type | TEXT | 角色/势力/地点/物品/招式 |
| canonical_name | TEXT | 标准名称 (如 "Stone Cold Steve Austin") |
| tier | TEXT | 核心/重要/次要/装饰 |
| desc | TEXT | 描述 |

### relationships (关系表)

| 字段 | 类型 | 说明 |
|-----|------|------|
| from_entity | TEXT | 起始实体 |
| to_entity | TEXT | 目标实体 |
| type | TEXT | 关系类型 (对手/盟友/同伴等) |
| description | TEXT | 关系描述 |
| chapter | INTEGER | 出场章节 |

### aliases (别名表)

支持一个角色多个别名 (如 "Austin" -> stone_cold)

## 开发

### 编译

```bash
# 编译到当前目录
go build -o wwe-agent ./cmd/wwe-agent

# 交叉编译 (Linux ARM64)
GOOS=linux GOARCH=arm64 go build -o wwe-agent-arm64 ./cmd/wwe-agent
```

### 测试

```bash
go test ./...
```

## 相关项目

- [Webnovel-Writer](https://github.com/lingfengQAQ/webnovel-writer) - 网文写作智能体

## License

MIT License - See [LICENSE](LICENSE) for details.