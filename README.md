# goInception (Personal Fork)

> **This repository is a personal fork of [hanchuanchuan/goInception](https://github.com/hanchuanchuan/goInception).**
> The original project was created by [hanchuanchuan](https://github.com/hanchuanchuan) and is maintained by its
> upstream contributors. All original design and the vast majority of the code belong to them.
> This fork is maintained for personal learning and internal use — it is **not** an official release.
> It is distributed under the upstream [Apache License 2.0](LICENSE), and the original copyright notices are retained.

[![Go](https://img.shields.io/badge/go-1.23%2B-brightgreen.svg)](https://golang.org/dl/)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)
[![Fork of](https://img.shields.io/badge/fork%20of-hanchuanchuan%2FgoInception-blue.svg)](https://github.com/hanchuanchuan/goInception)

**[[English]](README.md)** **[[Chinese]](README.zh.md)**

goInception is a MySQL operations tool that reviews, executes, backs up, and generates rollback statements for SQL.
It parses SQL and returns audit results based on configurable rules.

## What this fork changes

The main work lives on the `feature/parser-mysql8` branch. Compared with upstream `master`:

| Area | Change |
| --- | --- |
| SQL parser | Upgraded to `github.com/pingcap/tidb/pkg/parser`, adding MySQL 8.0 syntax support |
| CTE auditing | CTE aliases are no longer falsely reported as "table does not exist" |
| `ALGORITHM=INSTANT` | Corrected the judgment, including `ALTER INDEX ... VISIBLE/INVISIBLE` |
| Audit robustness | Fixed nil-pointer and stack-overflow crashes in the audit path; `CHECK` constraints are no longer misidentified as indexes |
| Query printing | Fixed a stack overflow caused by `--enable-query-print` |
| SET masking | Fixed a panic and a missed audit caused by parenthesized SET operands |
| Type changes | Fixed an out-of-range panic when checking a `TEXT` → `VARCHAR(n)` change |
| Tests | Added an integration test against a real MySQL 8 instance |

## Build

Requirements: **Go 1.23+**

```bash
git clone https://github.com/cutemolly/goInception.git
cd goInception
git checkout feature/parser-mysql8

go build -o goInception tidb-server/main.go
```

## Configure and run

```bash
cp config/config.toml.default config/config.toml
# edit backup_host / backup_port / backup_user / backup_password, etc.
./goInception -config=config/config.toml
```

> `config/config.toml` is listed in `.gitignore` because it holds database connection details.
> Never commit it and never force-add it with `git add -f`.

## Tests

```bash
# Unit tests
go test ./session

# MySQL 8 integration test (optional, requires a reachable instance)
export GOINCEPTION_TEST_MYSQL_HOST=127.0.0.1
export GOINCEPTION_TEST_MYSQL_PORT=3307
export GOINCEPTION_TEST_MYSQL_USER=root
export GOINCEPTION_TEST_MYSQL_PASSWORD='your-password'

go test ./session -run TestMysql8ModifyTextColumnToVarchar -v
```

The test creates a temporary `test_inc` database and drops its table afterwards.
If the target instance is unreachable, the test is skipped.

## Upstream and documentation

- Upstream repository: <https://github.com/hanchuanchuan/goInception>
- Upstream documentation: <https://hanchuanchuan.github.io/goInception/>
- Chinese documentation: <https://hanchuanchuan.github.io/goInception/zh/>
- Upstream changelog: <https://hanchuanchuan.github.io/goInception/changelog.html>
- Upstream releases / Docker image: <https://github.com/hanchuanchuan/goInception/releases>

Archery and other upstream integrations keep working through the same protocol.
For behavior that exists upstream, please consult the upstream docs and issue tracker first;
for the changes listed above, refer to this repository's commits.

## License

[Apache License 2.0](LICENSE). Upstream copyright notices are retained, and the changes in this
fork are released under the same license.

## Acknowledgments

- [hanchuanchuan/goInception](https://github.com/hanchuanchuan/goInception) — the upstream project this fork is based on
- [Inception](https://github.com/hanchuanchuan/inception) — the original MySQL auditing tool that goInception was rebuilt from
- [TiDB](https://github.com/pingcap/tidb) — SQL parser and execution engine
