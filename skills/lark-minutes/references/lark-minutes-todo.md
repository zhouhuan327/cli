# minutes +todo

> **前置条件：** 先阅读 [`../lark-shared/SKILL.md`](../../lark-shared/SKILL.md) 了解认证、全局参数和安全规则。

更新妙记中的一条或多条待办（内容与完成状态）。写操作。

本 skill 对应 shortcut：`lark-cli minutes +todo`（调用 `PUT /open-apis/minutes/v1/minutes/{minute_token}/todo`）。

## 典型触发表达

- "把这几条待办改成……"
- "标记某条待办为已完成"
- "批量更新妙记里的待办事项"
- "重新打开 / 取消完成某条待办"

## 命令

```bash
# 单条待办（内容与 is_done 成对）
lark-cli minutes +todo --minute-token obcnxxxxxxxxxxxxxxxxxxxx --todo "跟进预算审批" --is-done

# 单条待办标记为未完成
lark-cli minutes +todo --minute-token obcnxxxxxxxxxxxxxxxxxxxx --todo "整理会议纪要" --is-done=false

# 多条待办（JSON 数组）
lark-cli minutes +todo --minute-token obcnxxxxxxxxxxxxxxxxxxxx --todo-list @todos.json

# 从 stdin 读取 todo_list
cat todos.json | lark-cli minutes +todo --minute-token obcnxxxxxxxxxxxxxxxxxxxx --todo-list @-

# 预览 API 调用
lark-cli minutes +todo --minute-token obcnxxxxxxxxxxxxxxxxxxxx --todo-list @todos.json --dry-run
```

`todos.json` 示例：

```json
[
  {"content": "跟进预算审批", "is_done": true},
  {"content": "整理会议纪要", "is_done": false}
]
```

## 参数

| 参数 | 必填 | 说明 |
|------|------|------|
| `--minute-token <token>` | 是 | 妙记 Token |
| `--todo-list <json>` | 二选一 | 待办数组 JSON；支持 `@file` / `@-`（stdin）；与 `--todo`/`--is-done` 互斥 |
| `--todo <text>` | 二选一 | 单条待办纯文本；必须与 `--is-done` 成对出现 |
| `--is-done` | 二选一 | 单条完成状态布尔值；传 `--is-done` 表示 `true`，传 `--is-done=false` 表示 `false` |
| `--dry-run` | 否 | 预览 API 调用，不执行 |

## 核心约束

### 1. 先读后写

更新前建议先用 `lark-cli vc +notes --minute-tokens <token>` 读取当前待办列表，确认内容与 `is_done` 状态。

读取与写入均使用 `is_done` 布尔字段。已删除的待办不会出现在读取结果中。

### 2. 待办内容为纯文本

`content` **不是 Markdown**，请直接传入待办描述文字。

- 不要写 `# 标题`、`**加粗**`、`- 列表` 等 Markdown 语法
- 如需多行内容，可直接使用换行；但不会被渲染为 Markdown 格式

### 3. 请求体字段

| CLI / JSON | API 字段 | 说明 |
|------------|---------|------|
| `content` | `content` | 纯文本待办描述（必填） |
| `is_done` | `is_done` | 是否已完成（必填） |
| `assignees` | `assignees` | 可选负责人列表（OpenAPI 类型支持；下游编辑当前仅使用 content 与 is_done） |

### 4. 所需权限

| 身份 | 所需权限 |
|------|---------|
| user | `minutes:minutes:update` |

## 输出结果

```json
{
  "minute_token": "obcnxxxxxxxxxxxxxxxxxxxx",
  "todo_count": 2,
  "updated": true
}
```

| 字段 | 说明 |
|------|------|
| `minute_token` | 妙记 Token |
| `todo_count` | 本次提交的待办条数 |
| `updated` | 是否已成功更新 |

## 如何获取 minute_token

| 来源 | 获取方式 |
|------|---------|
| 妙记 URL | 从 URL 末尾提取，如 `https://sample.feishu.cn/minutes/obcnxxxxxxxxxxxxxxxxxxxx` |
| 妙记搜索 | `lark-cli minutes +search --query "关键词"` |
| 会议产物查询 | `lark-cli vc +notes --minute-tokens <token>` |

## 常见错误与排查

| 错误现象 | 根本原因 | 解决方案 |
|---------|---------|---------|
| 参数无效 | `minute_token` 缺失，或 `todo_list` 为空 | 检查 token 与 JSON 数组 |
| 缺少 `is_done` | 只传了 `--todo` 未传 `--is-done` | `--todo` 与 `--is-done` 必须成对出现 |
| 互斥参数 | 同时传了 `--todo-list` 与 `--todo`/`--is-done` | 只选一种写法 |
| 权限不足 | 缺少 `minutes:minutes:update` | 运行 `auth login --scope "minutes:minutes:update"` |

## 参考

- [lark-minutes](../SKILL.md) — 妙记全部命令
- [minutes +summary](lark-minutes-summary.md) — 替换 AI 总结（不支持的 Markdown 会按原始文本展示，详见该文档）
- [lark-vc-notes](../../lark-vc/references/lark-vc-notes.md) — 读取总结、待办等 AI 产物
- [lark-shared](../../lark-shared/SKILL.md) — 认证和全局参数
