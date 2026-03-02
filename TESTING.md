# Red-TLDR 试用测试指南

## 1. 编译

```bash
cd /home/arch/Git/red-tldr
go build -o red-tldr .
```

## 2. 基础 CLI 测试

```bash
# 查看帮助
./red-tldr --help

# 从 GitHub 下载数据库（首次使用必须）
./red-tldr upgrade

# 关键词搜索
./red-tldr mimikatz

# JSON 格式输出
./red-tldr -f json mimikatz

# Markdown 格式输出
./red-tldr -f markdown mimikatz
```

## 3. Bleve 全文搜索测试

```bash
# 构建 Bleve 索引（同时也会重建 JSON 索引）
./red-tldr update

# 以下搜索将自动走 Bleve 引擎（比旧版 tag 匹配能力强得多）

# 搜索 tag 中的关键词（和以前一样能工作）
./red-tldr nmap

# 搜索 data 内容中的命令片段（旧版做不到，现在可以）
./red-tldr wmiexec

# 模糊拼写（拼错也能匹配）
./red-tldr mimikat
```

### 验证 Bleve 索引位置

```bash
ls -la ~/.red-tldr/bleve.index/
```

如果该目录存在，说明 Bleve 索引已构建成功。

### 对比 Bleve vs 旧搜索

```bash
# 删掉 Bleve 索引后搜索 = fallback 到旧的 tag 匹配
rm -rf ~/.red-tldr/bleve.index
./red-tldr wmiexec        # 旧引擎：大概率搜不到（因为 wmiexec 不在 tag 里）

# 重建后再搜
./red-tldr update
./red-tldr wmiexec        # Bleve 引擎：能搜到（因为搜了 data 字段内容）
```

## 4. MCP Server 测试

### 4.1 手动 stdio 测试

MCP 协议走 stdin/stdout JSON-RPC，可以手工发消息测试：

```bash
# 启动 MCP Server（会阻塞等待 stdin 输入）
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}' | ./red-tldr serve
```

如果看到类似以下 JSON 输出，说明 MCP Server 工作正常：

```json
{"jsonrpc":"2.0","id":1,"result":{"protocolVersion":"...","capabilities":{"tools":{}},...}}
```

### 4.2 配置到 Claude Desktop / Cursor

在 AI IDE 的 MCP 配置文件中添加：

```json
{
  "mcpServers": {
    "red-tldr": {
      "command": "/home/arch/Git/red-tldr/red-tldr",
      "args": ["serve"]
    }
  }
}
```

配置后重启 IDE，AI 助手就能调用以下三个工具：

| Tool 名称 | 功能 | 示例调用 |
|-----------|------|----------|
| `search_redteam_commands` | 搜索红队命令 | `{"query": "mimikatz", "platform": "windows"}` |
| `get_command_details` | 获取条目详情 | `{"file": "files/Active_directory/mimikatz.yaml"}` |
| `list_techniques` | 列出所有技术 | `{"keyword": "scan"}` |

### 4.3 用 OpenCode 测试（当前环境）

如果你在 OpenCode 中配置了上面的 MCP，可以直接对 AI 说：

> 帮我查一下 mimikatz 的用法

AI 会自动调用 `search_redteam_commands` → `get_command_details`，返回经过验证的命令。

## 5. 运行测试套件

```bash
# 运行全部 35 个测试
go test -v ./...

# 带竞态检测
go test -v -race ./...

# 只跑 Bleve 搜索测试
go test -v ./internal/search/

# 只跑 MCP 参数解析测试
go test -v ./internal/mcp/
```

## 6. 项目结构速览

```
cmd/
  root.go          CLI 入口（搜索自动选择 Bleve 或 fallback）
  serve.go         MCP Server 入口

internal/
  config/          配置管理
  db/              数据模型 + JSON 索引 + 基础搜索
  search/          Bleve 全文搜索引擎
  mcp/             MCP Server（3 个 Tool）
  render/          多格式输出（text/json/markdown）
  updater/         数据库下载（含 zip-slip 防护）
```

## 7. 常见问题

**Q: `red-tldr upgrade` 下载失败？**
A: 需要能访问 GitHub API。如果网络受限，可以手动 clone `Rvn0xsy/red-tldr-db` 到 `~/red-tldr-db/`。

**Q: 搜索结果和以前不一样？**
A: `update` 之后会使用 Bleve 搜索引擎，按相关性排序（name 权重最高，tag 次之，内容最低）。删掉 `~/.red-tldr/bleve.index/` 可以回退到旧的 tag 匹配。

**Q: MCP Server 怎么停？**
A: Ctrl+C 或关闭 stdin。MCP Server 是无状态的 stdio 服务，没有后台进程。
