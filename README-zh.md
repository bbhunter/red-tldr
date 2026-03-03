# Red Team TL;DR

[English](./README.md) | [中文简体](./README-zh.md)

[![GitHub release](https://img.shields.io/github/release/Rvn0xsy/red-tldr.svg)](https://github.com/Rvn0xsy/red-tldr/releases)

## 什么是 Red Team TL;DR？

red-tldr 是一个轻量级的红队命令速查工具，帮助有经验的红队人员快速找到所需的命令、技术和关键要点 —— 就像一个专为攻击性安全打造的 `man` 命令。

从 v0.5.0 开始，red-tldr 内置了 [MCP](https://modelcontextprotocol.io/) 服务器，AI 助手（如 Claude Desktop、Cursor）可以直接查询你的红队知识库。

## 为什么选择 Red Team TL;DR？

在日常红队工作中，需要记忆大量命令，而多数时候你只记得开头几个字符。通过搜索引擎翻找文档既慢又嘈杂。red-tldr 从你自己掌控的策展数据库中给出即时、离线、确定性的答案。

## 特性

- **全文搜索** — 基于 Bleve 的 BM25 加权搜索（name×5、tags×3、data×1），支持模糊匹配
- **多格式输出** — `text`（带颜色的终端输出）、`json`、`markdown`
- **MCP 服务器** — 内置 Model Context Protocol 服务器，支持 AI 助手集成（stdio + Streamable HTTP）
- **MITRE ATT&CK 映射** — 条目可标记战术和技术编号
- **增强数据模型** — category、platforms、mitre_attack、metadata（author/source/confidence）、related
- **向后兼容** — 兼容现有 YAML 数据库，新增字段均为可选
- **离线优先** — 无需云服务、无需 API Key，完全确定性输出
- **安全修复** — 修复了 zip-slip 漏洞，所有解压操作均有路径校验

## 快速开始

### 安装

#### macOS

```bash
$ brew install red-tldr
```

#### Arch Linux

```bash
# AUR: https://aur.archlinux.org/packages/red-tldr
$ sudo pacman -S red-tldr
```

#### 源码编译

```bash
$ git clone https://github.com/Rvn0xsy/red-tldr
$ cd red-tldr
$ go build
```

#### 二进制安装

下载 [Release](https://github.com/Rvn0xsy/red-tldr/releases/) 版本。

```bash
$ tar -zxvf red-tldr_latest_linux_amd64.tar.gz
$ ./red-tldr
```

> 建议将 red-tldr 加入环境变量中使用。

## 使用方法

### 搜索

```bash
# 关键字搜索
$ red-tldr mimikatz
```

![search-mimikatz](./images/img_1.png)

```bash
# 模糊匹配 — 只需输入几个字符
$ red-tldr mi
```

![Fuzzy-match](./images/img_2.png)

当存在多个结果时，输入数字索引选择：

![Select-Number](./images/img_3.png)

### 输出格式

```bash
# JSON 输出
$ red-tldr mimikatz -f json

# Markdown 输出
$ red-tldr mimikatz -f markdown
```

### 更新与升级

```bash
# 重建本地索引（JSON + Bleve）
$ red-tldr update

# 从 GitHub 下载最新数据库
$ red-tldr upgrade
```

## MCP 服务器

red-tldr 内置 MCP 服务器，可以将红队知识库暴露给 AI 助手使用。

### 工具列表

| 工具 | 说明 |
|------|------|
| `search_redteam_commands` | 按关键字、平台、分类、战术或技术搜索命令 |
| `get_command_details` | 获取指定命令条目的完整详情 |
| `list_techniques` | 列出数据库中的 MITRE ATT&CK 技术 |

### stdio 模式（Claude Desktop / Cursor）

```bash
$ red-tldr serve
```

在 Claude Desktop 配置文件（`claude_desktop_config.json`）中添加：

```json
{
  "mcpServers": {
    "red-tldr": {
      "command": "red-tldr",
      "args": ["serve"]
    }
  }
}
```

### Streamable HTTP 模式

```bash
$ red-tldr serve --http --addr localhost:8080 --endpoint /mcp
```

## 配置文件

默认路径：`~/.red-tldr/config.toml`

```toml
[red-tldr]
  index-update = false
  github-update = false
  path = ""
  color = true
```

| 配置项 | 说明 | 类型 |
|--------|------|------|
| index-update | 搜索时是否自动重建索引 | Bool |
| github-update | 是否从 GitHub 自动更新数据库 | Bool |
| path | 数据库文件存放路径 | String |
| color | 终端输出是否带颜色高亮 | Bool |

## 数据模型

每个条目是 [red-tldr-db](https://github.com/Rvn0xsy/red-tldr-db) 仓库中的一个 YAML 文件。v0.5.0 的增强格式支持额外字段，同时保持向后兼容：

```yaml
name: mimikatz-sekurlsa
tags:
  - mimikatz
  - credentials
  - lsass
data: |
  # Mimikatz Sekurlsa
  ```
  privilege::debug
  sekurlsa::logonpasswords
  ```
category: credential-access
platforms:
  - windows
mitre_attack:
  tactics:
    - credential-access
  techniques:
    - T1003.001
metadata:
  author: Rvn0xsy
  confidence: high
related:
  - mimikatz-dpapi
```

所有新增字段（`category`、`platforms`、`mitre_attack`、`metadata`、`related`）均为可选 —— 现有 YAML 文件无需修改即可正常使用。

## 贡献

red-tldr 是一个免费且开源的项目，我们欢迎任何人贡献力量。

* 使用过程中遇到问题，请通过 [Issues](https://github.com/Rvn0xsy/red-tldr/issues) 反馈。
* Bug 修复请提交 Pull Request 到 **dev** 分支。
* 新功能请先创建 Issue 描述方案，采纳后再提交 PR。
* 欢迎改进文档，帮助更多人使用 red-tldr。
* 联系方式：rvn0xsy@gmail.com

**提醒：项目相关问题请优先在 [Issues](https://github.com/Rvn0xsy/red-tldr/issues) 中反馈，方便其他人搜索解决方案。**

## Stargazers over time

[![Stargazers over time](https://starchart.cc/Rvn0xsy/red-tldr.svg)](https://starchart.cc/Rvn0xsy/red-tldr)
