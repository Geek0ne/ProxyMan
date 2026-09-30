# ProxyMan

> **One binary to rule them all — manage Xray and Mihomo, every protocol, one command.**
> 一个二进制，管好双引擎。全协议，一条命令。
>
> **当前版本 v1.1.0** — 新增节点库、测速、日志与 dry-run 预演

[![Go](https://img.shields.io/badge/Go-1.26%2B-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-FF6B35?style=for-the-badge&logo=opensourceinitiative&logoColor=white)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Linux-F7DF1E?style=for-the-badge&logo=linux&logoColor=white)](https://github.com)

## 📖 简介

**ProxyMan** 是一个面向 Linux 的统一代理命令行工具，用单个静态二进制同时管理 **Xray-core** 与 **Mihomo (Clash Meta)** 两大内核。它覆盖 VMess、VLESS、Reality、gRPC、Trojan、Shadowsocks 2022、AnyTLS、TUIC、Hysteria2 等主流协议，内置订阅链接解析与 GeoIP/GeoSite 智能分流，并提供系统代理开关与 systemd 服务集成——安装、配置、启动、开机自启，全部在终端内完成。

> 统管 Xray 与 Mihomo 双内核的全协议 Linux 代理 CLI。

### ✨ 核心特性

- **全协议支持**
  - V2Ray / Xray 系列
    - VMess
    - VLESS
    - VLESS + Reality
    - VLESS + gRPC
    - Trojan
    - Trojan + gRPC
    - Shadowsocks
    - Shadowsocks-2022
    - AnyTLS
    - SOCKS5
    - HTTP
  - Clash / Mihomo 系列
    - VMess
    - VLESS
    - Trojan
    - Shadowsocks
    - HTTP / HTTPS
    - SOCKS5
    - TUIC
    - Hysteria
    - Hysteria2
    - AnyTLS

- **智能分流**
  - 🇨🇳 国内直连：GEOIP/CN 规则自动直连
  - 🌍 国际代理：海外流量自动走代理
  - 🛡️ 广告拦截：内置广告过滤规则

- **一键管理**
  - 自动下载官方二进制（支持 GitHub 镜像加速）
  - SHA256 校验确保安全
  - systemd 服务集成
  - 配置文件模板自动生成

## 📦 安装

### 从源码编译

```bash
# 克隆仓库
git clone https://github.com/geek0ne/ProxyMan.git
cd ProxyMan

# 编译
go build -o ProxyMan .

# 安装到系统路径（可选）
sudo cp ProxyMan /usr/local/bin/
```

### 直接下载

```bash
# 下载预编译二进制（Linux amd64）
wget https://github.com/geek0ne/ProxyMan/releases/latest/download/ProxyMan-linux-amd64.tar.gz
tar -xzf ProxyMan-linux-amd64.tar.gz
chmod +x ProxyMan
sudo cp ProxyMan /usr/local/bin/
```

## 🚀 快速开始

### 1. 安装引擎

```bash
# 安装 Xray-core (V2Ray)
ProxyMan install v2ray

# 安装 Mihomo (Clash Meta)
ProxyMan install clash
```

### 2. 配置代理服务器

编辑配置文件，填入你的代理服务器信息：

```bash
# 编辑 Xray 配置
ProxyMan config edit v2ray
# 或直接编辑配置文件
vim ~/.config/ProxyMan/engines/v2ray/config.json

# 编辑 Mihomo 配置
ProxyMan config edit clash
# 或直接编辑配置文件
vim ~/.config/ProxyMan/engines/clash/config.yaml
```

### 3. 启动代理

```bash
# 启动 Xray
ProxyMan start v2ray ~/.config/ProxyMan/engines/v2ray/config.json

# 启动 Mihomo
ProxyMan start clash ~/.config/ProxyMan/engines/clash/config.yaml
```

### 4. 开启系统代理

```bash
ProxyMan system-proxy enable
```

### 5. 安装为 systemd 服务（可选）

```bash
# 安装 Xray 服务
sudo cp ~/.config/ProxyMan/engines/v2ray/ProxyMan-xray.service /etc/systemd/system/
sudo systemctl enable --now ProxyMan-xray

# 安装 Mihomo 服务
sudo cp ~/.config/ProxyMan/engines/clash/ProxyMan-mihomo.service /etc/systemd/system/
sudo systemctl enable --now ProxyMan-mihomo
```

## 📚 命令参考

### 节点管理

| 命令 | 说明 |
|------|------|
| `ProxyMan import <link> --store` | 导入节点并保存到本地节点库 |
| `ProxyMan list` | 列出已保存的节点 |
| `ProxyMan remove <name>` | 删除指定节点 |
| `ProxyMan probe [name...]` | 测速（TCP 握手延迟），结果按速度排序 |
| `ProxyMan probe --timeout 10s --jobs 16` | 自定义超时与并发 |
| `ProxyMan apply <engine> [mode]` | 把节点库写入引擎配置 |

`apply` 的 `mode` 可选 `full`（全流量走代理）或 `china-split`（国内直连、国际代理）。

### 引擎管理

| 命令 | 说明 |
|------|------|
| `ProxyMan install <v2ray\|clash>` | 安装引擎 |
| `ProxyMan uninstall <v2ray\|clash>` | 卸载引擎 |
| `ProxyMan start <engine> <config>` | 启动引擎 |
| `ProxyMan stop <engine>` | 停止引擎 |
| `ProxyMan status` | 查看引擎状态 |

### 配置管理

| 命令 | 说明 |
|------|------|
| `ProxyMan config show` | 查看当前配置 |
| `ProxyMan config set <key> <value>` | 设置配置项 |
| `ProxyMan config edit <engine>` | 编辑配置文件 |

### 流量分流

| 命令 | 说明 |
|------|------|
| `ProxyMan split <engine> full` | 全代理模式 |
| `ProxyMan split <engine> china-split` | 国内直连模式 |

### 系统代理

| 命令 | 说明 |
|------|------|
| `ProxyMan system-proxy enable` | 开启系统代理 |
| `ProxyMan system-proxy disable` | 关闭系统代理 |
| `ProxyMan system-proxy status` | 查看代理状态 |

### 测试

| 命令 | 说明 |
|------|------|
| `ProxyMan test <engine>` | 测试引擎配置 |

### 日志

| 命令 | 说明 |
|------|------|
| `ProxyMan log <engine>` | 查看引擎日志 |
| `ProxyMan log <engine> -f` | 实时跟踪日志输出 |

### 全局选项

| 选项 | 说明 |
|------|------|
| `--dry-run` | 预演所有写操作，不实际落盘 |
| `--config <path>` | 指定配置文件路径 |

## 📁 配置文件

### 目录结构

```
~/.config/ProxyMan/
├── engines/
│   ├── v2ray/
│   │   ├── xray                    # Xray 二进制
│   │   ├── config.json             # Xray 配置
│   │   ├── geoip.dat               # GeoIP 数据库
│   │   ├── geosite.dat             # GeoSite 数据库
│   │   └── ProxyMan-xray.service  # systemd 服务
│   └── clash/
│       ├── mihomo                  # Mihomo 二进制
│       ├── config.yaml             # Mihomo 配置
│       └── ProxyMan-mihomo.service # systemd 服务
```

### Xray 配置示例 (config.json)

```json
{
  "log": {"loglevel": "info"},
  "inbounds": [
    {"tag":"http-in","port":10809,"protocol":"http","listen":"127.0.0.1"},
    {"tag":"socks-in","port":10808,"protocol":"socks","listen":"127.0.0.1","settings":{"udp":true}}
  ],
  "outbounds": [
    {
      "tag":"proxy","protocol":"vmess",
      "settings":{"vnext":[{"address":"YOUR_SERVER","port":443,"users":[{"id":"YOUR_UUID"}]}]},
      "streamSettings":{"network":"ws","security":"tls","tlsSettings":{"serverName":"YOUR_SNI"}}
    },
    {"tag":"direct","protocol":"freedom"},
    {"tag":"block","protocol":"blackhole"}
  ],
  "routing":{"domainStrategy":"IPIfNonMatch","rules":[
    {"type":"field","ip":["geoip:cn","geoip:private"],"outboundTag":"direct"},
    {"type":"field","domain":["geosite:cn"],"outboundTag":"direct"},
    {"type":"field","domain":["geosite:geolocation-!cn"],"outboundTag":"proxy"}
  ]}
}
```

### Mihomo 配置示例 (config.yaml)

```yaml
mixed-port: 7890
allow-lan: false
mode: rule
log-level: info

dns:
  enable: true
  enhanced-mode: fake-ip
  nameserver:
    - 223.5.5.5
    - 119.29.29.29
  fallback:
    - https://dns.google/dns-query
  fallback-filter:
    geoip: true
    geoip-code: CN

proxies:
  - name: "proxy-node"
    type: vmess
    server: YOUR_SERVER_ADDRESS
    port: 443
    uuid: YOUR_UUID
    alterId: 0
    cipher: auto
    tls: true
    network: ws

proxy-groups:
  - name: "Proxy"
    type: select
    proxies:
      - proxy-node
      - DIRECT

rules:
  - GEOSITE,cn,DIRECT
  - GEOIP,cn,DIRECT,no-resolve
  - GEOIP,private,DIRECT,no-resolve
  - MATCH,Proxy
```

## 🔧 高级功能

### 自动分流配置

```bash
# 国内直连模式（推荐）
ProxyMan split clash china-split

# 全代理模式
ProxyMan split clash full
```

### 分流规则说明

| 规则 | 说明 |
|------|------|
| `GEOSITE,cn,DIRECT` | 国内域名直连 |
| `GEOIP,cn,DIRECT` | 国内 IP 直连 |
| `GEOIP,private,DIRECT` | 私有网络直连 |
| `GEOSITE,geolocation-!cn,Proxy` | 国际网站走代理 |
| `MATCH,Proxy` | 兜底规则走代理 |

### 支持的协议

#### V2Ray/Xray 协议

| 协议 | 说明 |
|------|------|
| VMess | V2Ray 主力协议 |
| VLess | 新一代轻量协议 |
| Trojan | 伪装 HTTPS 协议 |
| Shadowsocks | 经典加密代理 |
| SOCKS5 | SOCKS5 代理 |
| HTTP | HTTP 代理 |

#### Clash/Mihomo 协议

| 协议 | 说明 |
|------|------|
| VMess | V2Ray 主力协议 |
| VLess | 新一代轻量协议 |
| Trojan | 伪装 HTTPS 协议 |
| Shadowsocks | 经典加密代理 |
| HTTP/HTTPS | HTTP/HTTPS 代理 |
| SOCKS5 | SOCKS5 代理 |
| Tuic | 基于 QUIC 的协议 |
| Hysteria2 | 高性能 UDP 协议 |

## 🛡️ 安全特性

- **SHA256 校验**：所有下载的二进制都经过哈希校验
- **官方源下载**：从 GitHub 官方 release 下载
- **镜像加速**：支持 ghfast.top 镜像加速
- **配置隔离**：每个引擎独立配置目录

## 🐛 故障排除

### 下载失败

```bash
# 使用镜像加速（自动）
ProxyMan install v2ray

# 或手动设置代理
export https_proxy=http://127.0.0.1:7890
ProxyMan install v2ray
```

### 端口冲突

```bash
# 检查端口占用
ss -tlnp | grep -E '10808|10809|7890'

# 修改配置文件中的端口
ProxyMan config edit v2ray
```

### 权限问题

```bash
# 确保二进制有执行权限
chmod +x ~/.config/ProxyMan/engines/v2ray/xray
chmod +x ~/.config/ProxyMan/engines/clash/mihomo
```

## 📊 系统要求

- **操作系统**：Linux (amd64)
- **Go 版本**：1.26+ (编译时)
- **磁盘空间**：约 150MB (两个引擎)
- **依赖**：无 (静态编译)

## 🤝 贡献

欢迎贡献代码！请遵循以下步骤：

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🙏 致谢

- [Xray-core](https://github.com/XTLS/Xray-core) - V2Ray 核心
- [Mihomo](https://github.com/MetaCubeX/mihomo) - Clash Meta 核心
- [GeoIP](https://github.com/Loyalsoldier/geoip) - IP 地理位置数据库
- [GeoSite](https://github.com/v2fly/domain-list-community) - 域名分类数据库

## 📞 联系方式

- Issues：[GitHub Issues](https://github.com/geek0ne/ProxyMan/issues)
- Email：nzl9100@gmail.com

---

**如果这个项目对你有帮助，请给个 ⭐ Star 支持一下！**
