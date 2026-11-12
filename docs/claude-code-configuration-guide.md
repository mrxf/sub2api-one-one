# Claude Code 接入与防风控配置建议指南

本文档面向 **Sub2API 平台管理员** 与 **下游 Claude Code CLI 用户**，针对「上游使用 Anthropic Claude OAuth / Setup Token 账号 + 固定地区代理（如美国住宅代理），下游用户分布在非美地区（如中国大陆、亚太、欧洲）」的跨区反代场景，提供一套完整的服务端与客户端配置建议。

---

## 一、风险权重总览

在跨区共享 Claude OAuth / Setup Token 账号的场景下，各维度的风控敏感度从高到低排序如下：

| 风险维度 | 风险等级 | 处理位置 | 推荐策略 |
| --- | --- | --- | --- |
| 出口 IP 纯净度与地区稳定性 | 极高 | Sub2API 账号代理 | 一号一静态原生住宅/ISP 代理（与授权登录地一致），严禁混用机房 IP 或频繁跳国家 |
| CLI 旁路遥测与错误上报直连官方 | 高 | 下游用户本机 CLI | 配置 `CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1` 等环境变量，阻断非模型流量直连 |
| 单账号多用户并发与设备指纹跳变 | 高 | Sub2API 系统与账号设置 | 开启「指纹统一化」「TLS 指纹模拟」「客户端亲和调度」，关闭「Metadata 透传」 |
| 请求头语言标识（`Accept-Language`） | 中 | Sub2API 网关转发设置 | 开启「Accept-Language 出站脱敏」（默认开启，统一改写为 `*`） |
| 跨区系统日期时差（`Today's date is …`） | 中低 | Sub2API 网关转发设置 | 开启「客户端 dateline 归一化」，按代理地区填写 IANA 时区（如 `America/Los_Angeles`） |
| 下游用户电脑系统时区 / 系统语言 | 低 | 无需修改 | 网关已完成请求头与提示词日期清洗，**无需强迫下游用户修改操作系统时区或语言** |

---

## 二、管理员端：Sub2API 服务端配置建议

### 1. 代理与网络层（最关键）

1. **一号一固定静态住宅 / ISP 代理**
   - 为每个 Claude OAuth / Setup Token 账号绑定独立的静态住宅（Static Residential / ISP）代理，避免多个账号共用同一个数据中心（IDC）机房段 IP。
   - **授权与日常转发同区**：获取 OAuth 凭证 / Setup Token 时浏览器所使用的代理地区（例如美国加州），应与绑定到 Sub2API 账号上的日常出站代理地区保持一致，避免「刚在东京登录，下一秒在洛杉矶高并发调用」。
2. **严禁出口地区反复横跳**
   - 不要为单个 Claude 账号挂载会自动轮换国家/城市的动态代理池。

---

### 2. 系统设置 -> 网关转发（全局指纹与跨区脱敏）

进入管理后台 **系统设置 -> 网关转发**，建议按以下清单核对配置：

| 配置项 | 推荐值 | 作用说明 |
| --- | --- | --- |
| **指纹统一化** (`enable_fingerprint_unification`) | **开启**（默认） | 为每个 OAuth 账号生成并锁定一套固定的 `X-Stainless-*`（OS / 架构 / Node 运行时）请求头，防止同一账号因不同下游用户交替请求而在 macOS、Windows、Linux 之间秒级切换。 |
| **Metadata 透传** (`enable_metadata_passthrough`) | **关闭**（默认） | **切勿开启**。保持关闭时，网关会将请求体中的 `metadata.user_id` 重写为绑定到该上游账号的固定 `account_uuid` 与 `device_id`，隐藏下游多个用户的真实本机设备 ID。 |
| **Accept-Language 出站脱敏** (`enable_accept_language_sanitization`) | **开启**（默认） | 将 Anthropic OAuth / Setup Token 出站请求（含 `/v1/messages` 与 `/v1/messages/count_tokens`）的 `accept-language` 统一规范化为 `*`，对齐原生 Node.js `undici` / Claude Code CLI 线格式，防止透传 `zh-CN` 等地区语言头或多用户语言头跳变。（API Key 账号不受影响） |
| **客户端 dateline 归一化** (`enable_client_dateline_normalization`) | **开启**（默认） | 抹除部分第三方客户端在检测到非官方 `ANTHROPIC_BASE_URL` 时往 `Today's date is …` 中植入的非 ASCII 撇号与 `/` 分隔符隐写指纹。 |
| **客户端 dateline 时区对齐** (`client_dateline_timezone`) | 按代理地区填写（如 `America/Los_Angeles`） | 若你的 OAuth 账号代理集中在美西，填写 `America/Los_Angeles`（美东填 `America/New_York`）。网关会将 `system` 提示词与 `<system-reminder>` 块内的 `Today's date is YYYY-MM-DD.` 自动对齐为该时区当前日期，消除东八区与美国出口跨日时段（15–16 小时）的日期错位。若代理分布在多个不同国家/时区，可留空（仅清洗隐写字符，不改日期）。 |
| **自动同步 Claude Code 版本号** (`claude_code_version_auto_sync_enabled`) | **开启**（默认） | 自动跟踪官方最新 Claude Code CLI 版本号，「Claude Code 客户端版本号」留空即可跟随同步值，避免旧版本号触发 `claude_code_version_too_old` 或风控降级。 |
| **Claude OAuth System 注入** (`enable_claude_oauth_system_prompt_injection`) | **开启**（默认） | 为非官方 CLI 的请求补齐标准 Claude Code `system` 块形态与计费归属头。 |

---

### 3. 账号管理 & 分组策略（单账号防风控与限流）

在 **账号管理** 编辑或新建 Anthropic OAuth / Setup Token 账号时，建议配置：

1. **开启「TLS 指纹模拟」**
   - 模拟 Node.js / Claude Code 客户端原生 TLS 握手特征（Cipher Suites / ALPN / 扩展顺序），避免上游通过 TLS 指纹识别出后端 Go HTTP 客户端。
2. **开启「客户端亲和调度」并配合粘性会话**
   - 让同一用户的连续对话和新会话优先固定落到同一个上游 Claude 账号上，减少跨账号切流与单账号短时多用户混杂。
3. **按需开启「会话 ID 伪装」**
   - 若单账号下存在多个下游用户发起高频短任务，开启后可在 15 分钟窗口内收敛 `metadata.user_id` 中的会话 ID 变化频率。
4. **控制单账号并发上限与 RPM**
   - 正常个人订阅的 Claude Code CLI 很少同时发起十几个无关会话的并发请求。建议将单个 OAuth 账号并发控制在合理区间（例如 `2`–`5`），并配置基础 RPM 与「软性限速 / 粘性豁免」，靠多账号横向扩容承载总并发，而不是把单个账号跑满。
5. **分组客户端限制（可选）**
   - 若某分组专门为 Claude Code CLI 用户提供服务，可在分组设置中开启 **仅允许 Claude Code 客户端**，拦截未做遥测屏蔽或特征粗糙的野生第三方脚本。

---

## 三、客户端：下游 Claude Code CLI 配置指南（可发给用户）

> **说明**：你无需修改自己电脑的系统时区或系统语言（服务端网关已自动完成请求头与日期脱敏）。但在使用非官方中转地址时，**务必在配置中加入 `CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1` 等环境变量**，防止 Claude Code CLI 在后台绕过中转站，直接用你本机的真实网络向 Anthropic / Statsig / Sentry 发送遥测、自动检查更新或错误报告。

### 方式一：持久化配置文件（推荐，一劳永逸）

编辑本机 Claude Code 配置文件：
- **macOS / Linux**：`~/.claude/settings.json`
- **Windows**：`%USERPROFILE%\.claude\settings.json`

写入以下内容（将 `ANTHROPIC_BASE_URL` 和 `ANTHROPIC_AUTH_TOKEN` 替换为平台分配给你的地址和 API Key）：

```json
{
  "$schema": "https://json.schemastore.org/claude-code-settings.json",
  "env": {
    "ANTHROPIC_BASE_URL": "https://your-sub2api-domain.com",
    "ANTHROPIC_AUTH_TOKEN": "sk-your-sub2api-api-key",
    "CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC": "1",
    "DISABLE_TELEMETRY": "1",
    "DISABLE_ERROR_REPORTING": "1",
    "DISABLE_BUG_COMMAND": "1"
  }
}
```

**关键环境变量说明：**
- `ANTHROPIC_BASE_URL`：指向 Sub2API 网关根地址（不要带 `/v1` 尾缀，例如 `https://api.example.com`）。
- `ANTHROPIC_AUTH_TOKEN`：你在 Sub2API 后台创建的 `sk-...` 密钥。
- `CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC="1"`：**核心总开关**。一次性关闭自动更新检查、`/bug` 上报、Sentry 错误上报与 Statsig 遥测等所有非模型推理请求。
- `DISABLE_TELEMETRY="1"` / `DISABLE_ERROR_REPORTING="1"` / `DISABLE_BUG_COMMAND="1"`：显式兜底开关，兼容不同历史版本的 Claude Code CLI，确保零旁路遥测请求直连官方。

---

### 方式二：终端环境变量（仅当前终端窗口生效）

如果你不想修改全局 `settings.json`，可以在启动 `claude` 前在终端执行：

**macOS / Linux (Bash / Zsh)：**

```bash
export ANTHROPIC_BASE_URL="https://your-sub2api-domain.com"
export ANTHROPIC_AUTH_TOKEN="sk-your-sub2api-api-key"
export CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1
export DISABLE_TELEMETRY=1
export DISABLE_ERROR_REPORTING=1
export DISABLE_BUG_COMMAND=1
claude
```

**Windows (PowerShell)：**

```powershell
$env:ANTHROPIC_BASE_URL="https://your-sub2api-domain.com"
$env:ANTHROPIC_AUTH_TOKEN="sk-your-sub2api-api-key"
$env:CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC="1"
$env:DISABLE_TELEMETRY="1"
$env:DISABLE_ERROR_REPORTING="1"
$env:DISABLE_BUG_COMMAND="1"
claude
```

**Windows (CMD)：**

```cmd
set ANTHROPIC_BASE_URL=https://your-sub2api-domain.com
set ANTHROPIC_AUTH_TOKEN=sk-your-sub2api-api-key
set CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1
set DISABLE_TELEMETRY=1
set DISABLE_ERROR_REPORTING=1
set DISABLE_BUG_COMMAND=1
claude
```

---

### 客户端日常使用注意事项

1. **清理本机旧的官方登录状态，避免凭证冲突**
   - 如果之前在本机执行过官方 `claude login` 登录过个人 Claude 账号，建议先在 CLI 内执行 `/logout` 退出官方登录态，同时检查系统环境变量中不要残留旧的 `ANTHROPIC_API_KEY`，仅保留上面的 `ANTHROPIC_BASE_URL` + `ANTHROPIC_AUTH_TOKEN`。
2. **定期手动更新 Claude Code CLI**
   - 开启 `CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1` 后，CLI 的后台自动静默升级也会被关闭（这是预期的安全行为）。当版本过旧时，可手动执行以下命令升级：
     ```bash
     npm install -g @anthropic-ai/claude-code@latest
     ```
3. **切勿泄露 API Key**
   - 若在项目目录下使用 `.claude/settings.local.json`，请确保将其加入 `.gitignore`，切勿将包含 `ANTHROPIC_AUTH_TOKEN` 的配置文件提交至代码仓库。

---

## 四、管理员排障与出站报文自检（可选）

如果需要验证网关发往 Anthropic 上游的实际请求头与请求体是否已完成脱敏，可在测试环境为 Sub2API 后端设置环境变量：

```bash
SUB2API_DEBUG_GATEWAY_BODY=/tmp/gateway_debug.log
```

发起一次测试请求后，查看 `/tmp/gateway_debug.log` 中的 `CLIENT_ORIGINAL` 与 `UPSTREAM_FORWARD` 快照对比，确认 `UPSTREAM_FORWARD` 满足以下特征即为生效：
1. `headers` 中 `accept-language` 已固定为 `*`（即使客户端原始请求带了 `zh-CN` 或未带该头）。
2. `headers` 中 `X-Stainless-*` 与 `User-Agent` 稳定一致，`metadata.user_id` 已被重写为上游账号绑定的身份。
3. 若配置了 `client_dateline_timezone`（如 `America/Los_Angeles`），`body` 的 `system` 与 `<system-reminder>` 中 `Today's date is YYYY-MM-DD.` 已与目标时区当前日期一致。
4. 验证完毕后请移除 `SUB2API_DEBUG_GATEWAY_BODY` 环境变量并清理调试日志文件，避免在生产磁盘上持续记录完整请求体。
