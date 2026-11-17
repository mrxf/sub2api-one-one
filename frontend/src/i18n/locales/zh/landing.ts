export default {
  batchImageGuide: {
    title: '图片批量生成',
    description: '一次提交多条提示词，任务完成后可统一下载图片结果'
  },
  // Home Page
  home: {
    viewOnGithub: '在 GitHub 上查看',
    viewDocs: '查看文档',
    docs: '文档',
    switchToLight: '切换到浅色模式',
    switchToDark: '切换到深色模式',
    dashboard: '控制台',
    login: '登录',
    getStarted: '立即开始',
    goToDashboard: '进入控制台',
    // 新增：面向用户的价值主张
    heroSubtitle: '一个密钥，畅用多个 AI 模型',
    heroDescription: '无需管理多个订阅账号，一站式接入 Claude、GPT、Gemini 等主流 AI 服务',
    tags: {
      subscriptionToApi: '订阅转 API',
      stickySession: '会话保持',
      realtimeBilling: '按量计费'
    },
    // 用户痛点区块
    painPoints: {
      title: '你是否也遇到这些问题？',
      items: {
        expensive: {
          title: '订阅费用高',
          desc: '每个 AI 服务都要单独订阅，每月支出越来越多'
        },
        complex: {
          title: '多账号难管理',
          desc: '不同平台的账号、密钥分散各处，管理起来很麻烦'
        },
        unstable: {
          title: '服务不稳定',
          desc: '单一账号容易触发限制，影响正常使用'
        },
        noControl: {
          title: '用量无法控制',
          desc: '不知道钱花在哪了，也无法限制团队成员的使用'
        }
      }
    },
    // 解决方案区块
    solutions: {
      title: '我们帮你解决',
      subtitle: '简单三步，开始省心使用 AI'
    },
    features: {
      unifiedGateway: '一键接入',
      unifiedGatewayDesc: '获取一个 API 密钥，即可调用所有已接入的 AI 模型，无需分别申请。',
      multiAccount: '稳定可靠',
      multiAccountDesc: '智能调度多个上游账号，自动切换和负载均衡，告别频繁报错。',
      balanceQuota: '用多少付多少',
      balanceQuotaDesc: '按实际使用量计费，支持设置配额上限，团队用量一目了然。'
    },
    // 优势对比
    comparison: {
      title: '为什么选择我们？',
      headers: {
        feature: '对比项',
        official: '官方订阅',
        us: '本平台'
      },
      items: {
        pricing: {
          feature: '付费方式',
          official: '固定月费，用不完也付',
          us: '按量付费，用多少付多少'
        },
        models: {
          feature: '模型选择',
          official: '单一服务商',
          us: '多模型随意切换'
        },
        management: {
          feature: '账号管理',
          official: '每个服务单独管理',
          us: '统一密钥，一站管理'
        },
        stability: {
          feature: '服务稳定性',
          official: '单账号易触发限制',
          us: '多账号池，自动切换'
        },
        control: {
          feature: '用量控制',
          official: '无法限制',
          us: '可设配额、查明细'
        }
      }
    },
    providers: {
      title: '已支持的 AI 模型',
      description: '一个 API，多种选择',
      supported: '已支持',
      soon: '即将推出',
      claude: 'Claude',
      gemini: 'Gemini',
      antigravity: 'Antigravity',
      more: '更多'
    },
    // CTA 区块
    cta: {
      title: '准备好开始了吗？',
      description: '注册即可获得免费试用额度，体验一站式 AI 服务',
      button: '免费注册'
    },
    footer: {
      allRightsReserved: '保留所有权利。'
    },
    // 独享账号主题首页重构文案
    dedicatedLanding: {
      nav: {
        capabilities: '独享能力',
        routing: '路由架构',
        marketplace: '现货与买断',
        comparison: '方案对比',
        ecosystem: '接入生态',
        keyUsage: '用量查询',
        dedicatedCenter: '独享账号中心',
        register: '免费注册'
      },
      hero: {
        overline: '企业级原生订阅网关 · 1:1 物理隔离独享架构',
        title: '专为高强度研发打造的原生独享账号基础设施',
        description:
          '告别共享号池频繁切号风控与隐形倍率扣费。每台独享账号均由独立代理环境隔离托管，开通即签发 1 对 1 专属锁定密钥，有效期内 0 钱包余额直通放行、5H / 7D 官方窗口实时透视，支持内置官方重置卡与随时一键买断提号。',
        primaryCtaGuest: '选购独享账号',
        primaryCtaAuthed: '进入独享账号中心',
        secondaryCtaGuest: '查询密钥用量',
        secondaryCtaAuthed: '打开控制台',
        metrics: {
          isolationValue: '1 : 1',
          isolationLabel: '单钥锁单号 · 物理防串用',
          balanceValue: '¥0',
          balanceLabel: '有效期内调用钱包门槛',
          overflowValue: '0%',
          overflowLabel: '公共计费号池静默溢出率',
          windowValue: '5H / 7D',
          windowLabel: '官方窗口实时只读同步'
        }
      },
      showcase: {
        tabConsole: '独享账号实时看板',
        tabTerminal: '1:1 网关路由验证',
        liveStatus: '独立隔离运行中',
        zeroBalanceTag: '0 余额直通',
        pinnedTag: '1:1 专属锁定',
        redactedShield: '底层邮箱 / 组织 ID / 凭据已物理脱敏隔离',
        fiveHourLabel: '5 小时滚动窗口剩余',
        sevenDayLabel: '7 天周期窗口剩余',
        resetsIn2h: '2h 18m 后自动重置',
        resetsIn4d: '4d 11h 后自动重置',
        resetCreditTag: '内置官方重置卡 × 2',
        buyoutReadyTag: '支持随时一键买断纯文本凭据',
        boundKeyLabel: '关联专属密钥',
        accounts: {
          claude: {
            name: 'Claude Max 20x 独享 #01',
            group: 'anthropic-max-dedicated',
            concurrency: '5 并发独享'
          },
          openai: {
            name: 'OpenAI Codex Pro 独享 #04',
            group: 'openai-pro-dedicated',
            concurrency: '8 并发独享'
          },
          gemini: {
            name: 'Gemini Ultra 原生独享 #02',
            group: 'gemini-ultra-dedicated',
            concurrency: '6 并发独享'
          }
        },
        terminalTitle: '1:1 专属锁定路由 · 零溢出验证',
        terminalCommentPin: '# 严格命中名下 1:1 独享实例 #01（绝不跨号切换）',
        terminalCommentPass: '# 计费校验：独享包月有效期内（钱包扣费 ¥0.00）'
      },
      capabilities: {
        overline: 'CORE CAPABILITIES',
        title: '围绕「确定性、防封安全与资产自主」重构独享体验',
        subtitle:
          '从网关调度内核到底层敏感数据清洗，每一层均针对重度 AI 编码与生产级长会话场景深度定制。',
        items: {
          isolation: {
            index: '01',
            title: '1:1 物理级防串号锁定',
            desc: '开通即自动生成默认「1 Key 对 1 号」严格锁定密钥，所有上下文与长连接固定落入同一原生账号与专属代理链路，彻底规避多人混用切号引发的连坐封禁。',
            spec1: '1 Key : 1 Account 默认锁定',
            spec2: '支持名下多号 N:1 私有轮询'
          },
          zeroOverflow: {
            index: '02',
            title: '0 余额直通与绝对零溢出承诺',
            desc: '采用一次性包月制，激活后平台钱包 0 余额即可满速调用。当官方 5H / 7D 窗口达到上限时，网关直接返回明确限流状态与恢复时间，绝不静默回退至公共池扣费。',
            spec1: '有效期内 ¥0 钱包余额放行',
            spec2: '窗口满额显式限流 · 零意外账单'
          },
          quotaWindow: {
            index: '03',
            title: '5H / 7D 官方窗口透视与重置卡',
            desc: '只读看板实时同步上游 5 小时滚动窗口、7 天周期窗口及 Sonnet 专项剩余百分比与重置倒计时；针对内置官方重置卡的账号，支持在线一键核销秒级满血恢复。',
            spec1: '毫秒级官方真实配额映射',
            spec2: '一键核销内置官方额度重置卡'
          },
          buyoutShield: {
            index: '04',
            title: '全链路隐私脱敏与随时买断提号',
            desc: '未买断期间，网关自动剥离并清洗上游邮箱、组织 ID 与敏感错误头，严防信息泄露导致恶意封号；支持商城直接买断或租用中途随时补差买断，即刻获取纯文本原始凭据。',
            spec1: '敏感响应头与身份全自动清洗',
            spec2: '一键买断解锁邮箱 / 密码 / 2FA'
          }
        }
      },
      routing: {
        overline: 'ROUTING ARCHITECTURE',
        title: '三态路由引擎，一套网关满足个人极客到研发团队',
        subtitle:
          '在「API 密钥」管理页可按单把密钥粒度自由指定路由策略，独享资产与公共弹性号池在同一平台内清晰分界、互不干扰。',
        clientLabel: '开发终端与 SDK 请求',
        clientSub: 'Claude Code · Codex · Cursor · API',
        gatewayLabel: 'Sub2API 路由与风控隔离网关',
        publicPoolBlocked: '公共计费号池已物理阻断（零溢出）',
        specsTitle: '当前路由策略运行契约',
        specBilling: '计费与余额要求',
        specFailover: '切换与容灾边界',
        specOverflow: '窗口耗尽行为',
        specScenario: '典型适用场景',
        modes: {
          pinned: {
            badge: '防串号首选',
            title: '1:1 锁定单个独享账号',
            summary: '单把 API Key 严格绑定指定独享号，零跨号切换，最大化长上下文缓存命中率与账号存活率。',
            targetTitle: '指定独享实例 #01（独立固定出口）',
            targetSub: '100% 流量与 Prompt Cache 固定命中同一原生账号',
            billing: '0 钱包余额直通，不扣任何按量费用',
            failover: '严格不跨号，保障指纹与会话绝对纯净',
            overflow: '返回显式 429 冷却倒计时，绝不溢出公共池',
            scenario: 'Claude Code / Cursor 单人高强度日常主力研发'
          },
          pool: {
            badge: '高并发团队',
            title: 'N:1 名下多独享号私有轮询',
            summary: '仅在您本人名下同分组的多台独享账号之间智能负载均衡与限流自动切换，兼顾私有安全与高吞吐。',
            targetTitle: '您名下的私有独享集群 [#01 · #02 · #03]',
            targetSub: '单号进入 5H 冷却时毫秒级无缝切至名下其余健康独享号',
            billing: '0 钱包余额直通，仅消耗您名下独享窗口',
            failover: '严格限定于当前用户名下私有独享池内故障转移',
            overflow: '全部名下独享号冷却时显式提示，绝不溢出公共池',
            scenario: '多席位工作室、CI 自动化流水线与高并发 Agent 任务'
          },
          shared: {
            badge: '弹性补充',
            title: '公共弹性共享号池',
            summary: '无需持有独享号，按实际 Token 用量或分组订阅额度弹性调用全平台多模型通道。',
            targetTitle: '平台多上游智能高可用共享池',
            targetSub: '跨多健康节点动态调度、会话粘性保持与自动容灾重试',
            billing: '按分组倍率从平台钱包余额或订阅额度实时结算',
            failover: '全池健康节点自动负载均衡与异常秒级重试',
            overflow: '受密钥配额上限与钱包余额精准管控',
            scenario: '多模型探索对比、低频弹性调用与峰值补充算力'
          }
        }
      },
      marketplace: {
        overline: 'DECAY MARKETPLACE & BUYOUT',
        title: '锁仓现货商城：每日透明递减定价，即买即用或一键买断',
        subtitle:
          '所有待售独享账号均经管理员预配置验证并物理锁仓隔离，确保交付到您手中时拥有 100% 未被污染的纯净窗口。',
        steps: {
          step1Title: '上架物理锁仓，杜绝售前串用损耗',
          step1Desc: '账号一旦设为「上架商城待售」，网关调度器立即将其从所有流量池摘除隔离，实时公示当前 5H / 7D 额度快照、并发上限与有效天数。',
          step2Title: '24 小时自动递减定价 + 15 分钟排他锁单',
          step2Desc: '支持按日自动立减机制（每满 24 小时自动降价直至保底底价），支持钱包余额秒开或在线扫码支付，发起支付即排他锁单 15 分钟防止抢单冲突。',
          step3Title: '秒级自动过户签发，全周期可升级买断',
          step3Desc: '付款完成瞬间自动划归您名下并生成 1:1 专属 API Key；对支持买断的账号，无论购入时还是使用中途，均可随时一键买断提取纯文本原始凭据。'
        },
        simulator: {
          headerTitle: '现货定价递减与买断交付机制示意',
          stockBadge: '锁仓现货 · 1人1号',
          sampleAlias: 'Claude Max 20x 原生独享 #08',
          sampleMeta: '有效期 30 天 · 5 并发独享 · 5H 剩余 100%',
          daySelectorLabel: '预览上架时间节点价格',
          day1: '上架首日',
          day3: '第 3 天',
          day7: '第 7 天（触达保底）',
          currentPriceLabel: '当前独享开通价',
          basePriceNote: '上架原价 ¥220 · 每日自动立减 ¥15（保底价 ¥130）',
          leaseModeTitle: '包月独享托管模式（含防封售后保障）',
          leaseModeDesc: '交付 1:1 专属 API Key + 5H/7D 只读看板，底层凭据由平台物理脱敏保护，享受平台约定的售后策略。',
          buyoutModeTitle: '可选：一键买断提号（解锁纯文本原始凭据）',
          buyoutModeDesc: '支付买断价后立即解锁预置的登录邮箱、密码及 2FA 恢复码纯文本信息，同时保留现有 1:1 API Key（买断移交控制权后不再适用封号补偿售后）。',
          actionCta: '前往独享商城查看实时在售列表'
        }
      },
      comparison: {
        overline: 'WHY DEDICATED',
        title: '为什么严肃研发场景更适合本平台 1:1 独享账号？',
        subtitle: '在账号纯净度、费用确定性与工程接入效率之间，无需再做妥协。',
        colDimension: '核心评估维度',
        colOfficial: '自行维护海外官方直连',
        colSharedRelay: '传统拼车 / 普通共享中转',
        colDedicated: '本平台 1:1 原生独享账号',
        recommendedTag: '工程首选',
        rows: {
          isolation: {
            dim: '账号隔离与防封安全性',
            official: '需自备纯净海外网络与外币卡，支付拒付或节点跳变易风控',
            shared: '多人高频混用同一账号，极易因他人违规触发连坐封号',
            dedicated: '1 人 1 号独立固定代理链路 + 1:1 密钥锁定，严防交叉污染'
          },
          quota: {
            dim: '5H / 7D 窗口确定性',
            official: '原生满额窗口，但缺少 API 化多端共享与只读监控面板',
            shared: '高峰期窗口常被他人抢占耗尽，实际可用额度不可预期',
            dedicated: '100% 独占官方原生窗口，面板实时同步 5H / 7D 剩余百分比'
          },
          billing: {
            dim: '计费透明度与溢出风险',
            official: '固定美元月费，无溢出扣费但闲置成本高',
            shared: '复杂分组倍率计费，长上下文与缓存极易产生高额账单',
            dedicated: '一次性包月制（支持每日递减淘号），0 余额直通且绝对零溢出'
          },
          recovery: {
            dim: '突发限流自救能力',
            official: '需人工登录网页后台排查操作',
            shared: '被动等待号池恢复，无法干预单号状态',
            dedicated: '支持在线一键核销内置官方重置卡，或配置名下多号 N:1 轮询'
          },
          privacy: {
            dim: '敏感身份隔离与资产买断',
            official: '完全自有，但团队协作时共享原始密码存在泄密风险',
            shared: '纯黑盒调用，无法获取底层账号所有权',
            dedicated: '托管期全链路剥离邮箱与敏感头防封；支持随时一键买断提取纯文本凭据'
          }
        }
      },
      ecosystem: {
        overline: 'ECOSYSTEM & QUICK START',
        title: '全栈主流模型原生协议兼容，三行配置即刻投产',
        subtitle:
          '完全兼容 Anthropic Messages、OpenAI Responses / Chat Completions 与 Gemini 原生接口，无缝接入您熟悉的 AI 编程终端。',
        providersTitle: '支持的上游原生平台',
        clientsTitle: '零改造直连主流研发工具',
        copyBtn: '复制配置代码',
        copiedBtn: '已复制',
        selfServiceKeyCheck: '已有 API Key？无需登录直接查询实时额度与消费明细'
      },
      bottomCta: {
        title: '立即开启纯净、可控、零溢出的独享 AI 研发体验',
        subtitle:
          '前往独享账号商城挑选现货原生账号，秒级自动签发 1:1 专属 API Key，亦可按需搭配弹性共享号池灵活扩展。',
        primaryAuthed: '进入独享账号中心',
        primaryGuest: '立即登录并选购独享号',
        secondaryRegister: '免费注册账号',
        secondaryKeyUsage: 'API Key 用量查询'
      },
      footerCols: {
        productTitle: '产品与架构',
        toolsTitle: '自助工具与资源',
        brandTagline: '集 1:1 物理隔离独享账号、每日递减现货商城、买断提号与多模型弹性网关于一体的专业级 AI 基础设施平台。'
      }
    }
  },

  // Key Usage Query Page
  keyUsage: {
    title: 'API Key 用量查询',
    subtitle: '输入您的 API Key 以查看实时消费金额与使用状态',
    placeholder: 'sk-ant-mirror-xxxxxxxxxxxx',
    query: '查询',
    querying: '查询中...',
    privacyNote: '您的 Key 仅在浏览器本地处理，不会被存储',
    dateRange: '统计范围:',
    dateRangeToday: '今日',
    dateRange7d: '7 天',
    dateRange30d: '30 天',
    dateRange90d: '90 天',
    dateRangeCustom: '自定义',
    apply: '应用',
    used: '已使用',
    detailInfo: '详细信息',
    tokenStats: 'Token 统计',
    dailyDetail: '按日明细',
    modelStats: '模型用量统计',
    // Table headers
    date: '日期',
    model: '模型',
    requests: '请求数',
    inputTokens: '输入 Tokens',
    outputTokens: '输出 Tokens',
    cacheCreationTokens: '缓存创建',
    cacheReadTokens: '缓存读取',
    cacheWriteTokens: '缓存写入',
    totalTokens: '总 Tokens',
    cost: '费用',
    // Status
    quotaMode: 'Key 限额模式',
    walletBalance: '钱包余额',
    // Ring card titles
    totalQuota: '总额度',
    limit5h: '5 小时限额',
    limitDaily: '日限额',
    limit7d: '7 天限额',
    limitWeekly: '周限额',
    limitMonthly: '月限额',
    // Detail rows
    remainingQuota: '剩余额度',
    expiresAt: '过期时间',
    todayExpires: '(今日到期)',
    daysLeft: '({days} 天)',
    usedQuota: '已用额度',
    resetNow: '即将重置',
    subscriptionType: '订阅类型',
    billingType: '计费方式',
    subscriptionExpires: '订阅到期',
    // Usage stat cells
    todayRequests: '今日请求',
    todayInputTokens: '今日输入',
    todayOutputTokens: '今日输出',
    todayTokens: '今日 Tokens',
    todayCacheCreation: '今日缓存创建',
    todayCacheRead: '今日缓存读取',
    todayCost: '今日费用',
    rpmTpm: 'RPM / TPM',
    totalRequests: '累计请求',
    totalInputTokens: '累计输入',
    totalOutputTokens: '累计输出',
    totalTokensLabel: '累计 Tokens',
    totalCacheCreation: '累计缓存创建',
    totalCacheRead: '累计缓存读取',
    totalCost: '累计费用',
    avgDuration: '平均耗时',
    // Messages
    enterApiKey: '请输入 API Key',
    querySuccess: '查询成功',
    queryFailed: '查询失败',
    queryFailedRetry: '查询失败，请稍后重试',
    noDailyUsage: '暂无按日用量数据',
  },

  // Setup Wizard
  setup: {
    title: 'Sub2API 安装向导',
    description: '配置您的 Sub2API 实例',
    database: {
      title: '数据库配置',
      description: '连接到您的 PostgreSQL 数据库',
      host: '主机',
      port: '端口',
      username: '用户名',
      password: '密码',
      databaseName: '数据库名称',
      sslMode: 'SSL 模式',
      passwordPlaceholder: '密码',
      ssl: {
        disable: '禁用',
        require: '要求',
        verifyCa: '验证 CA',
        verifyFull: '完全验证'
      }
    },
    redis: {
      title: 'Redis 配置',
      description: '连接到您的 Redis 服务器',
      host: '主机',
      port: '端口',
      username: '用户名（可选）',
      password: '密码（可选）',
      database: '数据库',
      usernamePlaceholder: '默认用户留空',
      passwordPlaceholder: '密码',
      enableTls: '启用 TLS',
      enableTlsHint: '连接 Redis 时使用 TLS（公共 CA 证书）'
    },
    admin: {
      title: '管理员账户',
      description: '创建您的管理员账户',
      email: '邮箱',
      password: '密码',
      confirmPassword: '确认密码',
      passwordPlaceholder: '至少 8 个字符',
      confirmPasswordPlaceholder: '确认密码',
      passwordMismatch: '密码不匹配'
    },
    ready: {
      title: '准备安装',
      description: '检查您的配置并完成安装',
      database: '数据库',
      redis: 'Redis',
      adminEmail: '管理员邮箱'
    },
    status: {
      testing: '测试中...',
      success: '连接成功',
      testConnection: '测试连接',
      installing: '安装中...',
      completeInstallation: '完成安装',
      completed: '安装完成！',
      redirecting: '正在跳转到登录页面...',
      restarting: '服务正在重启，请稍候...',
      timeout: '服务重启时间超出预期，请手动刷新页面。'
    }
  },

  // Common
}
