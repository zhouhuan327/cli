# 知识整理工作流：Discovery

Loaded by states: `PARSE_SCOPE`, `INVENTORY`.

This file owns target parsing, scope clarification, resource inventory, ResourceItem normalization, dedupe, and partial inventory handling. It MUST NOT generate classification rules, execution plans, or perform write operations.

## Required Context

Before executing rules in this file:

1. Follow [`../../lark-shared/SKILL.md`](../../lark-shared/SKILL.md) for identity, auth, and permission handling.
2. For Wiki / personal library targets, follow [`../../lark-wiki/SKILL.md`](../../lark-wiki/SKILL.md).
3. For Drive search targets, follow [`lark-drive-search.md`](lark-drive-search.md).
4. For URL / token inspection, follow [`lark-drive-inspect.md`](lark-drive-inspect.md) and [`../../lark-wiki/references/lark-wiki-node-get.md`](../../lark-wiki/references/lark-wiki-node-get.md).

## State: PARSE_SCOPE

Entry: workflow triggered.

MUST:

1. Identify `target_scope`, `environment_profile`, and `identity`.
2. Apply `Scope Parsing`.
3. Output `Scope Confirmation`.
4. Stop and wait for user confirmation before `INVENTORY`.

Exit: user confirms target scope.

### Scope Parsing

| Condition | Agent MUST Do | Set `target_scope` | Next State |
|-----------|---------------|--------------------|------------|
| Input is `/wiki/<token>` URL | Resolve the Wiki node and preserve both node identity and object identity | Wiki node | `INVENTORY` after user confirms scope |
| Input is Wiki space name / `space_id` | Resolve the Wiki space; 0 matches -> stop and ask; 1 exact match -> continue; multiple matches -> show candidates and wait for user selection; do not treat `my_library` as a normal listed space | Wiki space | `INVENTORY` after user confirms scope |
| Input is `我的文档库` / `个人知识库` / `my_library` | Treat as Wiki personal library and resolve real `space_id` before root write | Personal doc library | `INVENTORY` after user confirms scope |
| Input is `/drive/folder/<token>` URL | Extract `folder_token` | Drive folder | `INVENTORY` after user confirms scope |
| Input is single cloud resource URL | Resolve the resource type; if not folder / Wiki scope, do not expand automatically | Single resource | Ask whether scope is this resource, parent folder, owning Wiki, or related search results |
| Input is real keyword / name | Search with the real keyword according to `lark-drive-search` | Search scope | `INVENTORY` after user confirms scope |
| Input is range browsing / statistical description with no real keyword | Search by filters / empty-query browsing according to `lark-drive-search` | Search scope | `INVENTORY` after user confirms scope |
| Input is ambiguous | Ask the minimum clarification question and stop | Unknown | Stay in `PARSE_SCOPE` |

### Stop Conditions

Stop and ask for clarification when:

1. 用户只说"整理文件夹"、"整理目录"、"整理资料"、"整理文档"，但没有 URL、token、知识库名称、Drive 文件夹、个人文档库或明确搜索范围。
2. 用户说"我的文件夹"、"我的目录"，但无法唯一判断是 Drive 文件夹、个人文档库还是某个 Wiki 节点。
3. 用户给的是单个资源 URL，但要求"整理一批文档"或"整理相关资料"。
4. 用户目标环境不明确，且上下文中同时存在线上、BOE、PRE 或多个 profile。

Clarification template:

```text
请提供要整理的 Drive 文件夹链接、Wiki 节点 / 知识库链接，或明确说明要整理"我的文档库"；如果只想按关键词搜索整理，也请给出关键词或范围。
```

### Scope Confirmation

```text
我先确认本次整理范围。

目标：
范围：
环境 / profile：
身份：
预计操作：先盘点并生成整理方案，不执行移动或创建。

请确认是否按这个范围继续？
```

## State: INVENTORY

Entry: `target_scope` confirmed.

MUST:

1. Recursively list resources according to target type.
2. Generate `path` during traversal.
3. Normalize all results to `ResourceItem`.
4. Track pagination, depth, and item limits.
5. Set `partial=true` when limits are hit.
6. Output `Inventory Summary`.
7. Continue to `CONTENT_READ` without asking the user unless auth, permission, API, target scope, or environment blockers occur.

### Inventory Limits

| Scope | Default Limit | If Limit Is Hit |
|-------|---------------|-----------------|
| Wiki recursion | `max_depth=3`, `max_items=500`; follow `lark-wiki-node-list` pagination | Set `partial=true`; list covered paths and suggested next first-level directories |
| Drive folder recursion | `max_depth=3`, `max_items=500`, max 10 pages per folder, `page_size=50` | Set `partial=true`; list folders not drilled into |
| Search discovery | Follow `lark-drive-search` pagination policy | Set `partial=true`; keep continuation information |

If the user explicitly asks for full processing, batch by first-level directory, Wiki space, or time window. Do not remove all limits in one run.

### Wiki Inventory Rules

1. Follow [`../../lark-wiki/references/lark-wiki-node-list.md`](../../lark-wiki/references/lark-wiki-node-list.md) traversal semantics.
2. Generate stable paths from parent-child traversal.
3. Preserve Wiki node identity fields needed by `ResourceItem`.
4. Treat `my_library` as Wiki personal library, not Drive root.

### Drive Inventory Rules

1. Use CLI command family `drive files list` according to `lark-drive` API rules; its schema path is `drive.files.list`.
2. Recurse only into `folder` items.
3. Use `drive metas batch_query` when URL, owner, created time, or updated time is needed.
4. Continue pages by feeding `next_page_token` into request param `page_token`.
5. Prefer explicit `folder_token`; querying root with empty `folder_token` may return broad root data and may not paginate as expected.

### Search Inventory Rules

1. Search results may be normalized directly only when they include stable identity fields required by `ResourceItem`.
2. If a search result is a Wiki item and lacks `node_token`, resolve it with `drive +inspect` or `wiki +node-get` before dedupe.
3. If Wiki identity still cannot be resolved, keep the item, set `needs_review=true`, and record `needs_review_reason`.

## ResourceItem

Agent MUST normalize Wiki, Drive, and search results into `ResourceItem`. Later statistics, classification, and planning MUST use this model rather than raw API responses.

```json
{
  "source": "wiki|drive|search",
  "title": "资源标题",
  "type": "doc|docx|sheet|bitable|mindnote|file|wiki|folder|slides|shortcut|catalog",
  "path": "当前路径/资源标题",
  "depth": 2,
  "url": "https://...",
  "token": "canonical_token",
  "node_token": "wiki_node_token_or_empty",
  "obj_token": "wiki_obj_token_or_drive_file_token",
  "node_type": "origin|shortcut|empty",
  "origin_node_token": "wiki_origin_node_token_or_empty",
  "space_id": "wiki_space_id_or_empty",
  "parent_token": "parent_node_or_folder_token",
  "has_child": false,
  "dedupe_key": "wiki:<space_id>:<node_token>|drive:<type>:<token>|search:<type>:<token>",
  "created_at": "optional",
  "updated_at": "optional",
  "needs_review": false,
  "needs_review_reason": ""
}
```

ResourceItem rules:

1. `path` MUST be generated by recursion. Do not use title alone as path.
2. Wiki URL token may not be the underlying document token. Preserve both `node_token` and `obj_token`.
3. `type` MUST come from API fields such as `obj_type` / `doc_type`.
4. Wiki organization is by node instance. Prefer `wiki:<space_id>:<node_token>` as `dedupe_key`.
5. MUST NOT dedupe Wiki nodes only by `obj_token`; one document can appear under different Wiki paths or shortcuts.
6. If `node_type=shortcut` or dedupe is uncertain, use `wiki +node-get` to supplement `origin_node_token`; if unavailable, leave empty and set `needs_review=true`.
7. Drive folder tree dedupes by `drive:<type>:<token>`.
8. Search results may merge with recursive results only by exact identity: Wiki by same `node_token`, Drive by same `type + token`.

## Inventory Summary

```text
已完成盘点。

| 指标 | 数量 |
|------|------|
| 总资源数 |  |
| 各类型资源数 |  |
| 一级目录数量 |  |
| 根目录直接资源数 |  |
| 空目录数量 |  |
| 疑似临时 / 测试 / 未整理资源数 |  |
| 低置信度待确认资源数 |  |

下一步将自动读取低置信度资源并分析整理问题；不会执行移动或创建。
```

## Discovery Failure Handling

| Failure / Blocker | Agent MUST Do | Agent MUST NOT Do |
|-------------------|---------------|-------------------|
| Target scope is ambiguous | Ask the minimum scope clarification question and stop | Do not choose a whole cloud drive / personal library by default |
| Environment / profile is ambiguous | Ask user to confirm prod / BOE / PRE and profile | Do not cross environment boundaries |
| Missing API scope | Follow `lark-shared` permission handling and stop | Do not retry the same command repeatedly |
| Resource access denied | Stop and follow the main workflow `Permission Request Gate` | Do not request permission automatically or in batch |
| Pagination / depth / item limit reached | Set `partial=true`; record uncovered range and continuation command | Do not claim full coverage |
