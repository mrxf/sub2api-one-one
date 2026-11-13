# GitHub Actions 触发打包 Docker 新版本镜像指南

本文档说明如何在手动操作或**新的 AI Agent 会话**中，通过 GitHub Actions 触发打包并发布 `mrxf/sub2api-one-one` 的最新 Docker 镜像，以及如何在服务器上平滑拉取更新。

> **给其他会话 AI Agent 的快速提示**：
> 如果用户让你「打包新版本 Docker 镜像」或「改完代码顺便把新镜像打出来」，请直接按本文档 **「三、AI Agent 标准自动化执行流程（SOP）」** 逐步执行。

---

## 一、核心配置速览

| 项目 | 值 / 说明 |
| --- | --- |
| **GitHub 仓库** | `mrxf/sub2api-one-one`（默认分支：`main`） |
| **打包工作流文件** | `.github/workflows/docker-publish.yml`（名称：`Docker Publish (GHCR)`） |
| **目标镜像仓库** | `ghcr.io/mrxf/sub2api-one-one`（公开镜像，服务器免 `docker login` 直接拉取） |
| **默认目标架构** | `linux/amd64`（可通过 `platforms` 参数指定 `linux/amd64,linux/arm64`） |
| **自动生成的镜像标签** | `:latest`、`:main`、`:sha-<7位commit>`（打 `v*` 标签时额外生成 `:x.y.z` 与 `:x.y`） |
| **程序内置版本号来源** | 构建时自动读取 `backend/cmd/server/VERSION`（若当前提交精确命中 `v*` Git Tag 则优先取 Tag 号） |

---

## 二、触发打包的三种方式

### 方式 1：使用 `gh` CLI 命令行直接触发（最推荐，适合 AI Agent 会话）

> **为什么要显式执行 `gh workflow run`？**
> 1. 在云端 AI Agent（如 Antigravity）中，Git 推送使用的是 GitHub App 安装令牌，GitHub 防递归策略有时不会自动触发 `on: push` 工作流；
> 2. 工作流设置了 `paths-ignore`（忽略纯文档 `docs/**` 和编排配置 `deploy/**` 改动）。
> 因此推送代码后，直接用 `gh workflow run` 显式触发最稳妥。

```bash
# 默认触发（基于最新 main 分支打包 linux/amd64 架构，更新 :latest）
gh workflow run docker-publish.yml -R mrxf/sub2api-one-one --ref main

# 如需同时打包 x86_64 与 ARM64 多架构镜像：
gh workflow run docker-publish.yml -R mrxf/sub2api-one-one --ref main \
  -f platforms="linux/amd64,linux/arm64"
```

---

### 方式 2：在 GitHub 网页端手动点击触发

1. 打开仓库 Actions 页面：`https://github.com/mrxf/sub2api-one-one/actions/workflows/docker-publish.yml`
2. 点击右侧 **Run workflow** 按钮：
   - **Use workflow from**：选择 `Branch: main`
   - **Target Docker platforms**：保持默认 `linux/amd64`（若服务器为 ARM 架构则改为 `linux/amd64,linux/arm64`）
3. 点击绿色的 **Run workflow** 即可开始构建（开启了 Buildx 缓存，通常 2～4 分钟完成）。

---

### 方式 3：升级版本号并打包新版本镜像

如果你希望管理后台左下角/系统信息里显示新的版本号（或生成固定语义化版本标签）：

1. **轻量做法（仅更新镜像与内置版本号，不跑全平台二进制 Release）**：
   - 修改 `backend/cmd/server/VERSION` 文件中的版本号（例如从 `0.2.21` 改为 `0.2.22`）；
   - 提交并推送到 `main` 分支；
   - 执行 `gh workflow run docker-publish.yml -R mrxf/sub2api-one-one --ref main`。
   - 打包完成后，`:latest` 和对应的 `:sha-<commit>` 镜像内二进制版本即为 `0.2.22`。

2. **完整发版做法（生成镜像 `:0.2.22` 标签 + GitHub Release）**：
   - 推送 `v*` 格式的 Git Tag（例如 `git tag v0.2.22 && git push origin v0.2.22`）。
   - **注意**：推送 `v*` Tag 除了触发 `docker-publish.yml` 生成 `ghcr.io/mrxf/sub2api-one-one:0.2.22` 和 `:latest` 外，还会同时触发 `.github/workflows/release.yml` 构建全平台二进制压缩包并创建 GitHub Release。日常自用部署推荐用上面的「轻量做法」。

---

## 三、AI Agent 标准自动化执行流程（SOP）

在新的会话中，当用户要求修改代码并打包新镜像（或直接触发打包新镜像）时，请按以下 4 步闭环执行：

### Step 1：确认提交已推送到远端 `main`

```bash
git status --short
git push origin main
```

### Step 2：触发 `docker-publish.yml` 并查询新生成的 Run ID

```bash
gh workflow run docker-publish.yml -R mrxf/sub2api-one-one --ref main
sleep 4
gh run list -R mrxf/sub2api-one-one --workflow docker-publish.yml --limit 3 \
  --json databaseId,status,conclusion,headSha,url
```

### Step 3：监控构建直到成功退出

将上一步查到的最新 `databaseId` 填入 `<RUN_ID>`（在 AI Agent 中请置于后台任务运行，避免前台超时中断）：

```bash
gh run watch <RUN_ID> -R mrxf/sub2api-one-one --exit-status --compact
```

若构建失败，使用以下命令仅拉取失败步骤日志进行排查：

```bash
gh run view <RUN_ID> -R mrxf/sub2api-one-one --log-failed | tail -n 200
```

### Step 4：免登录验证 GHCR 远端 `:latest` 镜像摘要（Digest）

构建成功后，通过匿名 Token 验证新镜像已可公开拉取并记录最新 `docker-content-digest`：

```bash
ANON_TOK=$(curl -s "https://ghcr.io/token?scope=repository:mrxf/sub2api-one-one:pull" | python3 -c "import sys,json; print(json.load(sys.stdin).get('token',''))")
curl -sI -H "Authorization: Bearer $ANON_TOK" \
  -H "Accept: application/vnd.oci.image.index.v1+json, application/vnd.docker.distribution.manifest.v2+json" \
  "https://ghcr.io/v2/mrxf/sub2api-one-one/manifests/latest" | grep -iE "HTTP/|docker-content-digest"
```

返回 `HTTP/2 200` 及新的 `sha256:...` 即代表镜像发布成功。

---

## 四、服务器侧拉取新镜像并平滑升级

镜像打包完成后，去服务器上的独立部署目录（例如 `/opt/sub2api-one-one`）执行以下命令，即可在不影响线上其他实例、不丢失数据库与配置的前提下完成秒级滚动更新：

```bash
cd /opt/sub2api-one-one
docker compose pull sub2api && docker compose up -d sub2api
docker compose logs -f --tail=50 sub2api
```

如果需要回滚到某次历史提交对应的镜像，只需将 `/opt/sub2api-one-one/.env` 中的 `SUB2API_IMAGE` 改为对应的 `sha-` 标签（可在仓库 Packages 页面查看）：

```ini
SUB2API_IMAGE=ghcr.io/mrxf/sub2api-one-one:sha-a8ed053
```

然后重新执行 `docker compose up -d sub2api` 即可回滚。
