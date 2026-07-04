import { defineConfig } from 'vitepress'
import { withMermaid } from 'vitepress-plugin-mermaid'

// https://vitepress.dev/reference/site-config
export default withMermaid(defineConfig({
  lang: 'zh-CN',
  title: 'ipapi.co-skills',
  description: '命令行 IP 地理位置查询工具（ipapi CLI），默认 JSON 输出、Agent 友好；背后是零依赖的 Go SDK',
  lastUpdated: true,
  cleanUrls: true,

  // GitHub Pages 项目站点，路径前缀为仓库名
  // 部署到自定义域名或用户名.github.io 时改为 '/'
  base: '/ipapi.co-skills/',

  head: [
    ['link', { rel: 'icon', type: 'image/svg+xml', href: '/favicon.svg' }],
    ['meta', { name: 'theme-color', content: '#3c8c4a' }],
    ['meta', { name: 'og:title', content: 'ipapi.co-skills 文档' }],
    ['meta', { name: 'og:description', content: 'Go 语言的 IP 地理位置查询 SDK' }]
  ],

  themeConfig: {
    // 站点 logo 与标题
    logo: '/favicon.svg',

    // 顶部导航
    nav: [
      { text: '🖥 CLI', link: '/cli/', activeMatch: '/cli/' },
      { text: '⚡ 快速开始', link: '/cli/quickstart', activeMatch: '/cli/quickstart' },
      { text: '📖 库指南', link: '/guide/intro', activeMatch: '/guide/' },
      { text: '🎓 教程', link: '/tutorials/', activeMatch: '/tutorials/' },
      {
        text: '📚 API 参考',
        items: [
          { text: 'Client 客户端', link: '/api/client' },
          { text: 'API 方法', link: '/api/methods' },
          { text: '数据模型', link: '/api/models' },
          { text: '错误类型', link: '/api/errors' },
          { text: '选项函数', link: '/api/options' }
        ]
      },
      { text: '🧪 示例', link: '/examples/basic-usage', activeMatch: '/examples/' },
      {
        text: '🍳 实战',
        items: [
          { text: 'Cookbook 食谱', link: '/cookbook/' },
          { text: '最佳实践', link: '/best-practices/' },
          { text: 'FAQ', link: '/faq/' }
        ]
      },
      {
        text: '📚 深度参考',
        items: [
          { text: '错误详解', link: '/reference/errors/err-invalid-ip' },
          { text: '字段详解', link: '/reference/fields/ip' },
          { text: '常量与内部', link: '/reference/internal/format-json' },
          { text: '参考总览', link: '/reference/' }
        ]
      },
      { text: '⚙️ 配置', link: '/config/cicd', activeMatch: '/config/' },
      {
        text: '🔗 资源',
        items: [
          { text: 'GitHub 仓库', link: 'https://github.com/cyberspacesec/ipapi.co-skills' },
          { text: 'ipapi.co 官网', link: 'https://ipapi.co' },
          { text: 'Go 文档', link: 'https://pkg.go.dev/github.com/cyberspacesec/ipapi.co-skills' }
        ]
      }
    ],

    // 侧边栏
    sidebar: {
      '/cli/': [
        {
          text: '🚀 开始',
          collapsed: false,
          items: [
            { text: 'CLI 总览', link: '/cli/' },
            { text: '安装', link: '/cli/install' },
            { text: '快速开始', link: '/cli/quickstart' },
            { text: '命令速查', link: '/cli/commands' }
          ]
        },
        {
          text: '📋 命令详解',
          collapsed: false,
          items: [
            { text: 'info / me', link: '/cli/command-info' },
            { text: 'field / me-field', link: '/cli/command-field' },
            { text: 'raw / me-raw', link: '/cli/command-raw' },
            { text: 'fields', link: '/cli/command-fields' },
            { text: 'version / completion', link: '/cli/command-version' }
          ]
        },
        {
          text: '🔧 参考',
          collapsed: false,
          items: [
            { text: '全局旗标', link: '/cli/flags' },
            { text: '输出格式', link: '/cli/output' },
            { text: '退出码', link: '/cli/exit-codes' },
            { text: '配置', link: '/cli/config' }
          ]
        },
        {
          text: '🤖 进阶',
          collapsed: false,
          items: [
            { text: 'Agent 接入', link: '/cli/agent' },
            { text: 'CLI 与 SDK', link: '/cli/sdk-bridge' },
            { text: '实战食谱', link: '/cli/cookbook' }
          ]
        }
      ],

      '/guide/': [
        {
          text: '📍 入门',
          collapsed: false,
          items: [
            { text: '什么是 ipapi.co-skills', link: '/guide/intro' },
            { text: '它解决什么问题', link: '/guide/problem' },
            { text: '工作原理', link: '/guide/how-it-works' },
            { text: '快速开始', link: '/guide/getting-started' },
            { text: '安装', link: '/guide/installation' }
          ]
        },
        {
          text: '🧭 核心概念',
          collapsed: false,
          items: [
            { text: '客户端 Client', link: '/guide/client-concept' },
            { text: '响应格式 Format', link: '/guide/format-concept' },
            { text: '字段查询 Field', link: '/guide/field-concept' },
            { text: '错误处理', link: '/guide/error-concept' },
            { text: '认证机制', link: '/guide/auth-concept' },
            { text: '重试与限流', link: '/guide/retry-concept' }
          ]
        },
        {
          text: '🌍 IP 地址',
          collapsed: false,
          items: [
            { text: 'IPv4 查询', link: '/guide/ipv4' },
            { text: 'IPv6 查询', link: '/guide/ipv6' },
            { text: '客户端 IP 检测', link: '/guide/client-ip' },
            { text: '保留 IP 地址', link: '/guide/reserved-ip' }
          ]
        },
        {
          text: '🛠 进阶主题',
          collapsed: false,
          items: [
            { text: '上下文 Context', link: '/guide/context' },
            { text: '自定义 HTTP 客户端', link: '/guide/custom-http' },
            { text: 'JSONP 回调', link: '/guide/jsonp' },
            { text: '多格式响应', link: '/guide/formats' },
            { text: '批量查询', link: '/guide/batch' }
          ]
        }
      ],

      '/api/': [
        {
          text: '📦 包概览',
          collapsed: false,
          items: [
            { text: 'ipapi 包总览', link: '/api/' },
            { text: 'Client 客户端', link: '/api/client' },
            { text: 'API 方法', link: '/api/methods' },
            { text: '数据模型', link: '/api/models' },
            { text: '错误类型', link: '/api/errors' },
            { text: '选项函数 Options', link: '/api/options' }
          ]
        },
        {
          text: '🔧 方法详解',
          collapsed: false,
          items: [
            { text: 'GetIPInfo', link: '/api/get-ip-info' },
            { text: 'GetIPInfoRaw', link: '/api/get-ip-info-raw' },
            { text: 'GetField', link: '/api/get-field' },
            { text: 'GetClientIPInfo', link: '/api/get-client-ip-info' },
            { text: 'GetClientIPInfoRaw', link: '/api/get-client-ip-info-raw' },
            { text: 'GetClientField', link: '/api/get-client-field' }
          ]
        },
        {
          text: '🏭 构造与配置',
          collapsed: false,
          items: [
            { text: 'NewClient', link: '/api/new-client' },
            { text: 'WithAPIKey', link: '/api/with-api-key' },
            { text: 'WithAPIKeyQuery', link: '/api/with-api-key-query' },
            { text: 'WithCustomHTTPClient', link: '/api/with-custom-http-client' },
            { text: 'WithErrorHandler', link: '/api/with-error-handler' },
            { text: 'WithCallback', link: '/api/with-callback' },
            { text: 'WithBaseURL', link: '/api/with-base-url' },
            { text: 'WithUserAgent', link: '/api/with-user-agent' },
            { text: 'WithRetries', link: '/api/with-retries' },
            { text: 'WithTimeout', link: '/api/with-timeout' },
            { text: 'WithRateLimiter', link: '/api/with-rate-limiter' }
          ]
        },
        {
          text: '🗃 数据字段',
          collapsed: true,
          items: [
            { text: 'IPInfo 字段总览', link: '/api/fields' },
            { text: 'ip / network / version', link: '/api/field-network' },
            { text: 'city / region / postal', link: '/api/field-geo' },
            { text: 'country 系列', link: '/api/field-country' },
            { text: 'latitude / longitude / latlong', link: '/api/field-coord' },
            { text: 'timezone / utc_offset', link: '/api/field-time' },
            { text: 'currency / languages', link: '/api/field-currency' },
            { text: 'asn / org / hostname', link: '/api/field-asn' },
            { text: 'country_area / population', link: '/api/field-stats' }
          ]
        },
        {
          text: '🛡 错误与校验',
          collapsed: false,
          items: [
            { text: '错误类型一览', link: '/api/errors' },
            { text: 'ValidateIP', link: '/api/validate-ip' },
            { text: 'ValidateFormat', link: '/api/validate-format' },
            { text: 'IsRetryableError', link: '/api/is-retryable' },
            { text: 'WrapError', link: '/api/wrap-error' }
          ]
        }
      ],

      '/examples/': [
        {
          text: '🧪 示例代码',
          collapsed: false,
          items: [
            { text: '基础用法', link: '/examples/basic-usage' },
            { text: '高级用法', link: '/examples/advanced-usage' },
            { text: '错误处理', link: '/examples/error-handling' }
          ]
        },
        {
          text: '💡 实战场景',
          collapsed: false,
          items: [
            { text: '查询指定 IP', link: '/examples/lookup-specific-ip' },
            { text: '获取客户端 IP', link: '/examples/lookup-client-ip' },
            { text: '单字段查询', link: '/examples/single-field' },
            { text: 'XML/CSV/YAML 响应', link: '/examples/raw-formats' },
            { text: 'JSONP 回调', link: '/examples/jsonp' },
            { text: '批量 IP 查询', link: '/examples/batch-lookup' },
            { text: 'IPv6 查询', link: '/examples/ipv6' },
            { text: '带 API Key 认证', link: '/examples/with-api-key' },
            { text: '自定义错误处理', link: '/examples/custom-error' },
            { text: '经纬度解析', link: '/examples/parse-latlong' }
          ]
        }
      ],

      '/config/': [
        {
          text: '⚙️ 配置与部署',
          collapsed: false,
          items: [
            { text: 'CI/CD 概述', link: '/config/cicd' },
            { text: 'GitHub Actions', link: '/config/github-actions' },
            { text: 'GitHub Pages 部署', link: '/config/github-pages' },
            { text: '本地预览', link: '/config/local-preview' }
          ]
        }
      ],

      '/reference/': [
        {
          text: '📚 深度参考',
          collapsed: false,
          items: [
            { text: '参考总览', link: '/reference/' }
          ]
        },
        {
          text: '🛡 错误详解',
          collapsed: false,
          items: [
            { text: 'err-invalid-field', link: '/reference/errors/err-invalid-field' },
            { text: 'err-invalid-format', link: '/reference/errors/err-invalid-format' },
            { text: 'err-invalid-ip', link: '/reference/errors/err-invalid-ip' },
            { text: 'err-invalid-key', link: '/reference/errors/err-invalid-key' },
            { text: 'err-method-not-allowed', link: '/reference/errors/err-method-not-allowed' },
            { text: 'err-not-found', link: '/reference/errors/err-not-found' },
            { text: 'err-rate-limited', link: '/reference/errors/err-rate-limited' },
            { text: 'err-reserved-ip', link: '/reference/errors/err-reserved-ip' },
            { text: 'err-server-error', link: '/reference/errors/err-server-error' },
            { text: 'err-unexpected-data', link: '/reference/errors/err-unexpected-data' }
          ]
        },
        {
          text: '🗂 字段详解',
          collapsed: false,
          items: [
            { text: 'ip', link: '/reference/fields/ip' },
            { text: 'network', link: '/reference/fields/network' },
            { text: 'version', link: '/reference/fields/version' },
            { text: 'city', link: '/reference/fields/city' },
            { text: 'region', link: '/reference/fields/region' },
            { text: 'region_code', link: '/reference/fields/region_code' },
            { text: 'postal', link: '/reference/fields/postal' },
            { text: 'country', link: '/reference/fields/country' },
            { text: 'country_name', link: '/reference/fields/country_name' },
            { text: 'country_code', link: '/reference/fields/country_code' },
            { text: 'country_code_iso3', link: '/reference/fields/country_code_iso3' },
            { text: 'country_capital', link: '/reference/fields/country_capital' },
            { text: 'country_tld', link: '/reference/fields/country_tld' },
            { text: 'continent_code', link: '/reference/fields/continent_code' },
            { text: 'in_eu', link: '/reference/fields/in_eu' },
            { text: 'latitude', link: '/reference/fields/latitude' },
            { text: 'longitude', link: '/reference/fields/longitude' },
            { text: 'latlong', link: '/reference/fields/latlong' },
            { text: 'timezone', link: '/reference/fields/timezone' },
            { text: 'utc_offset', link: '/reference/fields/utc_offset' },
            { text: 'country_calling_code', link: '/reference/fields/country_calling_code' },
            { text: 'languages', link: '/reference/fields/languages' },
            { text: 'currency', link: '/reference/fields/currency' },
            { text: 'currency_name', link: '/reference/fields/currency_name' },
            { text: 'country_area', link: '/reference/fields/country_area' },
            { text: 'country_population', link: '/reference/fields/country_population' },
            { text: 'asn', link: '/reference/fields/asn' },
            { text: 'org', link: '/reference/fields/org' },
            { text: 'hostname', link: '/reference/fields/hostname' }
          ]
        },
        {
          text: '⚙️ 常量与内部',
          collapsed: false,
          items: [
            { text: 'format-json', link: '/reference/internal/format-json' },
            { text: 'format-jsonp', link: '/reference/internal/format-jsonp' },
            { text: 'format-xml', link: '/reference/internal/format-xml' },
            { text: 'format-csv', link: '/reference/internal/format-csv' },
            { text: 'format-yaml', link: '/reference/internal/format-yaml' },
            { text: 'apikey-header', link: '/reference/internal/apikey-header' },
            { text: 'apikey-query', link: '/reference/internal/apikey-query' },
            { text: 'default-base-url', link: '/reference/internal/default-base-url' },
            { text: 'default-timeout', link: '/reference/internal/default-timeout' },
            { text: 'max-redirects', link: '/reference/internal/max-redirects' },
            { text: 'default-retry-delay', link: '/reference/internal/default-retry-delay' },
            { text: 'valid-formats', link: '/reference/internal/valid-formats' },
            { text: 'valid-fields', link: '/reference/internal/valid-fields' },
            { text: 'do-request', link: '/reference/internal/do-request' },
            { text: 'apply-auth', link: '/reference/internal/apply-auth' },
            { text: 'set-headers', link: '/reference/internal/set-headers' },
            { text: 'map-status-code', link: '/reference/internal/map-status-code' },
            { text: 'handle-error', link: '/reference/internal/handle-error' },
            { text: 'new-get-request', link: '/reference/internal/new-get-request' },
            { text: 'parselatlong', link: '/reference/internal/parselatlong' },
            { text: 'get-postal', link: '/reference/internal/get-postal' },
            { text: 'apierror-error', link: '/reference/internal/apierror-error' },
            { text: 'apierror-toerror', link: '/reference/internal/apierror-toerror' }
          ]
        }
      ],

      '/cookbook/': [
        {
          text: '🍳 Cookbook 食谱',
          collapsed: false,
          items: [
            { text: '食谱总览', link: '/cookbook/' },
            { text: 'geoip-middleware', link: '/cookbook/geoip-middleware' },
            { text: 'rate-limit-by-country', link: '/cookbook/rate-limit-by-country' },
            { text: 'redirect-by-language', link: '/cookbook/redirect-by-language' },
            { text: 'currency-display', link: '/cookbook/currency-display' },
            { text: 'timezone-greeting', link: '/cookbook/timezone-greeting' },
            { text: 'eu-compliance', link: '/cookbook/eu-compliance' },
            { text: 'cdn-edge-detection', link: '/cookbook/cdn-edge-detection' },
            { text: 'log-enrichment', link: '/cookbook/log-enrichment' },
            { text: 'fraud-detection', link: '/cookbook/fraud-detection' },
            { text: 'analytics-aggregation', link: '/cookbook/analytics-aggregation' },
            { text: 'async-lookup', link: '/cookbook/async-lookup' },
            { text: 'cached-lookup', link: '/cookbook/cached-lookup' },
            { text: 'proxy-detection', link: '/cookbook/proxy-detection' },
            { text: 'grpc-interceptor', link: '/cookbook/grpc-interceptor' },
            { text: 'scheduled-batch', link: '/cookbook/scheduled-batch' },
            { text: 'jsonp-frontend', link: '/cookbook/jsonp-frontend' },
            { text: 'csv-export', link: '/cookbook/csv-export' },
            { text: 'yaml-config', link: '/cookbook/yaml-config' },
            { text: 'asn-blocklist', link: '/cookbook/asn-blocklist' },
            { text: 'nearest-server', link: '/cookbook/nearest-server' }
          ]
        }
      ],

      '/faq/': [
        {
          text: '❓ 常见问题',
          collapsed: false,
          items: [
            { text: 'FAQ 总览', link: '/faq/' },
            { text: 'free-quota', link: '/faq/free-quota' },
            { text: 'need-apikey', link: '/faq/need-apikey' },
            { text: 'ipv6-support', link: '/faq/ipv6-support' },
            { text: 'batch-endpoint', link: '/faq/batch-endpoint' },
            { text: 'rate-limit-429', link: '/faq/rate-limit-429' },
            { text: 'private-ip-error', link: '/faq/private-ip-error' },
            { text: 'best-format', link: '/faq/best-format' },
            { text: 'reuse-client', link: '/faq/reuse-client' },
            { text: 'context-timeout', link: '/faq/context-timeout' },
            { text: 'custom-error-handler', link: '/faq/custom-error-handler' },
            { text: 'jsonp-with-apikey', link: '/faq/jsonp-with-apikey' },
            { text: 'concurrent-safe', link: '/faq/concurrent-safe' },
            { text: 'error-vs-apierror', link: '/faq/error-vs-apierror' },
            { text: 'retry-count', link: '/faq/retry-count' },
            { text: 'redirect-limit', link: '/faq/redirect-limit' },
            { text: 'hostname-empty', link: '/faq/hostname-empty' },
            { text: 'postal-nil', link: '/faq/postal-nil' },
            { text: 'latlong-vs-latlon', link: '/faq/latlong-vs-latlon' },
            { text: 'set-baseurl', link: '/faq/set-baseurl' },
            { text: 'test-coverage', link: '/faq/test-coverage' }
          ]
        }
      ],

      '/best-practices/': [
        {
          text: '✅ 最佳实践',
          collapsed: false,
          items: [
            { text: '最佳实践总览', link: '/best-practices/' },
            { text: 'client-lifecycle', link: '/best-practices/client-lifecycle' },
            { text: 'error-handling-strategy', link: '/best-practices/error-handling-strategy' },
            { text: 'timeout-strategy', link: '/best-practices/timeout-strategy' },
            { text: 'retry-strategy', link: '/best-practices/retry-strategy' },
            { text: 'rate-limit-strategy', link: '/best-practices/rate-limit-strategy' },
            { text: 'secret-management', link: '/best-practices/secret-management' },
            { text: 'observability', link: '/best-practices/observability' },
            { text: 'testing', link: '/best-practices/testing' },
            { text: 'performance', link: '/best-practices/performance' },
            { text: 'security', link: '/best-practices/security' },
            { text: 'localization', link: '/best-practices/localization' },
            { text: 'graceful-degradation', link: '/best-practices/graceful-degradation' }
          ]
        }
      ],

      '/tutorials/': [
        {
          text: '🌱 入门',
          collapsed: false,
          items: [
            { text: '教程总览', link: '/tutorials/' },
            { text: 'hello-ipapi', link: '/tutorials/hello-ipapi' },
            { text: 'first-client', link: '/tutorials/first-client' },
            { text: 'explore-ipinfo', link: '/tutorials/explore-ipinfo' },
            { text: 'single-field-tutorial', link: '/tutorials/single-field-tutorial' },
            { text: 'client-ip-tutorial', link: '/tutorials/client-ip-tutorial' }
          ]
        },
        {
          text: '🛡 错误与校验',
          collapsed: false,
          items: [
            { text: 'error-branches-tutorial', link: '/tutorials/error-branches-tutorial' },
            { text: 'reserved-ip-tutorial', link: '/tutorials/reserved-ip-tutorial' },
            { text: 'error-handler-tutorial', link: '/tutorials/error-handler-tutorial' }
          ]
        },
        {
          text: '🔧 配置与认证',
          collapsed: false,
          items: [
            { text: 'apikey-setup', link: '/tutorials/apikey-setup' },
            { text: 'query-auth-modes', link: '/tutorials/query-auth-modes' },
            { text: 'custom-http-tutorial', link: '/tutorials/custom-http-tutorial' },
            { text: 'context-timeout-tutorial', link: '/tutorials/context-timeout-tutorial' },
            { text: 'rate-limit-tutorial', link: '/tutorials/rate-limit-tutorial' },
            { text: 'retry-tutorial', link: '/tutorials/retry-tutorial' }
          ]
        },
        {
          text: '🌐 数据与格式',
          collapsed: false,
          items: [
            { text: 'raw-formats-tutorial', link: '/tutorials/raw-formats-tutorial' },
            { text: 'jsonp-tutorial', link: '/tutorials/jsonp-tutorial' },
            { text: 'latlong-tutorial', link: '/tutorials/latlong-tutorial' },
            { text: 'ipv6-tutorial', link: '/tutorials/ipv6-tutorial' }
          ]
        },
        {
          text: '🚀 实战集成',
          collapsed: false,
          items: [
            { text: 'batch-tutorial', link: '/tutorials/batch-tutorial' },
            { text: 'middleware-tutorial', link: '/tutorials/middleware-tutorial' },
            { text: 'currency-localization-tutorial', link: '/tutorials/currency-localization-tutorial' },
            { text: 'timezone-display-tutorial', link: '/tutorials/timezone-display-tutorial' },
            { text: 'asn-filter-tutorial', link: '/tutorials/asn-filter-tutorial' },
            { text: 'csv-report-tutorial', link: '/tutorials/csv-report-tutorial' }
          ]
        },
        {
          text: '🧪 测试',
          collapsed: false,
          items: [
            { text: 'test-mock-tutorial', link: '/tutorials/test-mock-tutorial' }
          ]
        }
      ]
    },

    // 社交链接
    socialLinks: [
      { icon: 'github', link: 'https://github.com/cyberspacesec/ipapi.co-skills' }
    ],

    // 搜索
    search: {
      provider: 'local',
      options: {
        translations: {
          button: {
            buttonText: '搜索文档',
            buttonAriaLabel: '搜索文档'
          },
          modal: {
            noResultsText: '无法找到相关结果',
            resetButtonTitle: '清除查询条件',
            footer: {
              selectText: '选择',
              navigateText: '切换'
            }
          }
        }
      }
    },

    // 大纲
    outline: {
      level: [2, 3],
      label: '本页目录'
    },

    docFooter: {
      prev: '上一页',
      next: '下一页'
    },

    lastUpdatedText: '最后更新于',

    returnToTopLabel: '回到顶部',
    sidebarMenuLabel: '菜单',

    editLink: {
      pattern: 'https://github.com/cyberspacesec/ipapi.co-skills/edit/main/website/:path',
      text: '在 GitHub 上编辑此页'
    },

    footer: {
      message: '基于 MIT 许可证发布',
      copyright: '© 2024-present cyberspacesec'
    }
  }
}))
