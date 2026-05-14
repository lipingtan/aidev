---
name: api-test
description: 执行和解读 API 自动化测试。触发场景：(1) 修改代码后需要验证接口可用性 (2) git commit/push 前的接口检查 (3) 发布版本前的回归测试 (4) 用户明确要求运行 API 测试，或用具体环境跑相关测试 (5) 合并到 main 分支前的自动化检查。选择合适的测试场景或者测试套件，执行测试并解释结果，但不修改测试或命令本身。
---

# API Tests

使用 Python 脚本执行 API 自动化测试并解释结果。

## 工作流程

1. **选择测试**：

- 如果用户在对话中明确提供了：
  - 测试文件路径
  - 测试文件名
  - 或清晰可唯一匹配的测试名称
- 直接使用该测试，不再进行自动选择
- 若信息不明确，应优先通过运行 `python ./.trae/skills/api-tests/scripts/list_tests.py` 脚本来快速获取所有的测试文件路径及描述。
- 避免在庞大的项目目录中进行盲目的全局搜索，直接定位到该 Skill 专用的测试文件目录 `tests/`。

2. **多测试执行规则**：

- 默认只执行一个测试，但需给出是否批量执行的选择
- 如果用户明确表示：
  - "跑这几个"
  - "全部跑一遍"
- 则进入**批量执行模式**
- 批量执行模式下：
  - 明确列出将要执行的测试列表
  - **询问执行方式**：让用户选择 "按顺序执行" (更易读) 或 "并行执行" (速度快)。
    - **按顺序执行**：逐个运行测试并即时分析，适合调试。
    - **并行执行**：同时启动多个测试，适合快速回归，但日志可能会交织。
  - 请求用户确认执行方式及测试列表（是 / 否）
  - 按选择的方式执行测试
  - 最终汇总或依次解释每个测试的结果

3. **确认环境**：

- 支持的环境包括：
  - `dev` - 开发环境
  - `test` - 测试环境
  - `staging` - 预发布环境
  - `prod` - 生产环境
- 如用户未明确指定环境：
  - 列出上述环境名
  - 让用户确认使用哪一个

4. **执行测试**：

- 在以下信息明确后执行测试：
  - 测试文件路径
  - 环境名（dev / test / staging / prod）
  - 执行模式（顺序 / 并行）

```bash
# 单个测试
python ./.trae/skills/api-tests/scripts/run_cli.py <测试文件路径> <环境名>

# 批量顺序执行
python ./.trae/skills/api-tests/scripts/run_cli.py <测试文件路径> <环境名>

# 批量并行执行
python ./.trae/skills/api-tests/scripts/run_cli.py <测试文件路径> <环境名> parallel
```

5. **解释结果**：

- 分析测试脚本输出
- 解析 JSON 格式的测试报告
- 统计通过率、失败用例
- 解释失败原因

## 脚本说明

### list_tests.py

列出所有可用的测试文件。

```bash
python ./.trae/skills/api-tests/scripts/list_tests.py
```

**输出示例：**
```
可用的测试文件:

1. health-check.json (0.45 KB)
2. user-auth.json (1.23 KB)
```

### run_cli.py

执行指定的测试用例。

**语法：**
```bash
python ./.trae/skills/api-tests/scripts/run_cli.py <测试文件路径> <环境名> [parallel]
```

**参数：**
- `测试文件路径` - 测试文件的相对或绝对路径
- `环境名` - 环境名称（dev/test/staging/prod）
- `parallel` - 可选，启用并行执行

**示例：**
```bash
# 执行单个测试
python ./.trae/skills/api-tests/scripts/run_cli.py tests/user-auth.json test

# 批量顺序执行
python ./.trae/skills/api-tests/scripts/run_cli.py tests/ test

# 批量并行执行
python ./.trae/skills/api-tests/scripts/run_cli.py tests/ test parallel
```

## 配置文件

### env.config.json

环境配置文件，包含项目信息、环境配置、测试套件等。

**配置结构：**
```json
{
  "projectId": "your-api-project-id",
  "environments": {
    "dev": {
      "baseUrl": "http://localhost:8080"
    },
    "test": {
      "baseUrl": "https://api.test.example.com"
    }
  },
  "suites": {
    "smoke": {
      "testFiles": ["tests/health-check.json"]
    }
  }
}
```

## 测试报告

报告生成位置：`./test-reports/`

**报告格式：**
- **JSON** - 机器可读的结构化数据，用于 CI/CD 集成
- **TXT** - 人类可读的文本报告，包含详细的测试结果和失败原因

**报告内容：**
- 测试执行摘要
- 通过率统计
- 失败用例详情
- 请求/响应信息
- 失败原因分析

**文本报告示例：**
```
================================================================================
测试报告: SimpleController 测试
================================================================================

环境: dev
基础URL: http://localhost:80
执行时间: 2026-01-28T12:00:00Z

--------------------------------------------------------------------------------
测试统计
--------------------------------------------------------------------------------
总用例数: 9
通过: 9
失败: 0
通过率: 100.00%

--------------------------------------------------------------------------------
测试详情
--------------------------------------------------------------------------------

1. GET /simple/v1/testhello - 正常参数
   状态: ✅ 通过
   请求: GET /simple/v1/testhello
   参数: {'helloparam': 'test-parameter'}
   响应状态码: 200
   响应体: {"code": 200, "msg": "", "data": "hello"}

...
```

## 失败处理

- 不修改测试文件
- 不修改执行命令
- 基于测试名称、接口语义和 CLI 输出解释失败原因
- 提供可能的解决方案建议

## 快速开始

### 1. 配置环境

编辑 `env.config.json`，填入你的项目信息。

### 2. 列出可用测试

```bash
python ./.trae/skills/api-tests/scripts/list_tests.py
```

### 3. 执行测试

```bash
# 单个测试
python ./.trae/skills/api-tests/scripts/run_cli.py tests/health-check.json test

# 批量执行
python ./.trae/skills/api-tests/scripts/run_cli.py tests/ test parallel
```

## CI/CD 集成

### GitHub Actions

```yaml
name: API Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Setup Python
        uses: actions/setup-python@v4
        with:
          python-version: '3.9'
      - name: Run API tests
        env:
          APIFOX_TOKEN: ${{ secrets.APIFOX_TOKEN }}
        run: python ./.trae/skills/api-tests/scripts/run_cli.py tests/ test
      - name: Upload reports
        uses: actions/upload-artifact@v3
        with:
          name: test-reports
          path: test-reports/
```

### Jenkins

```groovy
pipeline {
    agent any
    
    stages {
        stage('API Test') {
            steps {
                sh 'python ./.trae/skills/api-tests/scripts/run_cli.py tests/ test'
            }
        }
    }
    
    post {
        always {
            junit 'test-reports/*.xml'
        }
    }
}
```

## 相关资源

- [Python urllib 文档](https://docs.python.org/3/library/urllib.html)
- [Python concurrent.futures 文档](https://docs.python.org/3/library/concurrent.futures.html)
- [Claude Skills 文档](https://docs.claude.ai/skills)
- [README.md](./README.md) - 详细使用说明