# OpenToken (中文文档)

OpenToken 是一个高性能、可扩展的 LLM（大语言模型）网关与分发系统。它采用 Server-Node 架构，旨在为开发者提供一个统一、灵活且安全的 API 接入点。

## 项目架构

- **opentoken-server**: 中心控制节点，负责 API 路由、用户鉴权、节点管理、额度控制以及前端展示。
- **opentoken-node**: 执行节点，负责与实际的 LLM 上游（如 OpenAI, Anthropic, Ollama 等）进行通信并流式回传结果。
- **frontend**: Web 前端项目，基于 React + Vite + TypeScript 构建，提供系统管理和监控的用户界面。

## 核心特性

- **架构解耦**: Server 与 Node 分离，支持跨地域、跨云部署。
- **高性能**: 基于 Go 语言开发，原生支持高并发长连接。
- **流式回传**: 全链路流式响应支持。
- **灵活配置**: 支持 MySQL/SQLite 存储，支持多种 OSS。

---

## 界面预览

### 节点管理

![在线节点](img/在线节点.png)

![节点管理](img/节点管理.png)

### 模型管理

![可用模型列表](img/可用模型列表.png)

### API 管理

![API Key 管理](img/api%20key%20管理.png)

![API使用文档](img/API使用文档.png)

---

## 本地部署指南

### 前置要求

- **Go**: 建议版本 1.25.1 或更高。
- **数据库**: 可选 MySQL 或内置 SQLite。

### 1. 配置文件位置

- Server 配置：`opentoken-server/config.yaml`
- Node 配置：`opentoken-node/config.yaml`

### 2. Server 配置与启动

#### 2.1 修改配置 (`opentoken-server/config.yaml`)

编辑配置文件，确保数据库连接正确：

```yaml
system:
  port: 8084
  db-type: sqlite # 或者使用 mysql
sqlite:
  db-path: "./opentoken-server.db"
```

#### 2.2 启动服务

进入 server 目录并运行：

```bash
cd opentoken-server
go run main.go
```

### 3. Node 配置与启动

#### 3.1 获取 Node Token

Node 需要凭证才能连接到 Server。在 Server 启动后，通过接口获取：

```bash
curl -X POST http://127.0.0.1:8084/api/v1/node/credentials \
  -H "Content-Type: application/json" \
  -d '{"name":"node-local-1"}'
```

记录响应中的 `token`。

#### 3.2 配置 Node (`opentoken-node/config.yaml`)

编辑 Node 配置文件：

```yaml
node-agent:
    enabled: true
    server-ws-url: ws://127.0.0.1:8084/ws/node
    token: "你的TOKEN"
    node-name: "node-local-1"
    models: ["gpt-4o-mini"]

upstream-llm:
    base-url: "http://localhost:11434/v1"
    api-key: "你的API_KEY"
```

#### 3.3 启动节点

进入 node 目录并运行：

```bash
cd opentoken-node
go run main.go
```

### 4. 前端配置与启动

前端是基于 React + Vite + TypeScript 构建的 Web 应用。

#### 4.1 前置要求

- **Node.js**: 建议版本 18.x 或更高。
- **npm** 或 **yarn**: 包管理器。

#### 4.2 安装依赖

进入前端目录并安装依赖：

```bash
cd frontend
npm install
```

#### 4.3 开发模式

启动开发服务器：

```bash
npm run dev
```

前端将在 `http://localhost:3000` 上运行。开发服务器会自动将 API 请求代理到后端服务器 `http://localhost:8084`。

#### 4.4 生产环境构建

构建生产环境前端：

```bash
npm run build
```

构建产物将生成在 `dist` 目录中。可以使用以下命令预览生产构建：

```bash
npm run preview
```

#### 4.5 测试环境

用于测试目的，可以使用：

```bash
npm run test:dev    # 以测试模式启动开发服务器
npm run test:build   # 构建测试环境
```

## 节点管理与监控

### 1. 查询在线节点

该接口用于查询当前所有已成功连接并处于活跃状态的节点及其信息（包括节点名称、支持的模型列表等）。

- **请求方式**: `GET`
- **接口地址**: `/api/v1/node/online`
- **示例**:

```bash
curl http://127.0.0.1:8084/api/v1/node/online
```

**响应说明**:
返回一个 JSON 数组，包含所有在线节点的详细信息。如果 `models` 列表中没有包含你想要请求的模型，说明该节点配置有误或未成功加载该模型。

---

## 最小联调验证

### 2. 通过 Server 调用模型

使用标准的 Chat Completion 格式请求 Server：

```bash
curl -N http://127.0.0.1:8084/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model":"qwen3.5:27b",
    "stream":true,
    "messages":[{"role":"user","content":"你好"}]
  }'
```

---

## 常见问题 (FAQ)

- **Node 无法连接**: 检查 `server-ws-url` 是否正确，以及网络防火墙。
- **模型路由失败**: 确保 `node-agent.models` 中声明了请求的模型名。
- **Token 失效**: 若 Token 丢失或泄露，请重新生成 Credential 并更新 Node 配置。

