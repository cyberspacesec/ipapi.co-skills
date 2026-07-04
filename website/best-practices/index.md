# ✅ 最佳实践

> 生产环境使用 ipapi.co-skills 的建议。

::: tip 🎨 一图抵千言
下面是最佳实践的全景图，按「客户端管理 → 可靠性 → 错误可观测 → 工程化」四大维度组织。
:::

```mermaid
flowchart TD
    A["🚀 生产使用 ipapi.co-skills"] --> B["🏗 客户端管理"]
    A --> C["⏱ 可靠性"]
    A --> D["🛡 错误与可观测"]
    A --> E["🧪 工程化"]

    B --> B1["客户端生命周期 全局单例"]
    B --> B2["性能优化 连接池复用"]

    C --> C1["超时策略 分层"]
    C --> C2["重试策略 指数退避"]
    C --> C3["限流策略 RateLimiter"]
    C --> C4["优雅降级 默认值兜底"]

    D --> D1["错误处理 分级"]
    D --> D2["可观测性 日志 metrics"]
    D --> D3["安全实践 校验 Key HTTPS"]
    D --> D4["密钥管理 环境变量"]

    E --> E1["测试实践 httptest"]
    E --> E2["本地化实践 地理"]

    B1 & B2 --> F["✅ 稳定客户端"]
    C1 & C2 & C3 & C4 --> G["✅ 高可用"]
    D1 & D2 & D3 & D4 --> H["✅ 可控可观测"]
    E1 & E2 --> I["✅ 可维护"]

    F & G & H & I --> J["🎯 生产就绪"]
```

::: info 📊 实践成熟度阶梯
从「能跑」到「生产就绪」，分四级阶梯——你处在哪一级？
:::

```mermaid
flowchart TD
    L1["🥉 入门级<br/>能发起查询"] --> L2["🥈 可用级<br/>超时+错误分流"]
    L2 --> L3["🥇 健壮级<br/>重试+限流+降级"]
    L3 --> L4["💎 生产级<br/>可观测+密钥管理+测试"]

    L1 -.->|必备| R1["NewClient + GetIPInfo"]
    L2 -.->|必备| R2["context.WithTimeout + errors.Is"]
    L3 -.->|必备| R3["Retries + RateLimiter + 优雅降级"]
    L4 -.->|必备| R4["metrics + 环境变量 Key + httptest"]

    classDef l1 fill:#fef9c3,stroke:#ca8a04
    classDef l2 fill:#fed7aa,stroke:#ea580c
    classDef l3 fill:#fef3c7,stroke:#d97706
    classDef l4 fill:#dcfce7,stroke:#16a34a
    class L1 l1
    class L2 l2
    class L3 l3
    class L4 l4
```

## 🏗 客户端管理

::: info 📦 维度说明
客户端是所有调用的入口，管理好生命周期与连接复用，是性能与稳定性的基础。
:::

- [客户端生命周期管理](./client-lifecycle) — 全局单例复用
- [性能优化](./performance) — 连接池与复用

## ⏱ 可靠性

::: warning 🛡 维度说明
可靠性四件套：超时控制上限、重试覆盖瞬时错误、限流防过载、降级保主流程。
:::

| 实践 | 一句话 | 解决的问题 |
|------|--------|------------|
| [超时策略](./timeout-strategy) | 分层超时 | 调用卡死拖垮链路 |
| [重试策略](./retry-strategy) | 内置 + 指数退避 | 瞬时网络错误 |
| [限流策略](./rate-limit-strategy) | RateLimiter 通道 | 突发流量打爆配额 |
| [优雅降级](./graceful-degradation) | 失败用默认值 | 弱依赖拖垮主流程 |

## 🛡 错误与可观测

::: danger 🔒 维度说明
错误与可观测是生产的「眼睛」与「保险丝」：分级处理避免静默失效，可观测让你看见降级率，安全与密钥管理守住底线。
:::

- [错误处理策略](./error-handling-strategy) — 分级处理
- [可观测性](./observability) — 日志与 metrics
- [安全实践](./security) — 校验、Key、HTTPS
- [密钥管理](./secret-management) — 环境变量

## 🧪 工程化

::: tip 🧪 维度说明
工程化保证代码可测、可演进：测试实践用 httptest 模拟远端，本地化实践按地理定制体验。
:::

- [测试实践](./testing) — httptest 模拟
- [本地化实践](./localization) — 按地理本地化

## 🚀 下一步

::: details 🧭 不知从哪开始？按场景选
| 你的场景 | 推荐先读 |
|----------|----------|
| 第一次接入 | [客户端生命周期](./client-lifecycle) → [超时策略](./timeout-strategy) |
| 线上偶发超时 | [重试策略](./retry-strategy) → [优雅降级](./graceful-degradation) |
| 想看降级率 | [可观测性](./observability) → [错误处理策略](./error-handling-strategy) |
| 担心 Key 泄露 | [密钥管理](./secret-management) → [安全实践](./security) |
| 要做国际化 | [本地化实践](./localization) |
| 想写单测 | [测试实践](./testing) |
:::

- 📖 看 [指南](../guide/intro)
- ❓ 看 [FAQ](../faq/
- 🍳 看 [Cookbook](../cookbook/
