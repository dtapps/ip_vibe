# vibe_ip

一个用于管理CDN、DNS和IP配置的工具，支持自动检查和更新IP地址。

## 功能特性

- 自动检查公网IP地址变化
- 支持更新CDN配置
- 支持更新DNS记录
- 配置管理功能
- 服务安装、启动、停止和卸载
- 多平台支持（Linux、Windows、macOS）

## 安装方法

### 使用安装脚本（推荐）

```bash
# 下载并执行安装脚本
curl https://cnb.cool/dtapp/vibe_ip/-/git/raw/master/install.sh -sSfL | sh
# 查看帮助信息
curl https://cnb.cool/dtapp/vibe_ip/-/git/raw/master/install.sh -sSfL | sh -s -- --help   
```

### 配置示例

```yaml

```

## 使用说明

### 基本命令

```bash
# 查看帮助信息
vibe_ip --help

# 检查IP
vibe_ip ip check

# 更新DNS记录
vibe_ip dns update

# 更新CDN配置
vibe_ip cdn update

# 运行服务
vibe_ip run
```

### 服务管理

```bash
# 安装为系统服务
vibe_ip service install

# 启动服务
vibe_ip service start

# 停止服务
vibe_ip service stop

# 重启服务
vibe_ip service restart

# 卸载服务
vibe_ip service uninstall
```

### 配置管理

```bash
# 初始化配置
vibe_ip config init

# 显示当前配置
vibe_ip config show
```

## 许可证

[MIT License](LICENSE)