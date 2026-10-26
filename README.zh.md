# goInception（个人改造版）

> **本项目 Fork 自 [hanchuanchuan/goInception](https://github.com/hanchuanchuan/goInception)。**
> 原始项目由 [hanchuanchuan](https://github.com/hanchuanchuan) 创建，并由上游贡献者共同维护。
> 本仓库是个人维护的改造版本，主要用于学习与内部实践，**不是官方版本**。
> 项目遵循上游的 [Apache License 2.0](LICENSE)，原始版权声明予以保留。

[![Go](https://img.shields.io/badge/go-1.23%2B-brightgreen.svg)](https://golang.org/dl/)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)
[![Fork of](https://img.shields.io/badge/fork%20of-hanchuanchuan%2FgoInception-blue.svg)](https://github.com/hanchuanchuan/goInception)

**[[English]](README.md)** **[[中文]](README.zh.md)**

goInception 是一个集审核、执行、备份及生成回滚语句于一身的 MySQL 运维工具。
它通过解析 SQL 语法，按照自定义规则返回审核结果，并提供执行、备份和回滚语句生成能力。

## 相对于上游的改造

主要改动位于 `feature/parser-mysql8` 分支。与上游 `master` 相比：

| 方向 | 说明 |
| --- | --- |
| SQL 解析器 | 升级到 `github.com/pingcap/tidb/pkg/parser`，支持解析 MySQL 8.0 语法 |
| CTE 审核 | CTE 别名不再被误报为“表不存在” |
| `ALGORITHM=INSTANT` | 修正判定逻辑，覆盖 `ALTER INDEX ... VISIBLE/INVISIBLE` 等场景 |
| 审核稳定性 | 修复审核路径中的空指针与栈溢出；`CHECK` 约束不再被误判为索引 |
| 查询打印 | 修复 `--enable-query-print` 导致的栈溢出 |
| SET 脱敏 | 修复括号包裹的 SET 操作数导致的 panic 与审核漏检 |
| 类型变更 | 修复 `TEXT` 改为 `VARCHAR(n)` 时类型判断的越界 panic |
| 测试 | 新增基于真实 MySQL 8 实例的集成测试 |

## 编译

环境要求：**Go 1.23+**

```bash
git clone https://github.com/cutemolly/goInception.git
cd goInception
git checkout feature/parser-mysql8

go build -o goInception tidb-server/main.go
```

## 配置与运行

```bash
cp config/config.toml.default config/config.toml
# 按需修改 backup_host / backup_port / backup_user / backup_password 等配置
./goInception -config=config/config.toml
```

> `config/config.toml` 已加入 `.gitignore`，其中包含数据库连接信息。
> 请勿提交该文件，也不要使用 `git add -f` 强制添加。

## 测试

```bash
# 单元测试
go test ./session

# MySQL 8 集成测试（可选，需要可连接的实例）
export GOINCEPTION_TEST_MYSQL_HOST=127.0.0.1
export GOINCEPTION_TEST_MYSQL_PORT=3307
export GOINCEPTION_TEST_MYSQL_USER=root
export GOINCEPTION_TEST_MYSQL_PASSWORD='你的密码'

go test ./session -run TestMysql8ModifyTextColumnToVarchar -v
```

该测试会创建临时的 `test_inc` 库并在结束后删除表；目标实例不可达时会自动跳过。

## 上游与文档

- 上游仓库：<https://github.com/hanchuanchuan/goInception>
- 上游文档：<https://hanchuanchuan.github.io/goInception/>
- 中文文档：<https://hanchuanchuan.github.io/goInception/zh/>
- 上游更新日志：<https://hanchuanchuan.github.io/goInception/changelog.html>
- 上游发布与 Docker 镜像：<https://github.com/hanchuanchuan/goInception/releases>

Archery 等上游生态仍可通过原有协议接入。
上游已有功能的行为请优先查阅上游文档与 issue；
本仓库列出的改造内容以本文档和提交记录为准。

## 许可证

[Apache License 2.0](LICENSE)。本仓库保留上游版权声明，改造部分同样以该许可证发布。

## 致谢

- [hanchuanchuan/goInception](https://github.com/hanchuanchuan/goInception) —— 本 Fork 的上游项目
- [Inception](https://github.com/hanchuanchuan/inception) —— goInception 重构所基于的 MySQL 审核工具
- [TiDB](https://github.com/pingcap/tidb) —— SQL 解析与执行引擎
