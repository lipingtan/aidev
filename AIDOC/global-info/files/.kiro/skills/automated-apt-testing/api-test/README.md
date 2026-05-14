# Api Tests Skill

使用 Python 编写的 Api 自动化测试 Skill，支持测试用例管理、环境配置、批量执行和报告生成。

## � 系统要求

### Python 版本
- **最低版本：** Python 3.4+
- **推荐版本：** Python 3.7+

### 依赖说明
- ✅ **无外部依赖** - 所有功能使用 Python 标准库
- ✅ **开箱即用** - 无需安装任何第三方包
- ✅ **跨平台兼容** - 支持 Windows、Linux、macOS

### 使用的 Python 标准库

| 模块 | 版本要求 | 用途 |
|--------|-----------|------|
| `os` | 所有版本 | 文件和路径操作 |
| `json` | 所有版本 | JSON 数据处理 |
| `sys` | 所有版本 | 系统参数和退出 |
| `pathlib` | Python 3.4+ | 面向对象路径操作 |
| `urllib.parse` | 所有版本 | URL 解析和编码 |
| `urllib.request` | 所有版本 | HTTP 请求 |
| `urllib.error` | 所有版本 | URL 错误处理 |
| `concurrent.futures` | Python 3.2+ | 并行执行 |

## �📁 目录结构

```
api-tests/
├── SKILL.md                    # Skill 说明文档
├── env.config.json             # 环境配置文件
├── scripts/
│   ├── list_tests.py           # 列出可用测试
│   └── run_cli.py             # 执行测试脚本
└── tests/
    └──  health-check.json        # 健康检查测试
```

## 🚀 快速开始

### 1. 配置环境

编辑 `env.config.json`，填入你的项目信息：

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
  }
}
```

### 2. 生成测试 JSON 文件

根据项目中的接口代码，调用大模型（如 Claude）生成测试 JSON 文件。

**步骤：**

1. **准备接口信息**
   - 查看项目中的 Controller 源码（如 `SimpleController.java`）
   - 确认接口路径、请求方法、参数、返回值

2. **提供接口信息给大模型**
   - 复制 Controller 代码片段
   - 说明需要测试的接口
   - 提供参考模板（如 `health-check.json`）

3. **生成测试文件**
   - 大模型根据接口信息生成完整的测试用例
   - 包含正常场景、边界场景、异常场景
   - 保存到 `tests/` 目录

**示例：**

```
用户输入：
  - Controller 代码：SimpleController.java
  - 需要测试的接口：/simple/v1/testhello, /simple/v1/testbody
  - 参考模板：health-check.json

大模型输出：
  - 生成 simple-controller-test.json
  - 包含 9 个测试用例（正常、空参数、特殊字符、长参数等）
```

**测试文件结构：**
```json
{
  "name": "测试套件名称",
  "description": "测试描述",
  "baseUrl": "{{baseUrl}}",
  "tests": [
    {
      "name": "测试用例名称",
      "request": {
        "method": "GET/POST",
        "url": "/接口路径",
        "params": {},  // GET 参数
        "headers": {},  // 请求头
        "body": {}  // POST 请求体
      },
      "expect": {
        "statusCode": 200,
        "body": {}  // 预期响应体
      }
    }
  ]
}
```

### 3. 执行测试

```bash
# 单个测试
python scripts/run_cli.py tests/health-check.json test

# 批量执行（顺序）
python scripts/run_cli.py tests/ test

# 批量执行（并行）
python scripts/run_cli.py tests/ test parallel
```

## 📋 命令说明

### run_cli.py

执行指定的测试用例。

**语法：**
```bash
python scripts/run_cli.py <测试文件路径> <环境名> [parallel]
```

**参数：**
- `测试文件路径` - 测试文件的相对或绝对路径
- `环境名` - 环境名称（dev/test/staging/prod）
- `parallel` - 可选，启用并行执行

**示例：**
```bash
# 执行单个测试
python scripts/run_cli.py tests/simple-controller-test.json dev

# 批量执行
python scripts/run_cli.py tests/ dev

# 并行执行
python scripts/run_cli.py tests/ dev parallel
```

### list_tests.py（可选）

列出所有可用的测试文件，用于查看已生成的测试。

```bash
python scripts/list_tests.py
```

**输出示例：**
```
可用的测试文件:

1. health-check.json (0.45 KB)
2. simple-controller-test.json (1.23 KB)
```

## 🔧 环境配置

支持的环境：

| 环境 | 说明 | 配置示例 |
|------|------|---------|
| `dev` | 开发环境 | `http://localhost:8080` |
| `test` | 测试环境 | `https://api.test.example.com` |
| `staging` | 预发布环境 | `https://api.staging.example.com` |
| `prod` | 生产环境 | `https://api.example.com` |

## 📊 测试报告

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

## 📝 测试文件格式

测试文件使用 JSON 格式：

```json
{
  "name": "测试套件名称",
  "description": "测试套件描述",
  "baseUrl": "{{baseUrl}}",
  "tests": [
    {
      "name": "测试用例名称",
      "request": {
        "method": "GET",
        "url": "/api/endpoint"
      },
      "expect": {
        "statusCode": 200,
        "body": {
          "field": "expected-value"
        }
      }
    }
  ]
}
```

## 🔄 CI/CD 集成

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

## 🛠️ 故障排除

### 常见问题

**Q: 找不到测试文件**
```
错误: 测试文件不存在: tests/xxx.json
```
解决：检查文件路径是否正确，使用 `node scripts/list-tests.js` 查看可用测试

**Q: 环境配置错误**
```
错误: 环境 test 未在配置文件中定义
```
解决：在 `env.config.json` 中添加对应的环境配置

**Q: API 测试执行失败**
```
错误: HTTP 请求失败
```
解决：确保目标服务正在运行，检查网络连接和防火墙设置

## 📚 相关资源

- [Python urllib 文档](https://docs.python.org/3/library/urllib.html)
- [Python concurrent.futures 文档](https://docs.python.org/3/library/concurrent.futures.html)
- [Claude Skills 文档](https://docs.claude.ai/skills)

## 📄 License

MIT License
