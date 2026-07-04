---
layout: home

hero:
  name: ipapi
  text: 命令行 IP 地理位置查询工具
  tagline: 🤖 默认 JSON 输出、稳定退出码、AI Agent 友好 —— 一行命令查 IP，背后是零依赖的 Go SDK
  image:
    src: /favicon.svg
    alt: ipapi CLI
  actions:
    - theme: brand
      text: ⚡ 快速开始
      link: /cli/quickstart
    - theme: alt
      text: 📥 安装
      link: /cli/install
    - theme: alt
      text: 🤖 Agent 接入
      link: /cli/agent
    - theme: alt
      text: ⭐ GitHub
      link: https://github.com/cyberspacesec/ipapi.co-skills

features:
  - icon: 🤖
    title: AI Agent 友好
    details: 默认输出 JSON 信封 {ok, command, data, meta}，错误码稳定可枚举，Agent 无需读文档即可调用。
    link: /cli/agent
    linkText: Agent 接入指南 →
  - icon: 🚪
    title: 稳定退出码
    details: 0 成功，3-12 对应 10 类业务错误，6/8/9 标记可重试。脚本用 $? 即可分支。
    link: /cli/exit-codes
    linkText: 退出码表 →
  - icon: 🎯
    title: 字段级查询
    details: 只想要国家或 ASN？ipapi field 8.8.8.8 country 一行搞定，--human 直出纯值便于 pipe。
    link: /cli/command-field
    linkText: field 命令 →
  - icon: 📡
    title: 多格式原始响应
    details: JSON / JSONP / XML / CSV / YAML 五种格式，raw 命令直出原始字节，喂给 jq、column 等工具。
    link: /cli/command-raw
    linkText: raw 命令 →
  - icon: 🧭
    title: 自描述字段
    details: ipapi fields 列出全部 28 个可查字段并按语义分组，Agent 先探查再查询。
    link: /cli/command-fields
    linkText: fields 命令 →
  - icon: 🔧
    title: 灵活配置
    details: 旗标 > 环境变量 > ~/.ipapi.json > 默认值，四级配置覆盖，适配 CI、容器、本地。
    link: /cli/config
    linkText: 配置方式 →
  - icon: 🧩
    title: 也是 Go SDK
    details: CLI 只是薄壳，背后是 pkg/ipapi —— 一个零运行时依赖的 Go 库，可直接嵌入你的程序。
    link: /cli/sdk-bridge
    linkText: CLI 与 SDK →
  - icon: 🛡
    title: 结构化错误
    details: 10 种哨兵错误 + errors.Is 精准匹配，APIError 携带原因、IP、保留标记等上下文。
    link: /api/errors
    linkText: 错误参考 →
---

<script setup>
</script>

# 🌐 一图看懂 ipapi.co-skills 全栈架构

从 AI Agent / 终端用户发起命令，到 SDK 调用 ipapi.co API 返回结构化数据，整条链路一目了然。

```mermaid
flowchart TD
    subgraph 调用方["👨‍💻 调用方"]
        Agent["🤖 AI Agent<br/>读 JSON 信封 + 退出码"]
        Human["🧑‍💻 终端用户<br/>--human 人类可读"]
        Script["📜 Shell 脚本<br/>$? 分支 + pipe"]
    end

    subgraph CLI层["🖥️ CLI 层（cmd/ipapi · cobra）"]
        Root["ipapi 根命令<br/>PersistentPreRunE 加载配置"]
        Sub["9 个子命令<br/>info / me / field / raw / fields / version..."]
        Envelope["JSON 信封渲染<br/>{ok,command,data|error,meta}"]
        Exit["退出码映射<br/>0/2-12/70 · retryable 标记"]
    end

    subgraph SDK层["🧩 SDK 层（pkg/ipapi · 零依赖）"]
        Client["ipapi.Client<br/>BaseURL/APIKey/Retries/RateLimiter"]
        Methods["6 个查询方法<br/>GetIPInfo / GetField / GetRaw..."]
        DoReq["doRequest 内核<br/>限流 → 重试 → 状态码映射 → APIError 解析"]
        Errors["10 个哨兵错误<br/>errors.Is 精准匹配"]
    end

    subgraph 服务端["🌐 ipapi.co API"]
        API["GET /{ip}/{format}/<br/>JSON/JSONP/XML/CSV/YAML"]
    end

    Agent --> Root
    Human --> Root
    Script --> Root
    Root --> Sub --> Envelope
    Envelope --> Exit
    Sub -.调用.-> Methods
    Methods --> DoReq
    DoReq -.错误.-> Errors
    DoReq -->|HTTPS| API
    API -->|响应| DoReq
    DoReq -->|*IPInfo / []byte / string| Methods
    Methods --> Envelope

    classDef call fill:#e0f2fe,stroke:#0284c7
    classDef cli fill:#fef3c7,stroke:#d97706
    classDef sdk fill:#dcfce7,stroke:#16a34a
    classDef api fill:#fce7f3,stroke:#db2777
    class Agent,Human,Script call
    class Root,Sub,Envelope,Exit cli
    class Client,Methods,DoReq,Errors sdk
    class API api
```

::: tip 🚀 30 秒上手
```bash
# 安装
go install github.com/cyberspacesec/ipapi.co-skills/cmd/ipapi@latest

# 查询（默认 JSON 信封，Agent 友好）
ipapi info 8.8.8.8

# 人类可读
ipapi info 8.8.8.8 -H

# 单字段（shell pipe 友好）
ipapi field 8.8.8.8 country
```
:::

::: info 📊 数据流：一次 `ipapi info 8.8.8.8` 的旅程
```mermaid
sequenceDiagram
    autonumber
    participant User as 🧑‍💻 用户
    participant CLI as 🖥️ ipapi CLI
    participant SDK as 🧩 ipapi.Client
    participant API as 🌐 ipapi.co

    User->>CLI: ipapi info 8.8.8.8
    CLI->>CLI: 加载配置 (旗标>env>~/.ipapi.json>默认)
    CLI->>SDK: GetIPInfo(ctx, "8.8.8.8", "json")
    SDK->>SDK: ValidateIP + ValidateFormat
    SDK->>SDK: applyAuth(Bearer) + setHeaders
    SDK->>API: GET /8.8.8.8/json/
    API-->>SDK: 200 JSON 28 字段
    SDK->>SDK: json.Decode → *IPInfo
    SDK-->>CLI: *IPInfo, nil
    CLI->>CLI: 渲染 JSON 信封 {ok,data,meta}
    CLI-->>User: stdout: 信封 JSON (退出码 0)
```
:::

::: details 🔍 28 个可查字段一览（ipapi fields 自描述）
按语义分 7 组，`ipapi fields --group geo` 只看地理组。

| 分组 | 字段 | 示例值 |
| --- | --- | --- |
| 🆔 identity | ip · network · version · hostname | 8.8.8.8 |
| 🌍 geo | city · region · country · country_name · country_code · country_code_iso3 · country_capital · country_tld · continent_code · in_eu · postal · latitude · longitude · latlong · region_code | Mountain View |
| ⏰ time | timezone · utc_offset | America/Los_Angeles |
| 📡 network | asn · org | AS15169 (Google LLC) |
| 🗣 culture | languages · country_calling_code | en |
| 💰 economy | currency · currency_name | USD |
| 📊 stats | country_area · country_population | 9833520 |
:::

::: tip 💡 接下来
- 🚀 [快速开始](/cli/quickstart) — 5 分钟跑通第一次查询
- 🤖 [Agent 接入指南](/cli/agent) — 把 ipapi 接进你的 AI 工作流
- 📖 [CLI 命令速查](/cli/commands) — 9 个子命令一览
- 🧩 [SDK 与 CLI 关系](/cli/sdk-bridge) — 何时用 CLI，何时嵌入 SDK
:::
