<template>
  <AppLayout>
    <div class="space-y-6">
      <!-- Top Isolation & Zero-Overflow Guarantee Banner -->
      <div
        class="relative overflow-hidden rounded-2xl border border-emerald-200/80 bg-gradient-to-r from-emerald-50/90 via-teal-50/70 to-cyan-50/60 p-4 shadow-sm dark:border-emerald-800/50 dark:from-emerald-950/40 dark:via-teal-950/30 dark:to-dark-800"
      >
        <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div class="flex items-start gap-3">
            <div
              class="mt-0.5 flex h-9 w-9 shrink-0 items-center justify-center rounded-xl bg-emerald-500/15 text-emerald-600 dark:bg-emerald-500/20 dark:text-emerald-400"
            >
              <Icon name="shield" size="md" />
            </div>
            <div>
              <h3 class="text-sm font-semibold text-emerald-950 dark:text-emerald-200">
                {{ t('dedicatedAccounts.banner.isolationTitle') }}
              </h3>
              <p class="mt-0.5 text-xs leading-relaxed text-emerald-800/90 dark:text-emerald-300/80">
                {{ t('dedicatedAccounts.banner.isolationDesc') }}
              </p>
            </div>
          </div>
          <div class="flex shrink-0 items-center gap-2">
            <button
              v-if="activeTab === 'my'"
              type="button"
              :disabled="refreshingUsage || loadingMy"
              @click="handleRefreshMyUsage"
              class="btn btn-secondary btn-sm inline-flex items-center gap-1.5"
            >
              <Icon name="refresh" size="sm" :class="{ 'animate-spin': refreshingUsage }" />
              <span>{{
                refreshingUsage
                  ? t('dedicatedAccounts.my.refreshing')
                  : t('dedicatedAccounts.my.refreshQuota')
              }}</span>
            </button>
            <button
              v-else
              type="button"
              :disabled="loadingMarket"
              @click="loadMarketplace"
              class="btn btn-secondary btn-sm inline-flex items-center gap-1.5"
            >
              <Icon name="refresh" size="sm" :class="{ 'animate-spin': loadingMarket }" />
              <span>{{ t('dedicatedAccounts.market.refreshMarket') }}</span>
            </button>
          </div>
        </div>
      </div>

      <!-- Primary Navigation Tabs -->
      <div class="flex flex-wrap items-center justify-between gap-4 border-b border-gray-200 pb-3 dark:border-dark-700">
        <div class="inline-flex rounded-xl bg-gray-100 p-1 dark:bg-dark-800">
          <button
            type="button"
            data-testid="tab-my-dedicated"
            @click="switchTab('my')"
            :class="[
              'inline-flex items-center gap-2 rounded-lg px-4 py-2 text-sm font-medium transition-all',
              activeTab === 'my'
                ? 'bg-white text-gray-900 shadow-sm dark:bg-dark-700 dark:text-white'
                : 'text-gray-600 hover:text-gray-900 dark:text-gray-400 dark:hover:text-gray-200'
            ]"
          >
            <Icon name="server" size="sm" />
            <span>{{ t('dedicatedAccounts.tabs.myAccounts') }}</span>
            <span
              :class="[
                'rounded-full px-2 py-0.5 text-xs font-semibold tabular-nums',
                activeTab === 'my'
                  ? 'bg-primary-100 text-primary-700 dark:bg-primary-900/40 dark:text-primary-300'
                  : 'bg-gray-200/80 text-gray-600 dark:bg-dark-600 dark:text-gray-300'
              ]"
            >
              {{ myAccounts.length }}
            </span>
          </button>
          <button
            type="button"
            data-testid="tab-dedicated-market"
            @click="switchTab('market')"
            :class="[
              'inline-flex items-center gap-2 rounded-lg px-4 py-2 text-sm font-medium transition-all',
              activeTab === 'market'
                ? 'bg-white text-gray-900 shadow-sm dark:bg-dark-700 dark:text-white'
                : 'text-gray-600 hover:text-gray-900 dark:text-gray-400 dark:hover:text-gray-200'
            ]"
          >
            <Icon name="shoppingBag" size="sm" />
            <span>{{ t('dedicatedAccounts.tabs.marketplace') }}</span>
            <span
              :class="[
                'rounded-full px-2 py-0.5 text-xs font-semibold tabular-nums',
                activeTab === 'market'
                  ? 'bg-amber-100 text-amber-700 dark:bg-amber-900/40 dark:text-amber-300'
                  : 'bg-gray-200/80 text-gray-600 dark:bg-dark-600 dark:text-gray-300'
              ]"
            >
              {{ marketItems.length }}
            </span>
          </button>
        </div>

        <!-- Platform Filter (Marketplace Tab) -->
        <div v-if="activeTab === 'market'" class="flex flex-wrap items-center gap-1.5">
          <button
            v-for="opt in platformFilters"
            :key="opt.value"
            type="button"
            @click="selectPlatformFilter(opt.value)"
            :class="[
              'rounded-lg px-3 py-1.5 text-xs font-medium transition-colors',
              marketPlatform === opt.value
                ? 'bg-primary-600 text-white shadow-sm'
                : 'bg-white text-gray-600 ring-1 ring-gray-200 hover:bg-gray-50 dark:bg-dark-800 dark:text-gray-300 dark:ring-dark-700 dark:hover:bg-dark-700'
            ]"
          >
            {{ opt.label }}
          </button>
        </div>
      </div>

      <!-- TAB 1: MY DEDICATED ACCOUNTS (STRICT READ-ONLY EXCEPT RESET CREDITS & OPTIONAL BUYOUT) -->
      <div v-if="activeTab === 'my'">
        <div v-if="loadingMy" class="grid gap-5 lg:grid-cols-2">
          <div
            v-for="idx in 2"
            :key="idx"
            class="h-72 animate-pulse rounded-2xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-800"
          />
        </div>

        <div
          v-else-if="myAccounts.length === 0"
          class="flex flex-col items-center justify-center rounded-2xl border border-dashed border-gray-300 bg-white px-6 py-16 text-center dark:border-dark-700 dark:bg-dark-800/60"
        >
          <div
            class="flex h-14 w-14 items-center justify-center rounded-2xl bg-primary-50 text-primary-600 dark:bg-primary-900/30 dark:text-primary-400"
          >
            <Icon name="key" size="lg" />
          </div>
          <h3 class="mt-4 text-base font-semibold text-gray-900 dark:text-white">
            {{ t('dedicatedAccounts.my.emptyTitle') }}
          </h3>
          <p class="mt-1.5 max-w-md text-sm text-gray-500 dark:text-gray-400">
            {{ t('dedicatedAccounts.my.emptyDesc') }}
          </p>
          <button
            type="button"
            @click="switchTab('market')"
            class="btn btn-primary mt-6 inline-flex items-center gap-2"
          >
            <Icon name="shoppingBag" size="sm" />
            <span>{{ t('dedicatedAccounts.my.goToMarket') }}</span>
          </button>
        </div>

        <div v-else class="grid gap-6 xl:grid-cols-2">
          <div
            v-for="acc in myAccounts"
            :key="acc.id"
            :data-testid="`my-dedicated-card-${acc.id}`"
            class="flex flex-col justify-between rounded-2xl border border-gray-200 bg-white p-6 shadow-sm transition-shadow hover:shadow-md dark:border-dark-700 dark:bg-dark-800"
          >
            <div class="space-y-5">
              <!-- Card Header: Platform + Alias + Health Status -->
              <div class="flex items-start justify-between gap-3">
                <div class="flex items-center gap-3 min-w-0">
                  <div
                    class="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-gray-100 text-gray-700 dark:bg-dark-700 dark:text-gray-200"
                  >
                    <PlatformIcon :platform="acc.platform" size="lg" />
                  </div>
                  <div class="min-w-0">
                    <div class="flex flex-wrap items-center gap-2">
                      <h3 class="truncate text-base font-bold text-gray-900 dark:text-white">
                        {{ acc.display_name }}
                      </h3>
                      <span
                        class="rounded-md bg-gray-100 px-2 py-0.5 font-mono text-xs text-gray-500 dark:bg-dark-700 dark:text-gray-400"
                      >
                        {{ t('dedicatedAccounts.my.accountId', { id: acc.id }) }}
                      </span>
                      <span
                        v-if="acc.bought_out"
                        :data-testid="`bought-out-badge-${acc.id}`"
                        class="inline-flex items-center gap-1 rounded-full bg-rose-100 px-2.5 py-0.5 text-xs font-semibold text-rose-700 dark:bg-rose-900/40 dark:text-rose-300"
                      >
                        <Icon name="shield" size="xs" />
                        {{ t('dedicatedAccounts.my.boughtOutBadge') }}
                      </span>
                    </div>
                    <div class="mt-1 flex flex-wrap items-center gap-2 text-xs text-gray-500 dark:text-gray-400">
                      <span class="font-medium uppercase tracking-wide">{{ acc.platform }}</span>
                      <span>·</span>
                      <span>{{ t('dedicatedAccounts.my.concurrencyUnit', { count: acc.concurrency }) }}</span>
                      <template v-if="acc.group_name">
                        <span>·</span>
                        <span class="text-primary-600 dark:text-primary-400">{{ acc.group_name }}</span>
                      </template>
                    </div>
                  </div>
                </div>

                <!-- Health Pill -->
                <span :class="['shrink-0 rounded-full px-3 py-1 text-xs font-semibold', healthBadgeClass(acc)]">
                  {{ healthBadgeLabel(acc) }}
                </span>
              </div>

              <!-- Cooldown Notice (if rate-limited / overloaded) -->
              <div
                v-if="cooldownTargetTime(acc)"
                class="flex items-center gap-2 rounded-xl bg-amber-50 px-3.5 py-2 text-xs text-amber-800 dark:bg-amber-900/25 dark:text-amber-300"
              >
                <Icon name="clock" size="sm" class="shrink-0 text-amber-500" />
                <span>{{ t('dedicatedAccounts.my.cooldownUntil', { time: formatDateTime(cooldownTargetTime(acc)!) }) }}</span>
              </div>

              <!-- Key Metadata Strip -->
              <div class="grid grid-cols-3 gap-3 rounded-xl bg-gray-50 p-3 text-xs dark:bg-dark-900/60">
                <div>
                  <div class="text-gray-400 dark:text-gray-500">{{ t('dedicatedAccounts.my.expiresAt') }}</div>
                  <div class="mt-1 font-semibold text-gray-900 dark:text-white">
                    {{ acc.expires_at ? formatShortDate(acc.expires_at) : t('dedicatedAccounts.my.neverExpires') }}
                  </div>
                  <div
                    v-if="acc.expires_at"
                    :class="[
                      'mt-0.5 text-[11px]',
                      remainingDays(acc.expires_at) <= 3
                        ? 'text-red-500 font-medium'
                        : 'text-emerald-600 dark:text-emerald-400'
                    ]"
                  >
                    {{
                      remainingDays(acc.expires_at) >= 0
                        ? t('dedicatedAccounts.my.remainingDays', { days: remainingDays(acc.expires_at) })
                        : t('dedicatedAccounts.my.expiredDays')
                    }}
                  </div>
                </div>
                <div>
                  <div class="text-gray-400 dark:text-gray-500">
                    {{ acc.bought_out ? t('dedicatedAccounts.my.boughtOutAtLabel') : t('dedicatedAccounts.my.soldAt') }}
                  </div>
                  <div class="mt-1 font-semibold text-gray-900 dark:text-white">
                    {{
                      acc.bought_out && acc.bought_out_at
                        ? formatShortDate(acc.bought_out_at)
                        : acc.sold_at
                          ? formatShortDate(acc.sold_at)
                          : '-'
                    }}
                  </div>
                </div>
                <div>
                  <div class="text-gray-400 dark:text-gray-500">{{ t('dedicatedAccounts.my.lastUsedAt') }}</div>
                  <div class="mt-1 font-semibold text-gray-900 dark:text-white">
                    {{ acc.last_used_at ? formatRelativeOrTime(acc.last_used_at) : t('dedicatedAccounts.my.neverUsed') }}
                  </div>
                </div>
              </div>

              <!-- Buyout Action Strip: View Bought-Out Info OR Upgrade Buyout -->
              <div
                v-if="acc.bought_out"
                class="flex flex-col gap-2 rounded-xl border border-rose-200/80 bg-rose-50/60 p-3.5 sm:flex-row sm:items-center sm:justify-between dark:border-rose-800/50 dark:bg-rose-950/25"
              >
                <p class="text-xs leading-relaxed text-rose-800 dark:text-rose-300">
                  {{ t('dedicatedAccounts.buyout.noWarrantyFooter') }}
                </p>
                <button
                  type="button"
                  :data-testid="`view-buyout-info-${acc.id}`"
                  :disabled="loadingBuyoutInfoId === acc.id"
                  @click="handleViewBoughtOutInfo(acc)"
                  class="inline-flex shrink-0 items-center justify-center gap-1.5 rounded-lg bg-rose-600 px-3.5 py-2 text-xs font-semibold text-white shadow-sm transition-colors hover:bg-rose-700 disabled:opacity-60"
                >
                  <Icon name="key" size="xs" />
                  <span>
                    {{
                      loadingBuyoutInfoId === acc.id
                        ? t('dedicatedAccounts.my.loadingBuyoutInfo')
                        : t('dedicatedAccounts.my.viewBuyoutInfo')
                    }}
                  </span>
                </button>
              </div>
              <div
                v-else-if="acc.buyout_available && (acc.buyout_price_cny ?? 0) > 0"
                class="flex flex-col gap-2.5 rounded-xl border border-amber-200/80 bg-gradient-to-r from-amber-50/80 to-orange-50/50 p-3.5 sm:flex-row sm:items-center sm:justify-between dark:border-amber-800/50 dark:from-amber-950/30 dark:to-dark-900/50"
              >
                <p class="text-xs leading-relaxed text-amber-900/90 dark:text-amber-300/90">
                  {{ t('dedicatedAccounts.my.buyoutUpgradeHint') }}
                </p>
                <button
                  type="button"
                  :data-testid="`owned-buyout-btn-${acc.id}`"
                  @click="openOwnedBuyoutDialog(acc)"
                  class="inline-flex shrink-0 items-center justify-center gap-1.5 rounded-lg bg-gradient-to-r from-amber-500 to-orange-600 px-4 py-2 text-xs font-semibold text-white shadow-sm transition-all hover:from-amber-600 hover:to-orange-700"
                >
                  <Icon name="key" size="xs" />
                  <span>
                    {{
                      t('dedicatedAccounts.my.buyoutUpgradeBtn', {
                        price: Number(acc.buyout_price_cny ?? 0).toFixed(2)
                      })
                    }}
                  </span>
                </button>
              </div>

              <!-- 5H / 7D Remaining Quota Bars (Read-Only) -->
              <div class="rounded-xl border border-gray-100 bg-gray-50/70 p-4 dark:border-dark-700 dark:bg-dark-900/40">
                <div class="mb-3 flex items-center justify-between">
                  <span class="text-xs font-semibold uppercase tracking-wider text-gray-600 dark:text-gray-300">
                    {{ t('dedicatedAccounts.my.quota.title') }}
                  </span>
                  <!-- Conditional Reset Credit Badge -->
                  <span
                    v-if="hasUsableResetCredit(acc)"
                    class="inline-flex items-center gap-1 rounded-full bg-violet-100 px-2.5 py-0.5 text-xs font-semibold text-violet-700 dark:bg-violet-900/40 dark:text-violet-300"
                  >
                    <Icon name="sparkles" size="xs" />
                    {{ t('dedicatedAccounts.my.resetCredit.availableBadge', { count: acc.reset_credits?.available_count }) }}
                  </span>
                </div>

                <div
                  v-if="acc.quota?.five_hour || acc.quota?.seven_day"
                  class="space-y-3.5"
                >
                  <!-- 5H Window -->
                  <div v-if="acc.quota?.five_hour" class="space-y-1.5">
                    <div class="flex items-center justify-between text-xs">
                      <span class="font-medium text-gray-700 dark:text-gray-300">
                        {{ t('dedicatedAccounts.my.quota.fiveHour') }}
                      </span>
                      <div class="flex items-center gap-2 tabular-nums">
                        <span
                          v-if="acc.quota.five_hour.resets_at"
                          class="text-gray-400 dark:text-gray-500"
                        >
                          {{ formatResetCountdown(acc.quota.five_hour.resets_at) }}
                        </span>
                        <span :class="['font-bold', quotaTextColor(acc.quota.five_hour.remaining_percent)]">
                          {{ t('dedicatedAccounts.my.quota.remainingPercent', { percent: formatPercent(acc.quota.five_hour.remaining_percent) }) }}
                        </span>
                      </div>
                    </div>
                    <div class="h-2.5 w-full overflow-hidden rounded-full bg-gray-200 dark:bg-dark-700">
                      <div
                        :class="['h-full rounded-full transition-all duration-500', quotaBarColor(acc.quota.five_hour.remaining_percent)]"
                        :style="{ width: clampPercent(acc.quota.five_hour.remaining_percent) + '%' }"
                      />
                    </div>
                  </div>

                  <!-- 7D Window -->
                  <div v-if="acc.quota?.seven_day" class="space-y-1.5">
                    <div class="flex items-center justify-between text-xs">
                      <span class="font-medium text-gray-700 dark:text-gray-300">
                        {{ t('dedicatedAccounts.my.quota.sevenDay') }}
                      </span>
                      <div class="flex items-center gap-2 tabular-nums">
                        <span
                          v-if="acc.quota.seven_day.resets_at"
                          class="text-gray-400 dark:text-gray-500"
                        >
                          {{ formatResetCountdown(acc.quota.seven_day.resets_at) }}
                        </span>
                        <span :class="['font-bold', quotaTextColor(acc.quota.seven_day.remaining_percent)]">
                          {{ t('dedicatedAccounts.my.quota.remainingPercent', { percent: formatPercent(acc.quota.seven_day.remaining_percent) }) }}
                        </span>
                      </div>
                    </div>
                    <div class="h-2.5 w-full overflow-hidden rounded-full bg-gray-200 dark:bg-dark-700">
                      <div
                        :class="['h-full rounded-full transition-all duration-500', quotaBarColor(acc.quota.seven_day.remaining_percent)]"
                        :style="{ width: clampPercent(acc.quota.seven_day.remaining_percent) + '%' }"
                      />
                    </div>
                  </div>

                  <!-- 7D Sonnet Window (if present) -->
                  <div v-if="acc.quota?.seven_day_sonnet" class="space-y-1.5">
                    <div class="flex items-center justify-between text-xs">
                      <span class="font-medium text-gray-700 dark:text-gray-300">
                        {{ t('dedicatedAccounts.my.quota.sevenDaySonnet') }}
                      </span>
                      <span :class="['font-bold tabular-nums', quotaTextColor(acc.quota.seven_day_sonnet.remaining_percent)]">
                        {{ t('dedicatedAccounts.my.quota.remainingPercent', { percent: formatPercent(acc.quota.seven_day_sonnet.remaining_percent) }) }}
                      </span>
                    </div>
                    <div class="h-2 w-full overflow-hidden rounded-full bg-gray-200 dark:bg-dark-700">
                      <div
                        :class="['h-full rounded-full transition-all duration-500', quotaBarColor(acc.quota.seven_day_sonnet.remaining_percent)]"
                        :style="{ width: clampPercent(acc.quota.seven_day_sonnet.remaining_percent) + '%' }"
                      />
                    </div>
                  </div>
                </div>

                <p v-else class="text-xs text-gray-500 dark:text-gray-400">
                  {{ t('dedicatedAccounts.my.quota.notApplicable') }}
                </p>

                <!-- ONLY Permitted Mutation Action: Redeem Official Reset Credit (Hidden When 0 Credits) -->
                <div v-if="hasUsableResetCredit(acc)" class="mt-4 border-t border-gray-200/70 pt-3 dark:border-dark-700">
                  <button
                    type="button"
                    :data-testid="`redeem-reset-credit-${acc.id}`"
                    :disabled="redeemingId === acc.id"
                    @click="openResetCreditConfirm(acc)"
                    class="inline-flex w-full items-center justify-center gap-2 rounded-xl bg-gradient-to-r from-violet-600 to-indigo-600 px-4 py-2 text-xs font-semibold text-white shadow-sm transition-all hover:from-violet-700 hover:to-indigo-700 disabled:opacity-60"
                  >
                    <Icon name="refresh" size="sm" :class="{ 'animate-spin': redeemingId === acc.id }" />
                    <span>
                      {{
                        redeemingId === acc.id
                          ? t('dedicatedAccounts.my.resetCredit.redeeming')
                          : t('dedicatedAccounts.my.resetCredit.button', { count: acc.reset_credits?.available_count })
                      }}
                    </span>
                  </button>
                </div>
              </div>

              <!-- Bound Dedicated API Keys Summary -->
              <div class="space-y-2">
                <div class="flex items-center justify-between text-xs">
                  <span class="font-semibold text-gray-700 dark:text-gray-300">
                    {{ t('dedicatedAccounts.my.boundKeys.title') }}
                  </span>
                  <button
                    type="button"
                    @click="router.push('/keys')"
                    class="font-medium text-primary-600 hover:text-primary-700 dark:text-primary-400"
                  >
                    {{ t('dedicatedAccounts.my.boundKeys.manageKeys') }} →
                  </button>
                </div>

                <div v-if="acc.bound_api_keys && acc.bound_api_keys.length > 0" class="space-y-1.5">
                  <div
                    v-for="bk in acc.bound_api_keys"
                    :key="bk.id"
                    class="flex items-center justify-between rounded-lg border border-gray-100 bg-gray-50/80 px-3 py-2 text-xs dark:border-dark-700 dark:bg-dark-900/50"
                  >
                    <div class="flex items-center gap-2 min-w-0">
                      <Icon name="key" size="xs" class="shrink-0 text-primary-500" />
                      <span class="truncate font-medium text-gray-800 dark:text-gray-200">{{ bk.name }}</span>
                      <code class="shrink-0 rounded bg-gray-200/70 px-1.5 py-0.5 font-mono text-[11px] text-gray-600 dark:bg-dark-700 dark:text-gray-300">
                        {{ maskKeyPreview(bk.key) }}
                      </code>
                    </div>
                    <span
                      :class="[
                        'shrink-0 rounded px-2 py-0.5 text-[11px] font-medium',
                        bk.dedicated_pool_mode
                          ? 'bg-indigo-50 text-indigo-700 dark:bg-indigo-900/30 dark:text-indigo-300'
                          : 'bg-emerald-50 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300'
                      ]"
                    >
                      {{
                        bk.dedicated_pool_mode
                          ? t('dedicatedAccounts.my.boundKeys.poolBadge')
                          : t('dedicatedAccounts.my.boundKeys.pinnedBadge')
                      }}
                    </span>
                  </div>
                </div>
                <p v-else class="text-xs text-gray-400 dark:text-gray-500">
                  {{ t('dedicatedAccounts.my.boundKeys.noBoundKey') }}
                </p>
              </div>
            </div>

            <!-- Read-Only Protection Notice Footer -->
            <div class="mt-5 flex items-center gap-2 border-t border-gray-100 pt-3 text-[11px] text-gray-400 dark:border-dark-700 dark:text-gray-500">
              <Icon name="lock" size="xs" class="shrink-0" />
              <span>{{ t('dedicatedAccounts.my.quota.readOnlyNotice') }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- TAB 2: PRE-LISTED DEDICATED MARKETPLACE -->
      <div v-else>
        <div v-if="loadingMarket" class="grid gap-5 lg:grid-cols-2 2xl:grid-cols-3">
          <div
            v-for="idx in 3"
            :key="idx"
            class="h-80 animate-pulse rounded-2xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-800"
          />
        </div>

        <div
          v-else-if="marketItems.length === 0"
          class="flex flex-col items-center justify-center rounded-2xl border border-dashed border-gray-300 bg-white px-6 py-16 text-center dark:border-dark-700 dark:bg-dark-800/60"
        >
          <div
            class="flex h-14 w-14 items-center justify-center rounded-2xl bg-amber-50 text-amber-600 dark:bg-amber-900/30 dark:text-amber-400"
          >
            <Icon name="shoppingBag" size="lg" />
          </div>
          <h3 class="mt-4 text-base font-semibold text-gray-900 dark:text-white">
            {{ t('dedicatedAccounts.market.emptyTitle') }}
          </h3>
          <p class="mt-1.5 max-w-md text-sm text-gray-500 dark:text-gray-400">
            {{ t('dedicatedAccounts.market.emptyDesc') }}
          </p>
        </div>

        <div v-else class="grid gap-6 lg:grid-cols-2 2xl:grid-cols-3">
          <div
            v-for="item in marketItems"
            :key="item.id"
            :data-testid="`market-dedicated-card-${item.id}`"
            class="flex flex-col justify-between rounded-2xl border border-gray-200 bg-white p-6 shadow-sm transition-all hover:-translate-y-0.5 hover:shadow-md dark:border-dark-700 dark:bg-dark-800"
          >
            <div class="space-y-4">
              <!-- Top Ribbon + Platform -->
              <div class="space-y-2.5">
                <div class="flex items-center justify-between gap-2">
                  <div class="flex items-center gap-2 text-xs text-gray-500 dark:text-gray-400">
                    <div
                      class="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl bg-primary-50 text-primary-600 dark:bg-primary-900/30 dark:text-primary-400"
                    >
                      <PlatformIcon :platform="item.platform" size="md" />
                    </div>
                    <span class="uppercase font-semibold tracking-wide text-gray-700 dark:text-gray-300">{{ item.platform }}</span>
                    <span>·</span>
                    <span class="whitespace-nowrap">{{ t('dedicatedAccounts.market.concurrency', { count: item.concurrency }) }}</span>
                  </div>
                  <span
                    class="shrink-0 whitespace-nowrap rounded-full bg-emerald-50 px-2.5 py-1 text-[11px] font-semibold text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300"
                  >
                    {{ t('dedicatedAccounts.market.stockBadge') }}
                  </span>
                </div>
                <h3 class="text-base font-bold leading-snug text-gray-900 dark:text-white">
                  {{ item.display_name }}
                </h3>
              </div>

              <!-- Validity & Group & Buyout Tag -->
              <div class="flex flex-wrap items-center gap-2 text-xs">
                <span
                  class="inline-flex items-center gap-1 rounded-lg bg-gray-100 px-2.5 py-1 font-medium text-gray-700 dark:bg-dark-700 dark:text-gray-200"
                >
                  <Icon name="calendar" size="xs" />
                  <template v-if="item.remaining_days != null">
                    {{ t('dedicatedAccounts.market.validityDays', { days: item.remaining_days }) }}
                  </template>
                  <template v-else>
                    {{ t('dedicatedAccounts.market.longTerm') }}
                  </template>
                </span>
                <span
                  v-if="item.group_name"
                  class="inline-flex items-center gap-1 rounded-lg bg-blue-50 px-2.5 py-1 font-medium text-blue-700 dark:bg-blue-900/30 dark:text-blue-300"
                >
                  {{ item.group_name }}
                </span>
                <span
                  v-if="item.buyout_available && (item.buyout_price_cny ?? 0) > 0"
                  class="inline-flex items-center gap-1 rounded-lg bg-amber-100/90 px-2.5 py-1 font-semibold text-amber-800 dark:bg-amber-900/40 dark:text-amber-300"
                >
                  <Icon name="key" size="xs" />
                  {{
                    t('dedicatedAccounts.buyout.availableTag', {
                      price: Number(item.buyout_price_cny ?? 0).toFixed(2)
                    })
                  }}
                </span>
              </div>

              <!-- Optional Seller Highlights -->
              <p
                v-if="item.description"
                class="rounded-xl bg-gray-50 px-3 py-2 text-xs leading-relaxed text-gray-600 dark:bg-dark-900/50 dark:text-gray-300"
              >
                {{ item.description }}
              </p>

              <!-- Quota Preview -->
              <div
                v-if="item.quota_preview?.five_hour || item.quota_preview?.seven_day"
                class="space-y-2 rounded-xl border border-gray-100 bg-gray-50/60 p-3 dark:border-dark-700 dark:bg-dark-900/30"
              >
                <div class="text-[11px] font-semibold text-gray-500 dark:text-gray-400">
                  {{ t('dedicatedAccounts.market.quotaPreviewTitle') }}
                </div>
                <div v-if="item.quota_preview?.five_hour" class="flex items-center justify-between text-xs">
                  <span class="text-gray-600 dark:text-gray-400">5H</span>
                  <div class="flex items-center gap-2 w-2/3">
                    <div class="h-1.5 flex-1 overflow-hidden rounded-full bg-gray-200 dark:bg-dark-700">
                      <div
                        :class="['h-full rounded-full', quotaBarColor(item.quota_preview.five_hour.remaining_percent)]"
                        :style="{ width: clampPercent(item.quota_preview.five_hour.remaining_percent) + '%' }"
                      />
                    </div>
                    <span class="w-16 text-right font-semibold tabular-nums text-gray-800 dark:text-gray-200">
                      {{ formatPercent(item.quota_preview.five_hour.remaining_percent) }}%
                    </span>
                  </div>
                </div>
                <div v-if="item.quota_preview?.seven_day" class="flex items-center justify-between text-xs">
                  <span class="text-gray-600 dark:text-gray-400">7D</span>
                  <div class="flex items-center gap-2 w-2/3">
                    <div class="h-1.5 flex-1 overflow-hidden rounded-full bg-gray-200 dark:bg-dark-700">
                      <div
                        :class="['h-full rounded-full', quotaBarColor(item.quota_preview.seven_day.remaining_percent)]"
                        :style="{ width: clampPercent(item.quota_preview.seven_day.remaining_percent) + '%' }"
                      />
                    </div>
                    <span class="w-16 text-right font-semibold tabular-nums text-gray-800 dark:text-gray-200">
                      {{ formatPercent(item.quota_preview.seven_day.remaining_percent) }}%
                    </span>
                  </div>
                </div>
              </div>

              <!-- Price & Daily Decay Box -->
              <div
                class="rounded-xl border border-amber-200/70 bg-gradient-to-br from-amber-50/70 to-orange-50/40 p-3.5 dark:border-amber-800/40 dark:from-amber-950/30 dark:to-dark-900/60"
              >
                <div class="flex items-start justify-between gap-3">
                  <div class="min-w-0">
                    <span class="text-xs text-gray-500 dark:text-gray-400">
                      {{ t('dedicatedAccounts.market.priceLabel') }}
                    </span>
                    <div class="mt-0.5 flex flex-wrap items-baseline gap-2">
                      <span class="text-2xl font-extrabold tabular-nums text-gray-900 dark:text-white">
                        ¥{{ item.current_price_cny.toFixed(2) }}
                      </span>
                      <span
                        v-if="item.current_price_cny < item.base_price_cny"
                        class="whitespace-nowrap text-xs text-gray-400 line-through tabular-nums"
                      >
                        {{ t('dedicatedAccounts.market.basePriceLabel', { price: item.base_price_cny.toFixed(2) }) }}
                      </span>
                    </div>
                  </div>
                  <span
                    v-if="item.daily_price_decay_cny > 0"
                    class="shrink-0 whitespace-nowrap rounded-lg bg-amber-500/15 px-2.5 py-1 text-[11px] font-semibold text-amber-800 dark:bg-amber-500/20 dark:text-amber-300"
                  >
                    {{ t('dedicatedAccounts.market.decayBadge', { decay: item.daily_price_decay_cny.toFixed(2) }) }}
                  </span>
                </div>

                <div v-if="item.daily_price_decay_cny > 0" class="mt-2 space-y-0.5 text-[11px] text-amber-800/80 dark:text-amber-300/80">
                  <div>
                    {{
                      t('dedicatedAccounts.market.decayFloorHint', {
                        decay: item.daily_price_decay_cny.toFixed(2),
                        min: item.min_sale_price_cny.toFixed(2)
                      })
                    }}
                  </div>
                  <div v-if="item.next_price_decay_at" class="font-medium">
                    {{ t('dedicatedAccounts.market.nextDropCountdown', { time: formatCountdownOnly(item.next_price_decay_at) }) }}
                  </div>
                  <div v-else-if="item.current_price_cny <= item.min_sale_price_cny" class="font-medium">
                    {{ t('dedicatedAccounts.market.atMinPrice') }}
                  </div>
                </div>
                <div v-else class="mt-1.5 text-[11px] text-gray-500 dark:text-gray-400">
                  {{ t('dedicatedAccounts.market.fixedPriceHint') }}
                </div>
              </div>
            </div>

            <!-- Action Buttons -->
            <div class="mt-5 flex flex-col gap-2">
              <button
                type="button"
                :data-testid="`buy-balance-btn-${item.id}`"
                @click="openBalanceBuyDialog(item)"
                class="btn btn-primary w-full justify-center py-2.5 text-sm font-semibold"
              >
                {{ t('dedicatedAccounts.market.buyWithBalance') }}
              </button>
              <button
                v-if="paymentEnabled"
                type="button"
                :data-testid="`buy-online-btn-${item.id}`"
                @click="openOnlinePayDialog(item)"
                class="btn btn-secondary w-full justify-center py-2 text-xs font-medium"
              >
                {{ t('dedicatedAccounts.market.buyOnline') }}
              </button>
              <button
                v-if="item.buyout_available && (item.buyout_price_cny ?? 0) > 0"
                type="button"
                :data-testid="`market-buyout-btn-${item.id}`"
                @click="openMarketBuyoutDialog(item)"
                class="inline-flex w-full items-center justify-center gap-1.5 rounded-xl bg-gradient-to-r from-amber-500 to-orange-600 px-4 py-2.5 text-xs font-semibold text-white shadow-sm transition-all hover:from-amber-600 hover:to-orange-700"
              >
                <Icon name="key" size="xs" />
                <span>
                  {{
                    t('dedicatedAccounts.buyout.marketBtn', {
                      price: Number(item.buyout_price_cny ?? 0).toFixed(2)
                    })
                  }}
                </span>
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Confirm Redeem Official Reset Credit Dialog -->
    <ConfirmDialog
      :show="!!resetConfirmTarget"
      :title="t('dedicatedAccounts.my.resetCredit.confirmTitle')"
      :message="
        resetConfirmTarget
          ? t('dedicatedAccounts.my.resetCredit.confirmMessage', {
              alias: resetConfirmTarget.display_name,
              count: resetConfirmTarget.reset_credits?.available_count ?? 0
            })
          : ''
      "
      :confirm-text="t('dedicatedAccounts.my.resetCredit.confirmAction')"
      @confirm="confirmRedeemResetCredit"
      @cancel="resetConfirmTarget = null"
    />

    <!-- Balance Activation Confirmation Dialog -->
    <BaseDialog
      :show="!!balanceBuyTarget"
      :title="t('dedicatedAccounts.market.confirmBalanceTitle')"
      width="normal"
      @close="balanceBuyTarget = null"
    >
      <div v-if="balanceBuyTarget" class="space-y-4">
        <p class="text-sm text-gray-600 dark:text-gray-300">
          {{ t('dedicatedAccounts.market.confirmBalanceDesc') }}
        </p>

        <div class="space-y-2.5 rounded-xl bg-gray-50 p-4 text-sm dark:bg-dark-900/60">
          <div class="flex justify-between">
            <span class="text-gray-500 dark:text-gray-400">{{ t('dedicatedAccounts.market.confirmItemAlias') }}</span>
            <span class="font-semibold text-gray-900 dark:text-white">{{ balanceBuyTarget.display_name }} ({{ balanceBuyTarget.platform }})</span>
          </div>
          <div class="flex justify-between">
            <span class="text-gray-500 dark:text-gray-400">{{ t('dedicatedAccounts.market.confirmItemValidity') }}</span>
            <span class="font-medium text-gray-900 dark:text-white">
              {{
                balanceBuyTarget.expires_at
                  ? t('dedicatedAccounts.market.validityUntil', { date: formatShortDate(balanceBuyTarget.expires_at) })
                  : t('dedicatedAccounts.market.longTerm')
              }}
            </span>
          </div>
          <div class="flex justify-between border-t border-gray-200 pt-2.5 dark:border-dark-700">
            <span class="text-gray-500 dark:text-gray-400">{{ t('dedicatedAccounts.market.confirmItemPrice') }}</span>
            <span class="text-lg font-extrabold text-primary-600 dark:text-primary-400">
              ¥{{ balanceBuyTarget.current_price_cny.toFixed(2) }}
            </span>
          </div>
        </div>

        <div
          :class="[
            'rounded-xl px-3.5 py-2.5 text-xs font-medium',
            userBalance >= requiredBalanceForItem(balanceBuyTarget)
              ? 'bg-emerald-50 text-emerald-700 dark:bg-emerald-900/25 dark:text-emerald-300'
              : 'bg-red-50 text-red-600 dark:bg-red-900/25 dark:text-red-300'
          ]"
        >
          {{
            userBalance >= requiredBalanceForItem(balanceBuyTarget)
              ? t('dedicatedAccounts.market.currentBalance', { balance: userBalance.toFixed(2) })
              : t('dedicatedAccounts.market.insufficientBalance', {
                  balance: userBalance.toFixed(2),
                  price: requiredBalanceForItem(balanceBuyTarget).toFixed(2)
                })
          }}
        </div>
      </div>

      <template #footer>
        <div class="flex justify-end gap-3">
          <button type="button" class="btn btn-secondary" :disabled="purchasingBalance" @click="balanceBuyTarget = null">
            {{ t('common.cancel') }}
          </button>
          <button
            v-if="balanceBuyTarget"
            type="button"
            data-testid="confirm-balance-buy-btn"
            class="btn btn-primary"
            :disabled="purchasingBalance || userBalance < requiredBalanceForItem(balanceBuyTarget)"
            @click="confirmBalanceBuyout"
          >
            {{
              purchasingBalance
                ? t('dedicatedAccounts.market.purchasing')
                : t('dedicatedAccounts.market.confirmBalanceAction', {
                    price: balanceBuyTarget.current_price_cny.toFixed(2)
                  })
            }}
          </button>
        </div>
      </template>
    </BaseDialog>

    <!-- Full Credentials Buyout Confirmation Dialog (Shelf or Owned Upgrade) -->
    <BaseDialog
      :show="!!buyoutTarget"
      :title="t('dedicatedAccounts.buyout.confirmTitle')"
      width="normal"
      @close="closeBuyoutDialog"
    >
      <div v-if="buyoutTarget" class="space-y-4">
        <p class="text-sm text-gray-600 dark:text-gray-300">
          {{
            buyoutTarget.source === 'owned'
              ? t('dedicatedAccounts.buyout.confirmDescOwned')
              : t('dedicatedAccounts.buyout.confirmDescShelf')
          }}
        </p>

        <div class="space-y-2.5 rounded-xl bg-gray-50 p-4 text-sm dark:bg-dark-900/60">
          <div class="flex justify-between">
            <span class="text-gray-500 dark:text-gray-400">{{ t('dedicatedAccounts.market.confirmItemAlias') }}</span>
            <span class="font-semibold text-gray-900 dark:text-white">
              {{ buyoutTarget.displayName }} ({{ buyoutTarget.platform }})
            </span>
          </div>
          <div class="flex justify-between border-t border-gray-200 pt-2.5 dark:border-dark-700">
            <span class="text-gray-500 dark:text-gray-400">{{ t('dedicatedAccounts.buyout.priceLabel') }}</span>
            <span class="text-lg font-extrabold text-amber-600 dark:text-amber-400">
              ¥{{ buyoutTarget.priceCny.toFixed(2) }}
            </span>
          </div>
        </div>

        <!-- Explicit Post-Buyout Warranty Waiver Notice -->
        <div class="rounded-xl border border-rose-200 bg-rose-50/80 p-3.5 text-xs dark:border-rose-800/60 dark:bg-rose-950/30">
          <div class="flex items-center gap-1.5 font-semibold text-rose-800 dark:text-rose-300">
            <Icon name="exclamationTriangle" size="sm" class="shrink-0 text-rose-500" />
            <span>{{ t('dedicatedAccounts.buyout.noWarrantyTitle') }}</span>
          </div>
          <p class="mt-1.5 leading-relaxed text-rose-700/90 dark:text-rose-300/90">
            {{ t('dedicatedAccounts.buyout.noWarrantyDesc') }}
          </p>
        </div>

        <label class="flex cursor-pointer items-start gap-2.5 rounded-xl border border-gray-200 bg-white p-3 text-xs text-gray-700 dark:border-dark-700 dark:bg-dark-800 dark:text-gray-200">
          <input
            v-model="buyoutAckChecked"
            type="checkbox"
            data-testid="buyout-ack-checkbox"
            class="mt-0.5 rounded border-gray-300 text-amber-600 focus:ring-amber-500"
          />
          <span class="leading-relaxed font-medium">{{ t('dedicatedAccounts.buyout.ackCheckbox') }}</span>
        </label>

        <div
          :class="[
            'rounded-xl px-3.5 py-2.5 text-xs font-medium',
            userBalance >= buyoutTarget.priceBalance
              ? 'bg-emerald-50 text-emerald-700 dark:bg-emerald-900/25 dark:text-emerald-300'
              : 'bg-amber-50 text-amber-800 dark:bg-amber-900/25 dark:text-amber-300'
          ]"
        >
          {{
            userBalance >= buyoutTarget.priceBalance
              ? t('dedicatedAccounts.market.currentBalance', { balance: userBalance.toFixed(2) })
              : t('dedicatedAccounts.market.insufficientBalance', {
                  balance: userBalance.toFixed(2),
                  price: buyoutTarget.priceBalance.toFixed(2)
                })
          }}
        </div>
      </div>

      <template #footer>
        <div class="flex flex-wrap justify-end gap-2.5">
          <button type="button" class="btn btn-secondary" :disabled="processingBuyout" @click="closeBuyoutDialog">
            {{ t('common.cancel') }}
          </button>
          <button
            v-if="buyoutTarget && paymentEnabled"
            type="button"
            data-testid="buyout-online-pay-btn"
            class="btn btn-secondary"
            :disabled="processingBuyout || !buyoutAckChecked"
            @click="switchBuyoutToOnlinePay"
          >
            {{
              t('dedicatedAccounts.buyout.payOnline', {
                price: buyoutTarget.priceCny.toFixed(2)
              })
            }}
          </button>
          <button
            v-if="buyoutTarget"
            type="button"
            data-testid="confirm-buyout-balance-btn"
            class="inline-flex items-center justify-center rounded-xl bg-gradient-to-r from-amber-500 to-orange-600 px-4 py-2 text-sm font-semibold text-white shadow-sm transition-all hover:from-amber-600 hover:to-orange-700 disabled:cursor-not-allowed disabled:opacity-50"
            :disabled="processingBuyout || !buyoutAckChecked || userBalance < buyoutTarget.priceBalance"
            @click="confirmFullBuyoutWithBalance"
          >
            {{
              processingBuyout
                ? t('dedicatedAccounts.buyout.processing')
                : t('dedicatedAccounts.buyout.payWithBalance', {
                    price: buyoutTarget.priceCny.toFixed(2)
                  })
            }}
          </button>
        </div>
      </template>
    </BaseDialog>

    <!-- Bought-Out Plaintext Account Info Modal -->
    <BaseDialog
      :show="!!buyoutInfoModal"
      :title="t('dedicatedAccounts.buyout.infoModalTitle')"
      width="normal"
      @close="closeBuyoutInfoModal"
    >
      <div v-if="buyoutInfoModal" class="space-y-4">
        <p class="text-sm text-gray-600 dark:text-gray-300">
          {{ t('dedicatedAccounts.buyout.infoModalSubtitle', { alias: buyoutInfoModal.displayName }) }}
        </p>

        <!-- Plaintext Buyout Info Box -->
        <div class="space-y-2">
          <div class="flex items-center justify-between">
            <label class="block text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400">
              {{ t('dedicatedAccounts.buyout.infoLabel') }}
            </label>
            <button
              type="button"
              data-testid="copy-buyout-info-btn"
              @click="copyBuyoutPlaintext(buyoutInfoModal.buyoutInfo)"
              class="btn btn-primary btn-sm"
            >
              {{
                copiedBuyoutInfo
                  ? t('dedicatedAccounts.buyout.copiedInfo')
                  : t('dedicatedAccounts.buyout.copyInfo')
              }}
            </button>
          </div>
          <pre
            data-testid="revealed-buyout-info"
            class="max-h-64 overflow-y-auto whitespace-pre-wrap break-all rounded-xl border border-amber-200 bg-amber-50/50 p-4 font-mono text-xs leading-relaxed text-gray-900 dark:border-amber-800/50 dark:bg-dark-900 dark:text-gray-100"
          >{{ buyoutInfoModal.buyoutInfo }}</pre>
        </div>

        <!-- Also show newly generated 1:1 API Key if bought directly from Marketplace -->
        <div v-if="buyoutInfoModal.apiKey" class="space-y-2">
          <label class="block text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400">
            {{ t('dedicatedAccounts.market.purchaseResultModal.keyLabel') }}
          </label>
          <div class="flex items-center gap-2 rounded-xl border border-emerald-200 bg-emerald-50/60 p-3 dark:border-emerald-800/50 dark:bg-emerald-950/30">
            <code
              class="flex-1 break-all font-mono text-xs font-semibold text-emerald-950 dark:text-emerald-200"
            >
              {{ buyoutInfoModal.apiKey }}
            </code>
            <button
              type="button"
              @click="copyFullApiKey(buyoutInfoModal.apiKey)"
              class="btn btn-secondary btn-sm shrink-0"
            >
              {{
                copiedNewKey
                  ? t('dedicatedAccounts.market.purchaseResultModal.copied')
                  : t('dedicatedAccounts.market.purchaseResultModal.copyKey')
              }}
            </button>
          </div>
        </div>

        <div class="rounded-xl bg-rose-50 p-3 text-xs leading-relaxed text-rose-800 dark:bg-rose-900/25 dark:text-rose-300">
          {{ t('dedicatedAccounts.buyout.noWarrantyFooter') }}
        </div>
      </div>

      <template #footer>
        <div class="flex justify-end">
          <button type="button" class="btn btn-primary" @click="closeBuyoutInfoModal">
            {{ t('common.confirm') }}
          </button>
        </div>
      </template>
    </BaseDialog>

    <!-- Post-Purchase Newly Generated 1:1 API Key Modal -->
    <BaseDialog
      :show="!!purchaseResult"
      :title="t('dedicatedAccounts.market.purchaseResultModal.title')"
      width="normal"
      @close="closePurchaseResultModal"
    >
      <div v-if="purchaseResult" class="space-y-4">
        <p class="text-sm text-gray-600 dark:text-gray-300">
          {{ t('dedicatedAccounts.market.purchaseResultModal.subtitle', { alias: purchaseResult.account.display_name }) }}
        </p>

        <div v-if="purchaseResult.api_key" class="space-y-2">
          <label class="block text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400">
            {{ t('dedicatedAccounts.market.purchaseResultModal.keyLabel') }}
          </label>
          <div class="flex items-center gap-2 rounded-xl border border-emerald-200 bg-emerald-50/60 p-3 dark:border-emerald-800/50 dark:bg-emerald-950/30">
            <code
              data-testid="created-dedicated-api-key"
              class="flex-1 break-all font-mono text-xs font-semibold text-emerald-950 dark:text-emerald-200"
            >
              {{ purchaseResult.api_key }}
            </code>
            <button
              type="button"
              @click="copyFullApiKey(purchaseResult.api_key)"
              class="btn btn-primary btn-sm shrink-0"
            >
              {{
                copiedNewKey
                  ? t('dedicatedAccounts.market.purchaseResultModal.copied')
                  : t('dedicatedAccounts.market.purchaseResultModal.copyKey')
              }}
            </button>
          </div>
        </div>

        <div class="rounded-xl bg-amber-50 p-3.5 text-xs leading-relaxed text-amber-800 dark:bg-amber-900/25 dark:text-amber-300">
          {{ t('dedicatedAccounts.market.purchaseResultModal.pinningHint') }}
        </div>
      </div>

      <template #footer>
        <div class="flex justify-end">
          <button type="button" class="btn btn-primary" @click="closePurchaseResultModal">
            {{ t('dedicatedAccounts.market.purchaseResultModal.viewMyAccount') }}
          </button>
        </div>
      </template>
    </BaseDialog>

    <!-- Online Payment Method Dialog (Supports both Activation and Full Buyout) -->
    <BaseDialog
      :show="!!onlinePayTarget"
      :title="
        onlinePayTarget?.orderType === 'dedicated_buyout'
          ? t('dedicatedAccounts.buyout.onlinePayTitle')
          : t('dedicatedAccounts.market.onlinePayTitle')
      "
      width="normal"
      @close="onlinePayTarget = null"
    >
      <div v-if="onlinePayTarget" class="space-y-4">
        <p class="text-sm text-gray-600 dark:text-gray-300">
          {{
            onlinePayTarget.orderType === 'dedicated_buyout'
              ? t('dedicatedAccounts.buyout.onlinePayDesc')
              : t('dedicatedAccounts.market.onlinePayDesc')
          }}
        </p>

        <div class="flex items-center justify-between rounded-xl bg-gray-50 p-3.5 text-sm dark:bg-dark-900/60">
          <span class="font-medium text-gray-700 dark:text-gray-200">{{ onlinePayTarget.displayName }}</span>
          <span class="text-lg font-extrabold text-primary-600 dark:text-primary-400">
            ¥{{ onlinePayTarget.priceCny.toFixed(2) }}
          </span>
        </div>

        <div v-if="loadingCheckout" class="flex justify-center py-6">
          <div class="h-6 w-6 animate-spin rounded-full border-2 border-primary-500 border-t-transparent" />
        </div>
        <PaymentMethodSelector
          v-else-if="onlinePaymentMethods.length > 0"
          :methods="onlinePaymentMethods"
          :selected="selectedPaymentMethod"
          @select="selectedPaymentMethod = $event"
        />
        <p v-else class="rounded-xl bg-amber-50 p-3 text-xs text-amber-700 dark:bg-amber-900/25 dark:text-amber-300">
          {{ t('dedicatedAccounts.market.noPaymentMethod') }}
        </p>
      </div>

      <template #footer>
        <div class="flex justify-end gap-3">
          <button type="button" class="btn btn-secondary" :disabled="creatingOnlineOrder" @click="onlinePayTarget = null">
            {{ t('common.cancel') }}
          </button>
          <button
            v-if="onlinePayTarget && onlinePaymentMethods.length > 0"
            type="button"
            class="btn btn-primary"
            :disabled="creatingOnlineOrder || !selectedPaymentMethod"
            @click="submitOnlineOrder"
          >
            {{
              creatingOnlineOrder
                ? t('dedicatedAccounts.market.creatingOrder')
                : t('dedicatedAccounts.market.payNow', { price: onlinePayTarget.priceCny.toFixed(2) })
            }}
          </button>
        </div>
      </template>
    </BaseDialog>

    <!-- QR / Popup Waiting Dialog for Online Activation or Buyout -->
    <PaymentQRDialog
      :show="qrDialog.show"
      :order-id="qrDialog.orderId"
      :qr-code="qrDialog.qrCode"
      :expires-at="qrDialog.expiresAt"
      :payment-type="qrDialog.paymentType"
      :pay-url="qrDialog.payUrl"
      @close="qrDialog.show = false"
      @success="handleOnlinePaySuccess"
    />
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import AppLayout from '@/components/layout/AppLayout.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import PlatformIcon from '@/components/common/PlatformIcon.vue'
import Icon from '@/components/icons/Icon.vue'
import PaymentMethodSelector, { type PaymentMethodOption } from '@/components/payment/PaymentMethodSelector.vue'
import PaymentQRDialog from '@/components/payment/PaymentQRDialog.vue'
import { buildCreateOrderPayload, getVisibleMethods, normalizeVisibleMethod } from '@/components/payment/paymentFlow'
import { getPaymentPopupFeatures } from '@/components/payment/providerConfig'
import { dedicatedAccountsAPI } from '@/api/dedicatedAccounts'
import { paymentAPI } from '@/api/payment'
import { useAuthStore, useAppStore } from '@/stores'
import { useClipboard } from '@/composables/useClipboard'
import { extractApiErrorMessage } from '@/utils/apiError'
import type {
  UserDedicatedAccountView,
  DedicatedMarketItem,
  DedicatedPurchaseResult,
  DedicatedQuotaSnapshot
} from '@/types'
import type { CheckoutInfoResponse } from '@/types/payment'

interface BuyoutDialogState {
  accountId: number
  displayName: string
  platform: string
  priceCny: number
  priceBalance: number
  source: 'owned' | 'shelf'
}

interface BuyoutInfoModalState {
  accountId: number
  displayName: string
  buyoutInfo: string
  apiKey?: string
}

interface OnlineOrderTargetState {
  accountId: number
  displayName: string
  priceCny: number
  orderType: 'dedicated_account' | 'dedicated_buyout'
}

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const appStore = useAppStore()
const { copyToClipboard } = useClipboard()

const activeTab = ref<'my' | 'market'>(route.query.tab === 'market' ? 'market' : 'my')
const marketPlatform = ref<string>('')

const myAccounts = ref<UserDedicatedAccountView[]>([])
const marketItems = ref<DedicatedMarketItem[]>([])
const loadingMy = ref(false)
const loadingMarket = ref(false)
const refreshingUsage = ref(false)

// Live clock tick for reset & decay countdowns
const nowTick = ref(Date.now())
let tickTimer: ReturnType<typeof setInterval> | null = null

// Reset credit redemption state
const resetConfirmTarget = ref<UserDedicatedAccountView | null>(null)
const redeemingId = ref<number | null>(null)

// Standard dedicated activation state
const balanceBuyTarget = ref<DedicatedMarketItem | null>(null)
const purchasingBalance = ref(false)
const purchaseResult = ref<DedicatedPurchaseResult | null>(null)
const copiedNewKey = ref(false)

// Full credentials buyout state
const buyoutTarget = ref<BuyoutDialogState | null>(null)
const buyoutAckChecked = ref(false)
const processingBuyout = ref(false)
const loadingBuyoutInfoId = ref<number | null>(null)
const buyoutInfoModal = ref<BuyoutInfoModalState | null>(null)
const copiedBuyoutInfo = ref(false)

// Online payment state
const onlinePayTarget = ref<OnlineOrderTargetState | null>(null)
const checkoutInfo = ref<CheckoutInfoResponse | null>(null)
const loadingCheckout = ref(false)
const selectedPaymentMethod = ref('')
const creatingOnlineOrder = ref(false)
const qrDialog = reactive({
  show: false,
  orderId: 0,
  qrCode: '',
  expiresAt: '',
  paymentType: '',
  payUrl: '',
  orderType: 'dedicated_account' as 'dedicated_account' | 'dedicated_buyout',
  accountId: 0
})

const userBalance = computed(() => Number(authStore.user?.balance ?? 0))
const paymentEnabled = computed(() => Boolean(appStore.cachedPublicSettings?.payment_enabled))

function requiredBalanceForItem(item: DedicatedMarketItem): number {
  const b = Number(item.current_price_balance)
  if (Number.isFinite(b) && b > 0) return b
  return Number(item.current_price_cny || 0)
}

const platformFilters = computed(() => [
  { value: '', label: t('dedicatedAccounts.market.filterAll') },
  { value: 'anthropic', label: t('dedicatedAccounts.market.filterAnthropic') },
  { value: 'openai', label: t('dedicatedAccounts.market.filterOpenAI') },
  { value: 'gemini', label: t('dedicatedAccounts.market.filterGemini') }
])

const onlinePaymentMethods = computed<PaymentMethodOption[]>(() => {
  if (!checkoutInfo.value || !onlinePayTarget.value) return []
  const visibleMap = getVisibleMethods(checkoutInfo.value.methods || {})
  const price = onlinePayTarget.value.priceCny
  return Object.entries(visibleMap).map(([mType, limits]) => {
    const feeRate = limits?.fee_rate ?? 0
    const minOk = !limits?.single_min || price >= limits.single_min
    const maxOk = !limits?.single_max || limits.single_max <= 0 || price <= limits.single_max
    const available = (limits ? limits.available : true) && minOk && maxOk
    return {
      type: mType,
      display_name: limits?.display_name,
      fee_rate: feeRate,
      available
    }
  })
})

function switchTab(tab: 'my' | 'market') {
  activeTab.value = tab
  const nextQuery: Record<string, string | null | (string | null)[] | undefined> = {
    ...route.query,
    tab
  }
  delete nextQuery.buyoutAccountId
  router.replace({ query: nextQuery }).catch(() => {})
  if (tab === 'market') {
    void loadMarketplace()
  } else {
    void loadMyAccounts()
  }
}

function selectPlatformFilter(platform: string) {
  marketPlatform.value = platform
  void loadMarketplace()
}

async function loadMyAccounts() {
  loadingMy.value = true
  try {
    myAccounts.value = await dedicatedAccountsAPI.listMyAccounts()
  } catch (err) {
    appStore.showError(extractApiErrorMessage(err, t('common.error')))
  } finally {
    loadingMy.value = false
  }
}

async function handleRefreshMyUsage() {
  refreshingUsage.value = true
  try {
    const ids = myAccounts.value.map((acc) => acc.id)
    const [freshAccounts, usageMap] = await Promise.all([
      dedicatedAccountsAPI.listMyAccounts(),
      ids.length > 0
        ? dedicatedAccountsAPI.getMyAccountsUsage(ids, true)
        : Promise.resolve({} as Record<string, DedicatedQuotaSnapshot>)
    ])
    myAccounts.value = freshAccounts.map((acc) => {
      const freshQuota = usageMap[String(acc.id)]
      return freshQuota ? { ...acc, quota: freshQuota } : acc
    })
    appStore.showSuccess(t('dedicatedAccounts.my.quotaRefreshed'))
  } catch (err) {
    appStore.showError(extractApiErrorMessage(err, t('common.error')))
  } finally {
    refreshingUsage.value = false
  }
}

async function loadMarketplace() {
  loadingMarket.value = true
  try {
    marketItems.value = await dedicatedAccountsAPI.listMarket(marketPlatform.value || undefined)
  } catch (err) {
    appStore.showError(extractApiErrorMessage(err, t('common.error')))
  } finally {
    loadingMarket.value = false
  }
}

function hasUsableResetCredit(acc: UserDedicatedAccountView): boolean {
  return Boolean(acc.reset_credits && acc.reset_credits.available_count > 0 && acc.reset_credits.redeemable)
}

function openResetCreditConfirm(acc: UserDedicatedAccountView) {
  if (!hasUsableResetCredit(acc)) return
  resetConfirmTarget.value = acc
}

async function confirmRedeemResetCredit() {
  const target = resetConfirmTarget.value
  if (!target) return
  resetConfirmTarget.value = null
  redeemingId.value = target.id
  try {
    const res = await dedicatedAccountsAPI.redeemResetCredit(target.id)
    const remaining = res.reset_credits?.available_count ?? 0
    appStore.showSuccess(
      t('dedicatedAccounts.my.resetCredit.success', { remaining })
    )
    await loadMyAccounts()
  } catch (err) {
    appStore.showError(extractApiErrorMessage(err, t('dedicatedAccounts.my.resetCredit.failed')))
  } finally {
    redeemingId.value = null
  }
}

function openBalanceBuyDialog(item: DedicatedMarketItem) {
  balanceBuyTarget.value = item
  void authStore.refreshUser().catch(() => {})
}

async function confirmBalanceBuyout() {
  const item = balanceBuyTarget.value
  if (!item) return
  purchasingBalance.value = true
  try {
    const result = await dedicatedAccountsAPI.buyWithBalance(item.id)
    balanceBuyTarget.value = null
    copiedNewKey.value = false
    purchaseResult.value = result
    appStore.showSuccess(t('dedicatedAccounts.market.purchaseSuccess'))
    await Promise.all([
      authStore.refreshUser().catch(() => {}),
      loadMyAccounts(),
      loadMarketplace()
    ])
  } catch (err) {
    appStore.showError(extractApiErrorMessage(err, t('common.error')))
  } finally {
    purchasingBalance.value = false
  }
}

// ==================== Full Credentials Buyout Actions ====================

function openOwnedBuyoutDialog(acc: UserDedicatedAccountView) {
  const priceCny = Number(acc.buyout_price_cny ?? 0)
  if (!acc.buyout_available || priceCny <= 0) return
  const priceBalance =
    typeof acc.buyout_price_balance === 'number' && acc.buyout_price_balance > 0
      ? acc.buyout_price_balance
      : priceCny
  buyoutAckChecked.value = false
  buyoutTarget.value = {
    accountId: acc.id,
    displayName: acc.display_name,
    platform: String(acc.platform),
    priceCny,
    priceBalance,
    source: 'owned'
  }
  void authStore.refreshUser().catch(() => {})
}

function openMarketBuyoutDialog(item: DedicatedMarketItem) {
  const priceCny = Number(item.buyout_price_cny ?? 0)
  if (!item.buyout_available || priceCny <= 0) return
  const priceBalance =
    typeof item.buyout_price_balance === 'number' && item.buyout_price_balance > 0
      ? item.buyout_price_balance
      : priceCny
  buyoutAckChecked.value = false
  buyoutTarget.value = {
    accountId: item.id,
    displayName: item.display_name,
    platform: String(item.platform),
    priceCny,
    priceBalance,
    source: 'shelf'
  }
  void authStore.refreshUser().catch(() => {})
}

function closeBuyoutDialog() {
  if (processingBuyout.value) return
  buyoutTarget.value = null
  buyoutAckChecked.value = false
}

async function confirmFullBuyoutWithBalance() {
  const target = buyoutTarget.value
  if (!target || !buyoutAckChecked.value) return
  processingBuyout.value = true
  try {
    const res = await dedicatedAccountsAPI.buyoutWithBalance(target.accountId)
    buyoutTarget.value = null
    buyoutAckChecked.value = false
    copiedBuyoutInfo.value = false
    copiedNewKey.value = false
    buyoutInfoModal.value = {
      accountId: res.account.id,
      displayName: res.account.display_name,
      buyoutInfo: res.buyout_info,
      apiKey: res.api_key || undefined
    }
    appStore.showSuccess(t('dedicatedAccounts.buyout.successToast'))
    await Promise.all([
      authStore.refreshUser().catch(() => {}),
      loadMyAccounts(),
      loadMarketplace()
    ])
  } catch (err) {
    appStore.showError(extractApiErrorMessage(err, t('common.error')))
  } finally {
    processingBuyout.value = false
  }
}

async function switchBuyoutToOnlinePay() {
  const target = buyoutTarget.value
  if (!target || !buyoutAckChecked.value) return
  buyoutTarget.value = null
  buyoutAckChecked.value = false
  await openOnlineCheckoutTarget({
    accountId: target.accountId,
    displayName: target.displayName,
    priceCny: target.priceCny,
    orderType: 'dedicated_buyout'
  })
}

async function handleViewBoughtOutInfo(acc: UserDedicatedAccountView) {
  loadingBuyoutInfoId.value = acc.id
  try {
    const info = await dedicatedAccountsAPI.getBuyoutInfo(acc.id)
    copiedBuyoutInfo.value = false
    copiedNewKey.value = false
    buyoutInfoModal.value = {
      accountId: info.account_id,
      displayName: info.display_name,
      buyoutInfo: info.buyout_info
    }
  } catch (err) {
    appStore.showError(extractApiErrorMessage(err, t('common.error')))
  } finally {
    loadingBuyoutInfoId.value = null
  }
}

async function copyBuyoutPlaintext(text: string) {
  const ok = await copyToClipboard(text)
  if (ok) {
    copiedBuyoutInfo.value = true
  }
}

function closeBuyoutInfoModal() {
  buyoutInfoModal.value = null
  if (activeTab.value !== 'my') {
    switchTab('my')
  }
}

async function copyFullApiKey(keyText: string) {
  const ok = await copyToClipboard(keyText)
  if (ok) {
    copiedNewKey.value = true
  }
}

function closePurchaseResultModal() {
  purchaseResult.value = null
  switchTab('my')
}

async function openOnlinePayDialog(item: DedicatedMarketItem) {
  await openOnlineCheckoutTarget({
    accountId: item.id,
    displayName: item.display_name,
    priceCny: item.current_price_cny,
    orderType: 'dedicated_account'
  })
}

async function openOnlineCheckoutTarget(target: OnlineOrderTargetState) {
  onlinePayTarget.value = target
  selectedPaymentMethod.value = ''
  loadingCheckout.value = true
  try {
    const res = await paymentAPI.getCheckoutInfo()
    checkoutInfo.value = res.data
    const firstAvail = onlinePaymentMethods.value.find((m) => m.available)
    if (firstAvail) {
      selectedPaymentMethod.value = firstAvail.type
    }
  } catch (err) {
    appStore.showError(extractApiErrorMessage(err, t('common.error')))
  } finally {
    loadingCheckout.value = false
  }
}

async function submitOnlineOrder() {
  if (!onlinePayTarget.value || !selectedPaymentMethod.value) return
  creatingOnlineOrder.value = true
  try {
    const target = onlinePayTarget.value
    const method = normalizeVisibleMethod(selectedPaymentMethod.value) || selectedPaymentMethod.value
    const payload = buildCreateOrderPayload({
      amount: target.priceCny,
      paymentType: method,
      orderType: target.orderType,
      dedicatedAccountId: target.accountId,
      origin: typeof window !== 'undefined' ? window.location.origin : '',
      isMobile: false,
      isWechatBrowser: false,
      forceQRCode: Boolean(checkoutInfo.value?.alipay_force_qrcode && method === 'alipay')
    })
    const res = await paymentAPI.createOrder(payload)
    const created = res.data
    onlinePayTarget.value = null

    if (created.qr_code) {
      qrDialog.orderId = created.order_id
      qrDialog.qrCode = created.qr_code
      qrDialog.expiresAt = created.expires_at || ''
      qrDialog.paymentType = method
      qrDialog.payUrl = ''
      qrDialog.orderType = target.orderType
      qrDialog.accountId = target.accountId
      qrDialog.show = true
      return
    }
    if (created.pay_url) {
      const win = window.open(created.pay_url, 'dedicatedPaymentPopup', getPaymentPopupFeatures())
      if (!win || win.closed) {
        window.location.href = created.pay_url
        return
      }
      qrDialog.orderId = created.order_id
      qrDialog.qrCode = ''
      qrDialog.expiresAt = created.expires_at || ''
      qrDialog.paymentType = method
      qrDialog.payUrl = created.pay_url
      qrDialog.orderType = target.orderType
      qrDialog.accountId = target.accountId
      qrDialog.show = true
    }
  } catch (err) {
    appStore.showError(extractApiErrorMessage(err, t('common.error')))
  } finally {
    creatingOnlineOrder.value = false
  }
}

async function handleOnlinePaySuccess() {
  const wasBuyout = qrDialog.orderType === 'dedicated_buyout'
  const targetAccountId = qrDialog.accountId
  qrDialog.show = false
  appStore.showSuccess(
    wasBuyout
      ? t('dedicatedAccounts.buyout.successToast')
      : t('dedicatedAccounts.market.purchaseSuccess')
  )
  await Promise.all([loadMyAccounts(), loadMarketplace()])
  switchTab('my')
  if (wasBuyout && targetAccountId > 0) {
    try {
      const info = await dedicatedAccountsAPI.getBuyoutInfo(targetAccountId)
      copiedBuyoutInfo.value = false
      buyoutInfoModal.value = {
        accountId: info.account_id,
        displayName: info.display_name,
        buyoutInfo: info.buyout_info
      }
    } catch {
      // Owner can still click "View Account Info" on the card at any time
    }
  }
}

// Health & Quota Presentation Helpers
function isExpired(expiresAt?: string | null): boolean {
  if (!expiresAt) return false
  return new Date(expiresAt).getTime() <= nowTick.value
}

function cooldownTargetTime(acc: UserDedicatedAccountView): string | null {
  if (acc.rate_limit_reset_at && new Date(acc.rate_limit_reset_at).getTime() > nowTick.value) {
    return acc.rate_limit_reset_at
  }
  return null
}

function healthBadgeLabel(acc: UserDedicatedAccountView): string {
  if (isExpired(acc.expires_at) || acc.health_status === 'expired') {
    return t('dedicatedAccounts.my.health.expired')
  }
  if (acc.health_status === 'error' || acc.health_status === 'paused' || acc.health_status === 'unknown') {
    return t('dedicatedAccounts.my.health.unavailable')
  }
  if (acc.health_status === 'rate_limited' || cooldownTargetTime(acc)) {
    return t('dedicatedAccounts.my.health.coolingDown')
  }
  if (acc.health_status === 'overloaded' || acc.health_status === 'cooldown') {
    return t('dedicatedAccounts.my.health.overload')
  }
  return t('dedicatedAccounts.my.health.active')
}

function healthBadgeClass(acc: UserDedicatedAccountView): string {
  if (isExpired(acc.expires_at) || acc.health_status === 'expired') {
    return 'bg-red-100 text-red-700 dark:bg-red-900/35 dark:text-red-300'
  }
  if (acc.health_status === 'error' || acc.health_status === 'paused' || acc.health_status === 'unknown') {
    return 'bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-gray-300'
  }
  if (
    acc.health_status === 'rate_limited' ||
    acc.health_status === 'overloaded' ||
    acc.health_status === 'cooldown' ||
    cooldownTargetTime(acc)
  ) {
    return 'bg-amber-100 text-amber-800 dark:bg-amber-900/35 dark:text-amber-300'
  }
  return 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/35 dark:text-emerald-300'
}

function clampPercent(val?: number): number {
  const n = Number(val ?? 0)
  if (!Number.isFinite(n)) return 0
  return Math.max(0, Math.min(100, n))
}

function formatPercent(val?: number): string {
  const c = clampPercent(val)
  return Number.isInteger(c) ? String(c) : c.toFixed(1)
}

function quotaBarColor(remainingPercent?: number): string {
  const p = clampPercent(remainingPercent)
  if (p <= 15) return 'bg-rose-500'
  if (p <= 40) return 'bg-amber-500'
  return 'bg-emerald-500'
}

function quotaTextColor(remainingPercent?: number): string {
  const p = clampPercent(remainingPercent)
  if (p <= 15) return 'text-rose-600 dark:text-rose-400'
  if (p <= 40) return 'text-amber-600 dark:text-amber-400'
  return 'text-emerald-600 dark:text-emerald-400'
}

function formatCountdownOnly(isoDate: string): string {
  const diffMs = new Date(isoDate).getTime() - nowTick.value
  if (diffMs <= 0) return t('dedicatedAccounts.my.quota.resetSoon')
  const totalMinutes = Math.ceil(diffMs / 60000)
  const days = Math.floor(totalMinutes / 1440)
  const hours = Math.floor((totalMinutes % 1440) / 60)
  const mins = totalMinutes % 60
  if (days > 0) return `${days}d ${hours}h`
  if (hours > 0) return `${hours}h ${mins}m`
  return `${mins}m`
}

function formatResetCountdown(isoDate: string): string {
  const diffMs = new Date(isoDate).getTime() - nowTick.value
  if (diffMs <= 0) return t('dedicatedAccounts.my.quota.resetSoon')
  return t('dedicatedAccounts.my.quota.resetCountdown', { time: formatCountdownOnly(isoDate) })
}

function remainingDays(isoDate: string): number {
  const diffMs = new Date(isoDate).getTime() - nowTick.value
  if (diffMs < 0) return -1
  return Math.ceil(diffMs / 86400000)
}

function formatShortDate(isoDate: string): string {
  return new Date(isoDate).toLocaleDateString()
}

function formatDateTime(isoDate: string): string {
  return new Date(isoDate).toLocaleString()
}

function maskKeyPreview(key: string): string {
  if (!key) return ''
  if (key.length <= 14) return key
  return `${key.slice(0, 9)}...${key.slice(-4)}`
}

function formatRelativeOrTime(isoDate: string): string {
  const diffSec = Math.floor((nowTick.value - new Date(isoDate).getTime()) / 1000)
  if (diffSec < 60) return t('common.time.justNow')
  if (diffSec < 3600) return t('common.time.minutesAgo', { n: Math.floor(diffSec / 60) })
  if (diffSec < 86400) return t('common.time.hoursAgo', { n: Math.floor(diffSec / 3600) })
  return formatShortDate(isoDate)
}

onMounted(async () => {
  tickTimer = setInterval(() => {
    nowTick.value = Date.now()
  }, 30000)
  await Promise.all([loadMyAccounts(), loadMarketplace()])
  if (!route.query.tab && myAccounts.value.length === 0 && marketItems.value.length > 0) {
    activeTab.value = 'market'
  }
  const redirectBuyoutId = Number(route.query.buyoutAccountId || 0)
  if (Number.isInteger(redirectBuyoutId) && redirectBuyoutId > 0) {
    activeTab.value = 'my'
    const matched = myAccounts.value.find((a) => a.id === redirectBuyoutId && a.bought_out)
    if (matched) {
      await handleViewBoughtOutInfo(matched)
    }
  }
})

onUnmounted(() => {
  if (tickTimer) clearInterval(tickTimer)
})
</script>
