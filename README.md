# DidaOJ-go

[![CodeQL](https://github.com/hpu-developer/DidaOJ-go/actions/workflows/codeql.yml/badge.svg)](https://github.com/hpu-developer/DidaOJ-go/actions/workflows/codeql.yml) [![License](https://img.shields.io/badge/license-GPLv3-blue.svg)](LICENSE)

DidaOJ-go是一个基于Go语言开发的在线评测系统，提供编程题目的在线评测、竞赛管理、用户系统等功能。

## 技术栈

- **后端语言**: Go
- **项目结构**: 多模块Go工作区(go.work)
- **主要模块**:
  - `web`: Web前端服务
  - `judge`: 评测系统核心
  - `meta`: 通用工具库和中间件
  - `foundation`: 基础框架组件

## 功能特性

### 核心功能
- ✅ 在线代码提交与评测
- ✅ 多种编程语言支持
- ✅ 实时评测结果反馈
- ✅ 题目管理系统
- ✅ 用户认证与授权
- ✅ 竞赛管理系统

### 高级特性
- 📊 实时排行榜
- 📝 详细评测报告
- 🔒 安全的沙箱执行环境
- 📱 响应式设计支持
- ⚡ 高性能评测引擎

## 项目结构

```
DidaOJ-go/
├── .github/          # GitHub配置文件
│   └── workflows/    # CI/CD工作流
├── foundation/       # 基础框架组件
│   ├── error-code/   # 错误码定义
│   ├── foundation-auth/  # 认证模块
│   ├── foundation-config/ # 配置管理
│   ├── foundation-dao/    # 数据访问层
│   └── ...           # 更多基础组件
├── judge/            # 评测系统核心
│   ├── application/  # 应用层逻辑
│   ├── bot-judge/    # 机器人评测模块
│   ├── config/       # 配置文件
│   ├── go-judge/     # Go评测引擎
│   ├── service/      # 业务服务层
│   ├── build.bat     # Windows构建脚本
│   ├── run.sh        # Linux运行脚本
│   └── main.go       # 入口文件
├── meta/             # 通用工具库
│   ├── auth/         # 认证工具
│   ├── controller/   # 通用控制器
│   ├── meta-config/  # 配置工具
│   ├── meta-db/      # 数据库工具
│   └── ...           # 更多工具模块
├── web/              # Web前端服务
│   ├── application/  # 应用层逻辑
│   ├── controller/   # 控制器
│   ├── request/      # 请求定义
│   ├── response/     # 响应定义
│   ├── router/       # 路由配置
│   ├── service/      # 业务服务层
│   ├── build.bat     # Windows构建脚本
│   ├── run.sh        # Linux运行脚本
│   └── main.go       # 入口文件
├── .gitignore        # Git忽略文件
├── .gitmodules       # Git子模块配置
└── go.work           # Go工作区配置
```

## 快速开始

### 环境要求
- Go 1.24+
- Git

### 安装步骤

1. 克隆仓库
```bash
git clone https://github.com/hpu-developer/DidaOJ-go.git
cd DidaOJ-go
```

2. 初始化子模块
```bash
git submodule update --init --recursive
```

3. 构建项目

### Windows本地构建
```bash
# 构建web服务
cd web && go build

# 构建judge服务
cd ../judge && go build
```

### 交叉编译Linux版本
```bash
# 使用提供的构建脚本
web\build.bat
judge\build.bat
```

4. 运行服务

### 直接运行（Windows）
```bash
# 运行web服务
cd web && .\web.exe

# 运行judge服务（新开一个终端）
cd judge && .\judge.exe
```

### 使用运行脚本（Linux）
```bash
# 赋予执行权限
chmod +x web/web judge/judge
chmod +x web/run.sh judge/run.sh

# 运行web服务
cd web && ./run.sh

# 运行judge服务（新开一个终端）
cd judge && ./run.sh
```

### 带配置文件运行
```bash
# 运行web服务并指定配置文件
cd web && .\web.exe --meta-config=meta.yaml --foundation-config=foundation.yaml --config=config.yaml --log-config=log.yaml

# 运行judge服务并指定配置文件
cd judge && .\judge.exe --meta-config=meta.yaml --foundation-config=foundation.yaml --config=config.yaml --log-config=log.yaml
```

## 配置说明

各模块的配置文件位于对应模块的根目录下，支持YAML格式配置：
- `meta.yaml`: 基础配置
- `foundation.yaml`: 框架组件配置
- `config.yaml`: 业务模块配置
- `log.yaml`: 日志配置

主要配置项：
- 数据库连接信息
- Redis配置
- 评测引擎参数
- Web服务端口和地址
- 日志级别和输出路径
- 认证相关配置

配置文件可以通过命令行参数指定，如：

```bash
./web --config=my-config.yaml
```

## 开发指南

### 代码规范
- 遵循Go语言官方代码规范
- 使用`gofmt`格式化代码
- 编写单元测试

### 提交代码
1. 创建功能分支
2. 提交代码并编写清晰的提交信息
3. 提交Pull Request进行代码审查

## 安全

本项目使用CodeQL进行代码安全扫描，确保代码质量和安全性。

[![CodeQL](https://github.com/hpu-developer/DidaOJ-go/actions/workflows/codeql.yml/badge.svg)](https://github.com/hpu-developer/DidaOJ-go/actions/workflows/codeql.yml)

## 许可证

本项目采用GNU General Public License v3.0 (GPLv3)许可证，详情请查看[LICENSE](LICENSE)文件。

GPLv3是一种copyleft许可证，意味着如果您修改并分发本项目的代码，必须也以GPLv3许可证开放源代码。这确保了项目的自由和开源性质得以延续。

## 贡献

欢迎提交Issue和Pull Request来帮助改进项目！

## 联系方式

如有问题或建议，请通过以下方式联系：
- GitHub Issues: [https://github.com/hpu-developer/DidaOJ-go/issues](https://github.com/hpu-developer/DidaOJ-go/issues)
- 邮箱: 请在项目中查找联系方式

---

<div align="center">
  <strong>DiDaOJ - 让编程评测更简单</strong>
</div>
