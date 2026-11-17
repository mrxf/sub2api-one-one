<template>
  <!-- Custom Home Content: Full Page Mode -->
  <div v-if="hasHomeContent" class="min-h-screen">
    <!-- iframe mode -->
    <iframe
      v-if="isHomeContentUrl"
      :src="homeContent.trim()"
      class="h-screen w-full border-0"
      allowfullscreen
    ></iframe>
    <!-- HTML mode - SECURITY: homeContent is admin-only setting, XSS risk is acceptable -->
    <div v-else v-html="homeContent"></div>
  </div>

  <!-- Compact Home Page -->
  <div
    v-else-if="compactHomeEnabled"
    data-testid="compact-home"
    class="flex min-h-screen flex-col bg-gray-50 text-gray-900 dark:bg-dark-950 dark:text-white"
  >
    <header class="border-b border-gray-200 px-4 py-4 sm:px-6 dark:border-dark-800">
      <nav class="mx-auto flex max-w-5xl flex-wrap items-center justify-between gap-3 sm:gap-4">
        <div class="flex min-w-0 flex-1 items-center gap-3">
          <img
            :src="siteLogo || '/logo.svg'"
            alt="Logo"
            class="h-9 w-9 shrink-0 rounded-lg object-contain"
          />
          <span class="min-w-0 truncate text-base font-semibold">{{ siteName }}</span>
        </div>
        <div class="flex max-w-full shrink-0 flex-wrap items-center justify-end gap-2">
          <LocaleSwitcher />
          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg text-gray-500 hover:bg-gray-100 dark:text-dark-400 dark:hover:bg-dark-800"
            :title="t('home.viewDocs')"
          >
            <Icon name="book" size="md" />
          </a>
          <router-link
            v-if="showModelPlazaEntry"
            to="/model-plaza"
            class="flex h-10 shrink-0 items-center gap-1.5 rounded-lg px-2.5 text-sm font-medium text-gray-500 hover:bg-gray-100 hover:text-gray-700 dark:text-dark-400 dark:hover:bg-dark-800 dark:hover:text-white"
            :title="t('nav.modelPlaza')"
          >
            <Icon name="grid" size="md" />
            <span class="hidden sm:inline">{{ t('nav.modelPlaza') }}</span>
          </router-link>
          <button
            id="compact-theme-toggle"
            type="button"
            class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg text-gray-500 hover:bg-gray-100 dark:text-dark-400 dark:hover:bg-dark-800"
            :title="isDark ? t('home.switchToLight') : t('home.switchToDark')"
            @click="toggleTheme"
          >
            <Icon v-if="isDark" name="sun" size="md" />
            <Icon v-else name="moon" size="md" />
          </button>
          <router-link
            :to="isAuthenticated ? dashboardPath : '/login'"
            class="inline-flex min-h-10 shrink-0 items-center justify-center rounded-lg bg-gray-900 px-4 py-2 text-sm font-medium text-white hover:bg-gray-800 dark:bg-white dark:text-gray-900 dark:hover:bg-gray-200"
          >
            {{ isAuthenticated ? t('home.dashboard') : t('home.login') }}
          </router-link>
        </div>
      </nav>
    </header>

    <main class="flex min-w-0 flex-1 items-center justify-center px-4 py-16 sm:px-6">
      <div class="min-w-0 max-w-2xl text-center">
        <img
          :src="siteLogo || '/logo.svg'"
          alt="Logo"
          class="mx-auto mb-6 h-20 w-20 rounded-2xl object-contain"
        />
        <h1 class="[overflow-wrap:anywhere] text-3xl font-bold md:text-4xl">{{ siteName }}</h1>
        <p class="mt-4 whitespace-pre-wrap [overflow-wrap:anywhere] text-base text-gray-600 dark:text-dark-300">{{ siteSubtitle }}</p>
        <router-link
          :to="isAuthenticated ? dashboardPath : '/login'"
          class="mt-8 inline-flex min-h-10 items-center justify-center rounded-lg bg-primary-600 px-5 py-2.5 text-sm font-medium text-white hover:bg-primary-700"
        >
          {{ isAuthenticated ? t('home.goToDashboard') : t('home.login') }}
        </router-link>
      </div>
    </main>

    <footer class="min-w-0 border-t border-gray-200 px-4 py-5 text-center text-sm text-gray-500 [overflow-wrap:anywhere] sm:px-6 dark:border-dark-800 dark:text-dark-400">
      &copy; {{ currentYear }} {{ siteName }}
    </footer>
  </div>

  <!-- Default Home Page: Dedicated Accounts Enterprise Introduction -->
  <div
    v-else
    data-testid="dedicated-home"
    class="landing-root flex min-h-screen flex-col bg-gray-50 text-gray-900 antialiased dark:bg-dark-950 dark:text-dark-100"
  >
    <!-- Sticky Enterprise Header -->
    <header
      class="sticky top-0 z-30 border-b border-gray-200/80 bg-white/95 backdrop-blur-md dark:border-dark-800 dark:bg-dark-950/95"
    >
      <nav
        aria-label="Primary"
        class="mx-auto flex max-w-7xl items-center justify-between gap-4 px-4 py-3.5 sm:px-6 lg:px-8"
      >
        <!-- Brand Identity -->
        <div class="flex min-w-0 items-center gap-8">
          <a href="#" id="brand-home-link" class="flex min-w-0 items-center gap-3">
            <div
              class="flex h-9 w-9 shrink-0 items-center justify-center overflow-hidden rounded-lg border border-gray-200/80 bg-white p-1 shadow-sm dark:border-dark-700 dark:bg-dark-900"
            >
              <img :src="siteLogo || '/logo.svg'" alt="Logo" class="h-full w-full object-contain" />
            </div>
            <span class="truncate text-base font-semibold tracking-tight text-gray-900 dark:text-white">
              {{ siteName }}
            </span>
          </a>

          <!-- Section Anchor Links (Desktop) -->
          <div class="hidden items-center gap-6 lg:flex">
            <a
              href="#capabilities"
              id="nav-link-capabilities"
              class="text-sm font-medium text-gray-600 transition-colors hover:text-gray-900 dark:text-dark-300 dark:hover:text-white"
            >
              {{ t('home.dedicatedLanding.nav.capabilities') }}
            </a>
            <a
              href="#routing"
              id="nav-link-routing"
              class="text-sm font-medium text-gray-600 transition-colors hover:text-gray-900 dark:text-dark-300 dark:hover:text-white"
            >
              {{ t('home.dedicatedLanding.nav.routing') }}
            </a>
            <a
              href="#marketplace"
              id="nav-link-marketplace"
              class="text-sm font-medium text-gray-600 transition-colors hover:text-gray-900 dark:text-dark-300 dark:hover:text-white"
            >
              {{ t('home.dedicatedLanding.nav.marketplace') }}
            </a>
            <a
              href="#comparison"
              id="nav-link-comparison"
              class="text-sm font-medium text-gray-600 transition-colors hover:text-gray-900 dark:text-dark-300 dark:hover:text-white"
            >
              {{ t('home.dedicatedLanding.nav.comparison') }}
            </a>
            <a
              href="#ecosystem"
              id="nav-link-ecosystem"
              class="text-sm font-medium text-gray-600 transition-colors hover:text-gray-900 dark:text-dark-300 dark:hover:text-white"
            >
              {{ t('home.dedicatedLanding.nav.ecosystem') }}
            </a>
          </div>
        </div>

        <!-- Right Utility & Auth Actions -->
        <div class="flex shrink-0 items-center gap-2 sm:gap-2.5">
          <router-link
            to="/key-usage"
            id="nav-key-usage-link"
            class="hidden items-center gap-1.5 rounded-lg px-2.5 py-2 text-sm font-medium text-gray-600 transition-colors hover:bg-gray-100 hover:text-gray-900 md:inline-flex dark:text-dark-300 dark:hover:bg-dark-800 dark:hover:text-white"
          >
            <Icon name="chart" size="sm" />
            <span>{{ t('home.dedicatedLanding.nav.keyUsage') }}</span>
          </router-link>

          <router-link
            v-if="showModelPlazaEntry"
            to="/model-plaza"
            id="nav-model-plaza-link"
            class="inline-flex items-center gap-1.5 rounded-lg px-2.5 py-2 text-sm font-medium text-gray-600 transition-colors hover:bg-gray-100 hover:text-gray-900 dark:text-dark-300 dark:hover:bg-dark-800 dark:hover:text-white"
            :title="t('nav.modelPlaza')"
          >
            <Icon name="grid" size="sm" />
            <span class="hidden sm:inline">{{ t('nav.modelPlaza') }}</span>
          </router-link>

          <a
            v-if="docUrl"
            :href="docUrl"
            id="nav-docs-link"
            target="_blank"
            rel="noopener noreferrer"
            class="inline-flex h-9 w-9 items-center justify-center rounded-lg text-gray-500 transition-colors hover:bg-gray-100 hover:text-gray-900 dark:text-dark-400 dark:hover:bg-dark-800 dark:hover:text-white"
            :title="t('home.viewDocs')"
          >
            <Icon name="book" size="md" />
          </a>

          <LocaleSwitcher />

          <button
            id="nav-theme-toggle"
            type="button"
            @click="toggleTheme"
            class="inline-flex h-9 w-9 items-center justify-center rounded-lg text-gray-500 transition-colors hover:bg-gray-100 hover:text-gray-900 dark:text-dark-400 dark:hover:bg-dark-800 dark:hover:text-white"
            :title="isDark ? t('home.switchToLight') : t('home.switchToDark')"
          >
            <Icon v-if="isDark" name="sun" size="md" />
            <Icon v-else name="moon" size="md" />
          </button>

          <div class="hidden h-4 w-px bg-gray-200 sm:block dark:bg-dark-700"></div>

          <!-- Authenticated Controls -->
          <template v-if="isAuthenticated">
            <router-link
              to="/dedicated-accounts"
              id="nav-dedicated-center-btn"
              class="hidden items-center gap-1.5 rounded-lg border border-gray-200 bg-white px-3 py-1.5 text-sm font-medium text-gray-800 shadow-sm transition-colors hover:bg-gray-50 sm:inline-flex dark:border-dark-700 dark:bg-dark-800 dark:text-dark-100 dark:hover:bg-dark-700"
            >
              <Icon name="key" size="sm" class="text-primary-600 dark:text-primary-400" />
              <span>{{ t('home.dedicatedLanding.nav.dedicatedCenter') }}</span>
            </router-link>
            <router-link
              :to="dashboardPath"
              id="nav-dashboard-btn"
              class="btn btn-primary btn-sm inline-flex items-center gap-2 px-3.5 py-2 text-sm"
            >
              <span
                v-if="userInitial"
                class="flex h-5 w-5 items-center justify-center rounded bg-white/20 text-xs font-semibold text-white"
              >
                {{ userInitial }}
              </span>
              <span>{{ t('home.dashboard') }}</span>
            </router-link>
          </template>

          <!-- Unauthenticated Controls -->
          <template v-else>
            <router-link
              to="/login"
              id="nav-login-btn"
              class="inline-flex items-center rounded-lg px-3 py-2 text-sm font-medium text-gray-700 transition-colors hover:bg-gray-100 hover:text-gray-900 dark:text-dark-200 dark:hover:bg-dark-800 dark:hover:text-white"
            >
              {{ t('home.login') }}
            </router-link>
            <router-link
              v-if="registrationEnabled"
              to="/register"
              id="nav-register-btn"
              class="btn btn-primary btn-sm px-3.5 py-2 text-sm"
            >
              {{ t('home.dedicatedLanding.nav.register') }}
            </router-link>
          </template>
        </div>
      </nav>
    </header>

    <!-- Main Editorial Content -->
    <main class="flex-1">
      <!-- ==================== BAND 1: HERO & INTERACTIVE DEDICATED CONSOLE ==================== -->
      <section class="border-b border-gray-200/80 bg-white dark:border-dark-800 dark:bg-dark-900/50">
        <div class="mx-auto max-w-7xl px-4 pb-16 pt-12 sm:px-6 sm:pb-20 sm:pt-16 lg:px-8 lg:pb-24 lg:pt-20">
          <div class="grid items-start gap-12 lg:grid-cols-12 lg:gap-12">
            <!-- Left Editorial Pitch (6 Columns) -->
            <div class="lg:col-span-6 lg:pt-2">
              <div class="mb-4 flex items-center gap-3">
                <span class="h-px w-8 bg-primary-600 dark:bg-primary-400"></span>
                <p
                  class="text-xs font-semibold uppercase tracking-[0.14em] text-primary-600 dark:text-primary-400"
                >
                  {{ t('home.dedicatedLanding.hero.overline') }}
                </p>
              </div>

              <h1
                class="landing-display-title text-3xl font-semibold tracking-tight text-gray-900 sm:text-4xl lg:text-[2.75rem] dark:text-white"
              >
                {{ t('home.dedicatedLanding.hero.title') }}
              </h1>

              <p
                v-if="hasCustomSubtitle"
                class="mt-4 text-base font-medium text-gray-800 dark:text-dark-200"
              >
                {{ siteSubtitle }}
              </p>

              <p
                class="mt-5 max-w-2xl text-base leading-relaxed text-gray-600 sm:text-[1.0625rem] sm:leading-7 dark:text-dark-300"
              >
                {{ t('home.dedicatedLanding.hero.description') }}
              </p>

              <!-- Primary & Secondary Actions -->
              <div class="mt-8 flex flex-wrap items-center gap-3.5">
                <router-link
                  :to="primaryDedicatedPath"
                  id="hero-primary-cta"
                  class="btn btn-primary px-6 py-3 text-sm font-semibold sm:text-base"
                >
                  <span>
                    {{
                      isAuthenticated
                        ? t('home.dedicatedLanding.hero.primaryCtaAuthed')
                        : t('home.dedicatedLanding.hero.primaryCtaGuest')
                    }}
                  </span>
                  <Icon name="arrowRight" size="sm" :stroke-width="2" />
                </router-link>

                <router-link
                  :to="isAuthenticated ? dashboardPath : '/key-usage'"
                  id="hero-secondary-cta"
                  class="btn btn-secondary px-5 py-3 text-sm font-medium sm:text-base"
                >
                  <Icon :name="isAuthenticated ? 'server' : 'chart'" size="sm" />
                  <span>
                    {{
                      isAuthenticated
                        ? t('home.dedicatedLanding.hero.secondaryCtaAuthed')
                        : t('home.dedicatedLanding.hero.secondaryCtaGuest')
                    }}
                  </span>
                </router-link>
              </div>

              <!-- Architectural Metrics Strip -->
              <dl
                class="mt-10 grid grid-cols-2 gap-x-4 gap-y-6 border-t border-gray-200 pt-8 sm:grid-cols-4 dark:border-dark-800"
              >
                <div>
                  <dt class="font-mono text-2xl font-semibold tracking-tight text-gray-900 tabular-nums dark:text-white">
                    {{ t('home.dedicatedLanding.hero.metrics.isolationValue') }}
                  </dt>
                  <dd class="mt-1 text-xs leading-snug text-gray-500 sm:whitespace-nowrap dark:text-dark-400">
                    {{ t('home.dedicatedLanding.hero.metrics.isolationLabel') }}
                  </dd>
                </div>
                <div>
                  <dt class="font-mono text-2xl font-semibold tracking-tight text-gray-900 tabular-nums dark:text-white">
                    {{ t('home.dedicatedLanding.hero.metrics.balanceValue') }}
                  </dt>
                  <dd class="mt-1 text-xs leading-snug text-gray-500 sm:whitespace-nowrap dark:text-dark-400">
                    {{ t('home.dedicatedLanding.hero.metrics.balanceLabel') }}
                  </dd>
                </div>
                <div>
                  <dt class="font-mono text-2xl font-semibold tracking-tight text-gray-900 tabular-nums dark:text-white">
                    {{ t('home.dedicatedLanding.hero.metrics.overflowValue') }}
                  </dt>
                  <dd class="mt-1 text-xs leading-snug text-gray-500 sm:whitespace-nowrap dark:text-dark-400">
                    {{ t('home.dedicatedLanding.hero.metrics.overflowLabel') }}
                  </dd>
                </div>
                <div>
                  <dt class="font-mono text-2xl font-semibold tracking-tight text-gray-900 tabular-nums dark:text-white">
                    {{ t('home.dedicatedLanding.hero.metrics.windowValue') }}
                  </dt>
                  <dd class="mt-1 text-xs leading-snug text-gray-500 sm:whitespace-nowrap dark:text-dark-400">
                    {{ t('home.dedicatedLanding.hero.metrics.windowLabel') }}
                  </dd>
                </div>
              </dl>
            </div>

            <!-- Right Interactive Product Showcase (6 Columns) -->
            <div class="lg:col-span-6">
              <div
                class="showcase-frame overflow-hidden rounded-xl border border-gray-200 bg-white shadow-card dark:border-dark-700 dark:bg-dark-900"
              >
                <!-- Frame Top Bar: Mode Switcher -->
                <div
                  class="flex flex-wrap items-center justify-between gap-3 border-b border-gray-200 bg-gray-50/90 px-4 py-3 dark:border-dark-700 dark:bg-dark-950/70"
                >
                  <div class="inline-flex rounded-lg border border-gray-200 bg-white p-0.5 dark:border-dark-700 dark:bg-dark-900">
                    <button
                      id="showcase-tab-console"
                      type="button"
                      @click="activeShowcaseTab = 'console'"
                      :class="[
                        'inline-flex items-center gap-1.5 rounded-md px-3 py-1.5 text-xs font-semibold transition-colors',
                        activeShowcaseTab === 'console'
                          ? 'bg-gray-900 text-white dark:bg-primary-600 dark:text-white'
                          : 'text-gray-600 hover:text-gray-900 dark:text-dark-300 dark:hover:text-white'
                      ]"
                    >
                      <Icon name="shield" size="xs" />
                      <span>{{ t('home.dedicatedLanding.showcase.tabConsole') }}</span>
                    </button>
                    <button
                      id="showcase-tab-terminal"
                      type="button"
                      @click="activeShowcaseTab = 'terminal'"
                      :class="[
                        'inline-flex items-center gap-1.5 rounded-md px-3 py-1.5 text-xs font-semibold transition-colors',
                        activeShowcaseTab === 'terminal'
                          ? 'bg-gray-900 text-white dark:bg-primary-600 dark:text-white'
                          : 'text-gray-600 hover:text-gray-900 dark:text-dark-300 dark:hover:text-white'
                      ]"
                    >
                      <Icon name="terminal" size="xs" />
                      <span>{{ t('home.dedicatedLanding.showcase.tabTerminal') }}</span>
                    </button>
                  </div>

                  <span
                    class="inline-flex items-center gap-1.5 text-xs font-medium text-emerald-700 dark:text-emerald-400"
                  >
                    <span class="h-2 w-2 rounded-full bg-emerald-500"></span>
                    {{ t('home.dedicatedLanding.showcase.liveStatus') }}
                  </span>
                </div>

                <!-- TAB 1: Interactive Dedicated Account Console Preview -->
                <div v-show="activeShowcaseTab === 'console'" class="p-5 sm:p-6">
                  <!-- Upstream Account Selector Strip -->
                  <div class="mb-5 grid grid-cols-3 gap-2">
                    <button
                      v-for="item in showcaseAccounts"
                      :key="item.id"
                      :id="`showcase-account-${item.id}`"
                      type="button"
                      @click="selectedShowcaseAccountId = item.id"
                      :class="[
                        'flex items-center gap-2 rounded-lg border px-3 py-2 text-left text-xs font-medium transition-all',
                        selectedShowcaseAccountId === item.id
                          ? 'border-primary-600 bg-primary-50/70 text-gray-900 dark:border-primary-400 dark:bg-primary-950/40 dark:text-white'
                          : 'border-gray-200 bg-gray-50/60 text-gray-600 hover:border-gray-300 hover:bg-white dark:border-dark-700 dark:bg-dark-800/50 dark:text-dark-300 dark:hover:border-dark-600'
                      ]"
                    >
                      <PlatformIcon :platform="item.platform" size="sm" />
                      <span class="truncate font-semibold">{{ item.shortLabel }}</span>
                    </button>
                  </div>

                  <!-- Active Dedicated Slot Header -->
                  <div class="flex flex-wrap items-start justify-between gap-3 border-b border-gray-100 pb-4 dark:border-dark-800">
                    <div class="flex min-w-0 items-center gap-3">
                      <div
                        class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg border border-gray-200 bg-gray-50 text-gray-800 dark:border-dark-700 dark:bg-dark-800 dark:text-dark-100"
                      >
                        <PlatformIcon :platform="currentShowcaseAccount.platform" size="lg" />
                      </div>
                      <div class="min-w-0">
                        <div class="flex flex-wrap items-center gap-2">
                          <span class="truncate text-sm font-semibold text-gray-900 dark:text-white">
                            {{ t(currentShowcaseAccount.nameKey) }}
                          </span>
                          <span
                            class="rounded bg-primary-50 px-2 py-0.5 text-[11px] font-semibold text-primary-700 dark:bg-primary-900/40 dark:text-primary-300"
                          >
                            {{ t('home.dedicatedLanding.showcase.pinnedTag') }}
                          </span>
                        </div>
                        <div class="mt-1 flex flex-wrap items-center gap-2 font-mono text-xs text-gray-500 dark:text-dark-400">
                          <span>{{ t(currentShowcaseAccount.groupKey) }}</span>
                          <span>·</span>
                          <span>{{ t(currentShowcaseAccount.concurrencyKey) }}</span>
                        </div>
                      </div>
                    </div>

                    <span
                      class="rounded bg-emerald-50 px-2.5 py-1 text-xs font-semibold text-emerald-700 dark:bg-emerald-950/60 dark:text-emerald-300"
                    >
                      {{ t('home.dedicatedLanding.showcase.zeroBalanceTag') }}
                    </span>
                  </div>

                  <!-- Live 5H / 7D Window Bars -->
                  <div class="mt-4 space-y-4">
                    <!-- 5H Rolling Window -->
                    <div>
                      <div class="mb-1.5 flex items-center justify-between text-xs">
                        <span class="font-medium text-gray-700 dark:text-dark-200">
                          {{ t('home.dedicatedLanding.showcase.fiveHourLabel') }}
                        </span>
                        <div class="flex items-center gap-2.5 tabular-nums">
                          <span class="text-gray-400 dark:text-dark-400">
                            {{ t('home.dedicatedLanding.showcase.resetsIn2h') }}
                          </span>
                          <span class="font-mono font-semibold text-emerald-600 dark:text-emerald-400">
                            {{ currentShowcaseAccount.fiveHourRemaining }}%
                          </span>
                        </div>
                      </div>
                      <div class="h-2 w-full overflow-hidden rounded-full bg-gray-100 dark:bg-dark-800">
                        <div
                          class="h-full rounded-full bg-emerald-500 transition-all duration-300"
                          :style="{ width: `${currentShowcaseAccount.fiveHourRemaining}%` }"
                        ></div>
                      </div>
                    </div>

                    <!-- 7D Weekly Window -->
                    <div>
                      <div class="mb-1.5 flex items-center justify-between text-xs">
                        <span class="font-medium text-gray-700 dark:text-dark-200">
                          {{ t('home.dedicatedLanding.showcase.sevenDayLabel') }}
                        </span>
                        <div class="flex items-center gap-2.5 tabular-nums">
                          <span class="text-gray-400 dark:text-dark-400">
                            {{ t('home.dedicatedLanding.showcase.resetsIn4d') }}
                          </span>
                          <span class="font-mono font-semibold text-primary-600 dark:text-primary-400">
                            {{ currentShowcaseAccount.sevenDayRemaining }}%
                          </span>
                        </div>
                      </div>
                      <div class="h-2 w-full overflow-hidden rounded-full bg-gray-100 dark:bg-dark-800">
                        <div
                          class="h-full rounded-full bg-primary-600 transition-all duration-300 dark:bg-primary-400"
                          :style="{ width: `${currentShowcaseAccount.sevenDayRemaining}%` }"
                        ></div>
                      </div>
                    </div>
                  </div>

                  <!-- Pinned API Key + Reset Credit & Buyout Strip -->
                  <div
                    class="mt-5 flex flex-wrap items-center justify-between gap-2 border-t border-gray-100 pt-4 text-xs dark:border-dark-800"
                  >
                    <div class="flex items-center gap-2">
                      <span class="text-gray-500 dark:text-dark-400">
                        {{ t('home.dedicatedLanding.showcase.boundKeyLabel') }}:
                      </span>
                      <code
                        class="rounded bg-gray-100 px-2 py-0.5 font-mono text-gray-800 dark:bg-dark-800 dark:text-dark-200"
                      >
                        {{ currentShowcaseAccount.maskedKey }}
                      </code>
                    </div>

                    <span
                      v-if="currentShowcaseAccount.hasResetCredits"
                      class="inline-flex items-center gap-1 rounded bg-primary-50 px-2 py-0.5 font-medium text-primary-700 dark:bg-primary-900/40 dark:text-primary-300"
                    >
                      <Icon name="refresh" size="xs" />
                      {{ t('home.dedicatedLanding.showcase.resetCreditTag') }}
                    </span>
                    <span
                      v-else
                      class="inline-flex items-center gap-1 rounded bg-amber-50 px-2 py-0.5 font-medium text-amber-800 dark:bg-amber-950/50 dark:text-amber-300"
                    >
                      <Icon name="key" size="xs" />
                      {{ t('home.dedicatedLanding.showcase.buyoutReadyTag') }}
                    </span>
                  </div>

                  <!-- Redaction Privacy Footer -->
                  <div
                    class="mt-4 flex items-center gap-2 rounded-lg bg-gray-50 px-3 py-2 text-xs text-gray-600 dark:bg-dark-950/70 dark:text-dark-300"
                  >
                    <Icon name="lock" size="xs" class="shrink-0 text-primary-600 dark:text-primary-400" />
                    <span>{{ t('home.dedicatedLanding.showcase.redactedShield') }}</span>
                  </div>
                </div>

                <!-- TAB 2: 1:1 Gateway Verification Terminal (preserves .terminal-container contract) -->
                <div v-show="activeShowcaseTab === 'terminal'" class="p-5 sm:p-6">
                  <div class="terminal-container w-full">
                    <div class="terminal-window w-full">
                      <div class="terminal-header">
                        <div class="terminal-buttons">
                          <span class="btn-close"></span>
                          <span class="btn-minimize"></span>
                          <span class="btn-maximize"></span>
                        </div>
                        <span class="terminal-title">
                          {{ t('home.dedicatedLanding.showcase.terminalTitle') }}
                        </span>
                      </div>
                      <div class="terminal-body">
                        <div class="code-line">
                          <span class="code-prompt">$</span>
                          <span class="code-cmd">curl</span>
                          <span class="code-flag">-X POST</span>
                          <span class="code-url">{{ effectiveApiBaseUrl }}/v1/messages</span>
                        </div>
                        <div class="code-line">
                          <span class="code-comment">{{
                            t('home.dedicatedLanding.showcase.terminalCommentPin')
                          }}</span>
                        </div>
                        <div class="code-line">
                          <span class="code-comment">{{
                            t('home.dedicatedLanding.showcase.terminalCommentPass')
                          }}</span>
                        </div>
                        <div class="code-line">
                          <span class="code-success">200 OK</span>
                          <span class="code-response">x-dedicated-isolation: strict-1to1</span>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      <!-- ==================== BAND 2: FOUR CORE DEDICATED CAPABILITIES ==================== -->
      <section id="capabilities" class="scroll-mt-16 border-b border-gray-200/80 py-16 sm:py-24 dark:border-dark-800">
        <div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
          <div class="max-w-3xl">
            <p class="text-xs font-semibold uppercase tracking-[0.14em] text-primary-600 dark:text-primary-400">
              {{ t('home.dedicatedLanding.capabilities.overline') }}
            </p>
            <h2 class="mt-2 text-2xl font-semibold tracking-tight text-gray-900 sm:text-3xl dark:text-white">
              {{ t('home.dedicatedLanding.capabilities.title') }}
            </h2>
            <p class="mt-3 text-base leading-relaxed text-gray-600 dark:text-dark-300">
              {{ t('home.dedicatedLanding.capabilities.subtitle') }}
            </p>
          </div>

          <!-- 2x2 Architectural Capability Matrix (Hairline Separated) -->
          <div
            class="mt-12 grid overflow-hidden rounded-xl border border-gray-200 bg-white shadow-card md:grid-cols-2 dark:border-dark-700 dark:bg-dark-900"
          >
            <article
              v-for="(cap, idx) in capabilityItems"
              :key="cap.key"
              :class="[
                'flex flex-col justify-between p-6 sm:p-8',
                idx === 0 ? 'border-b border-gray-200 md:border-r dark:border-dark-700' : '',
                idx === 1 ? 'border-b border-gray-200 dark:border-dark-700' : '',
                idx === 2 ? 'border-b border-gray-200 md:border-b-0 md:border-r dark:border-dark-700' : ''
              ]"
            >
              <div>
                <div class="flex items-center justify-between gap-4">
                  <div
                    class="flex h-10 w-10 items-center justify-center rounded-lg border border-gray-200 bg-gray-50 text-primary-600 dark:border-dark-700 dark:bg-dark-800 dark:text-primary-400"
                  >
                    <Icon :name="cap.icon" size="md" />
                  </div>
                  <span class="font-mono text-xs font-semibold tracking-widest text-gray-400 dark:text-dark-500">
                    {{ t(`home.dedicatedLanding.capabilities.items.${cap.key}.index`) }}
                  </span>
                </div>

                <h3 class="mt-5 text-lg font-semibold text-gray-900 dark:text-white">
                  {{ t(`home.dedicatedLanding.capabilities.items.${cap.key}.title`) }}
                </h3>
                <p class="mt-2.5 text-sm leading-relaxed text-gray-600 dark:text-dark-300">
                  {{ t(`home.dedicatedLanding.capabilities.items.${cap.key}.desc`) }}
                </p>
              </div>

              <div class="mt-6 flex flex-wrap items-center gap-2 border-t border-gray-100 pt-4 dark:border-dark-800">
                <span
                  class="inline-flex items-center gap-1.5 rounded bg-gray-100 px-2.5 py-1 text-xs font-medium text-gray-700 dark:bg-dark-800 dark:text-dark-200"
                >
                  <Icon name="check" size="xs" class="text-primary-600 dark:text-primary-400" />
                  {{ t(`home.dedicatedLanding.capabilities.items.${cap.key}.spec1`) }}
                </span>
                <span
                  class="inline-flex items-center gap-1.5 rounded bg-gray-100 px-2.5 py-1 text-xs font-medium text-gray-700 dark:bg-dark-800 dark:text-dark-200"
                >
                  <Icon name="check" size="xs" class="text-primary-600 dark:text-primary-400" />
                  {{ t(`home.dedicatedLanding.capabilities.items.${cap.key}.spec2`) }}
                </span>
              </div>
            </article>
          </div>
        </div>
      </section>

      <!-- ==================== BAND 3: INTERACTIVE TRI-MODE ROUTING ARCHITECTURE ==================== -->
      <section
        id="routing"
        class="scroll-mt-16 border-b border-gray-200/80 bg-white py-16 sm:py-24 dark:border-dark-800 dark:bg-dark-900/40"
      >
        <div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
          <div class="max-w-3xl">
            <p class="text-xs font-semibold uppercase tracking-[0.14em] text-primary-600 dark:text-primary-400">
              {{ t('home.dedicatedLanding.routing.overline') }}
            </p>
            <h2 class="mt-2 text-2xl font-semibold tracking-tight text-gray-900 sm:text-3xl dark:text-white">
              {{ t('home.dedicatedLanding.routing.title') }}
            </h2>
            <p class="mt-3 text-base leading-relaxed text-gray-600 dark:text-dark-300">
              {{ t('home.dedicatedLanding.routing.subtitle') }}
            </p>
          </div>

          <div class="mt-12 grid items-start gap-8 lg:grid-cols-12">
            <!-- Left Selector Column (5 Cols) -->
            <div class="space-y-3 lg:col-span-5">
              <button
                v-for="mode in routingModes"
                :key="mode.id"
                :id="`routing-tab-${mode.id}`"
                type="button"
                @click="activeRoutingMode = mode.id"
                :class="[
                  'w-full rounded-xl border p-5 text-left transition-all',
                  activeRoutingMode === mode.id
                    ? 'border-primary-600 bg-primary-50/40 shadow-sm dark:border-primary-400 dark:bg-primary-950/25'
                    : 'border-gray-200 bg-white hover:border-gray-300 dark:border-dark-700 dark:bg-dark-900 dark:hover:border-dark-600'
                ]"
              >
                <div class="flex items-center justify-between gap-2">
                  <span class="text-base font-semibold text-gray-900 dark:text-white">
                    {{ t(`home.dedicatedLanding.routing.modes.${mode.id}.title`) }}
                  </span>
                  <span
                    :class="[
                      'shrink-0 rounded px-2 py-0.5 text-xs font-semibold',
                      activeRoutingMode === mode.id
                        ? 'bg-primary-600 text-white dark:bg-primary-500'
                        : 'bg-gray-100 text-gray-600 dark:bg-dark-800 dark:text-dark-300'
                    ]"
                  >
                    {{ t(`home.dedicatedLanding.routing.modes.${mode.id}.badge`) }}
                  </span>
                </div>
                <p class="mt-2 text-sm leading-relaxed text-gray-600 dark:text-dark-300">
                  {{ t(`home.dedicatedLanding.routing.modes.${mode.id}.summary`) }}
                </p>
              </button>
            </div>

            <!-- Right Live Routing Schematic & Contract Readout (7 Cols) -->
            <div
              class="overflow-hidden rounded-xl border border-gray-200 bg-gray-50/70 shadow-card lg:col-span-7 dark:border-dark-700 dark:bg-dark-900"
            >
              <!-- Visual Topology Flow -->
              <div class="border-b border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-900">
                <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
                  <!-- Node 1: Client -->
                  <div class="flex-1 rounded-lg border border-gray-200 bg-gray-50 p-3.5 dark:border-dark-700 dark:bg-dark-800/70">
                    <div class="flex items-center gap-2 text-xs font-semibold text-gray-900 dark:text-white">
                      <Icon name="terminal" size="sm" class="text-primary-600 dark:text-primary-400" />
                      <span>{{ t('home.dedicatedLanding.routing.clientLabel') }}</span>
                    </div>
                    <p class="mt-1 font-mono text-[11px] text-gray-500 dark:text-dark-400">
                      {{ t('home.dedicatedLanding.routing.clientSub') }}
                    </p>
                  </div>

                  <!-- Arrow -->
                  <div class="flex items-center justify-center text-gray-400 dark:text-dark-500">
                    <Icon name="arrowRight" size="md" class="rotate-90 sm:rotate-0" />
                  </div>

                  <!-- Node 2: Isolation Kernel -->
                  <div class="flex-1 rounded-lg border border-primary-200 bg-primary-50/60 p-3.5 dark:border-primary-800/60 dark:bg-primary-950/30">
                    <div class="flex items-center gap-2 text-xs font-semibold text-primary-900 dark:text-primary-200">
                      <Icon name="shield" size="sm" class="text-primary-600 dark:text-primary-400" />
                      <span>{{ t('home.dedicatedLanding.routing.gatewayLabel') }}</span>
                    </div>
                    <p class="mt-1 font-mono text-[11px] text-primary-700 dark:text-primary-300">
                      mode = {{ activeRoutingMode }}
                    </p>
                  </div>
                </div>

                <!-- Downstream Target Outcome -->
                <div class="mt-4 grid gap-3 sm:grid-cols-2">
                  <div
                    class="rounded-lg border border-emerald-200 bg-emerald-50/60 p-4 dark:border-emerald-800/60 dark:bg-emerald-950/25"
                  >
                    <div class="flex items-center gap-2 text-xs font-semibold text-emerald-900 dark:text-emerald-200">
                      <Icon name="checkCircle" size="sm" class="text-emerald-600 dark:text-emerald-400" />
                      <span>{{ t(`home.dedicatedLanding.routing.modes.${activeRoutingMode}.targetTitle`) }}</span>
                    </div>
                    <p class="mt-1.5 text-xs leading-relaxed text-emerald-800/90 dark:text-emerald-300/80">
                      {{ t(`home.dedicatedLanding.routing.modes.${activeRoutingMode}.targetSub`) }}
                    </p>
                  </div>

                  <div
                    v-if="activeRoutingMode !== 'shared'"
                    class="flex items-center gap-3 rounded-lg border border-gray-200 bg-gray-50 p-4 dark:border-dark-700 dark:bg-dark-800/60"
                  >
                    <Icon name="ban" size="md" class="shrink-0 text-gray-400 dark:text-dark-400" />
                    <span class="text-xs font-medium leading-relaxed text-gray-600 dark:text-dark-300">
                      {{ t('home.dedicatedLanding.routing.publicPoolBlocked') }}
                    </span>
                  </div>
                  <div
                    v-else
                    class="flex items-center gap-3 rounded-lg border border-gray-200 bg-gray-50 p-4 dark:border-dark-700 dark:bg-dark-800/60"
                  >
                    <Icon name="swap" size="md" class="shrink-0 text-primary-600 dark:text-primary-400" />
                    <span class="text-xs font-medium leading-relaxed text-gray-600 dark:text-dark-300">
                      {{ t('home.dedicatedLanding.routing.modes.shared.failover') }}
                    </span>
                  </div>
                </div>
              </div>

              <!-- Routing Contract Specification Grid -->
              <div class="p-6">
                <h3 class="text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-dark-400">
                  {{ t('home.dedicatedLanding.routing.specsTitle') }}
                </h3>
                <dl class="mt-4 grid gap-4 sm:grid-cols-2">
                  <div class="border-l-2 border-primary-600 pl-3.5 dark:border-primary-400">
                    <dt class="text-xs text-gray-500 dark:text-dark-400">
                      {{ t('home.dedicatedLanding.routing.specBilling') }}
                    </dt>
                    <dd class="mt-1 text-sm font-medium text-gray-900 dark:text-white">
                      {{ t(`home.dedicatedLanding.routing.modes.${activeRoutingMode}.billing`) }}
                    </dd>
                  </div>
                  <div class="border-l-2 border-primary-600 pl-3.5 dark:border-primary-400">
                    <dt class="text-xs text-gray-500 dark:text-dark-400">
                      {{ t('home.dedicatedLanding.routing.specFailover') }}
                    </dt>
                    <dd class="mt-1 text-sm font-medium text-gray-900 dark:text-white">
                      {{ t(`home.dedicatedLanding.routing.modes.${activeRoutingMode}.failover`) }}
                    </dd>
                  </div>
                  <div class="border-l-2 border-gray-300 pl-3.5 dark:border-dark-600">
                    <dt class="text-xs text-gray-500 dark:text-dark-400">
                      {{ t('home.dedicatedLanding.routing.specOverflow') }}
                    </dt>
                    <dd class="mt-1 text-sm font-medium text-gray-900 dark:text-white">
                      {{ t(`home.dedicatedLanding.routing.modes.${activeRoutingMode}.overflow`) }}
                    </dd>
                  </div>
                  <div class="border-l-2 border-gray-300 pl-3.5 dark:border-dark-600">
                    <dt class="text-xs text-gray-500 dark:text-dark-400">
                      {{ t('home.dedicatedLanding.routing.specScenario') }}
                    </dt>
                    <dd class="mt-1 text-sm font-medium text-gray-900 dark:text-white">
                      {{ t(`home.dedicatedLanding.routing.modes.${activeRoutingMode}.scenario`) }}
                    </dd>
                  </div>
                </dl>
              </div>
            </div>
          </div>
        </div>
      </section>

      <!-- ==================== BAND 4: DECAY MARKETPLACE & BUYOUT WORKFLOW ==================== -->
      <section id="marketplace" class="scroll-mt-16 border-b border-gray-200/80 py-16 sm:py-24 dark:border-dark-800">
        <div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
          <div class="grid items-start gap-12 lg:grid-cols-12">
            <!-- Left: 3-Step Shelf-to-Ownership Lifecycle (6 Cols) -->
            <div class="lg:col-span-6">
              <p class="text-xs font-semibold uppercase tracking-[0.14em] text-primary-600 dark:text-primary-400">
                {{ t('home.dedicatedLanding.marketplace.overline') }}
              </p>
              <h2 class="mt-2 text-2xl font-semibold tracking-tight text-gray-900 sm:text-3xl dark:text-white">
                {{ t('home.dedicatedLanding.marketplace.title') }}
              </h2>
              <p class="mt-3 text-base leading-relaxed text-gray-600 dark:text-dark-300">
                {{ t('home.dedicatedLanding.marketplace.subtitle') }}
              </p>

              <ol class="mt-10 space-y-8">
                <li class="flex gap-4">
                  <span
                    class="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg border border-gray-200 bg-white font-mono text-xs font-semibold text-primary-600 shadow-sm dark:border-dark-700 dark:bg-dark-900 dark:text-primary-400"
                  >
                    01
                  </span>
                  <div>
                    <h3 class="text-base font-semibold text-gray-900 dark:text-white">
                      {{ t('home.dedicatedLanding.marketplace.steps.step1Title') }}
                    </h3>
                    <p class="mt-1.5 text-sm leading-relaxed text-gray-600 dark:text-dark-300">
                      {{ t('home.dedicatedLanding.marketplace.steps.step1Desc') }}
                    </p>
                  </div>
                </li>

                <li class="flex gap-4">
                  <span
                    class="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg border border-gray-200 bg-white font-mono text-xs font-semibold text-primary-600 shadow-sm dark:border-dark-700 dark:bg-dark-900 dark:text-primary-400"
                  >
                    02
                  </span>
                  <div>
                    <h3 class="text-base font-semibold text-gray-900 dark:text-white">
                      {{ t('home.dedicatedLanding.marketplace.steps.step2Title') }}
                    </h3>
                    <p class="mt-1.5 text-sm leading-relaxed text-gray-600 dark:text-dark-300">
                      {{ t('home.dedicatedLanding.marketplace.steps.step2Desc') }}
                    </p>
                  </div>
                </li>

                <li class="flex gap-4">
                  <span
                    class="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg border border-gray-200 bg-white font-mono text-xs font-semibold text-primary-600 shadow-sm dark:border-dark-700 dark:bg-dark-900 dark:text-primary-400"
                  >
                    03
                  </span>
                  <div>
                    <h3 class="text-base font-semibold text-gray-900 dark:text-white">
                      {{ t('home.dedicatedLanding.marketplace.steps.step3Title') }}
                    </h3>
                    <p class="mt-1.5 text-sm leading-relaxed text-gray-600 dark:text-dark-300">
                      {{ t('home.dedicatedLanding.marketplace.steps.step3Desc') }}
                    </p>
                  </div>
                </li>
              </ol>
            </div>

            <!-- Right: Interactive Decay Pricing & Buyout Mechanics Card (6 Cols) -->
            <div class="lg:col-span-6">
              <div
                class="overflow-hidden rounded-xl border border-gray-200 bg-white shadow-card dark:border-dark-700 dark:bg-dark-900"
              >
                <div
                  class="flex flex-wrap items-center justify-between gap-2 border-b border-gray-200 bg-gray-50/90 px-5 py-3.5 dark:border-dark-700 dark:bg-dark-950/70"
                >
                  <span class="text-xs font-semibold text-gray-700 dark:text-dark-200">
                    {{ t('home.dedicatedLanding.marketplace.simulator.headerTitle') }}
                  </span>
                  <span
                    class="rounded bg-emerald-50 px-2 py-0.5 text-xs font-semibold text-emerald-700 dark:bg-emerald-950/60 dark:text-emerald-300"
                  >
                    {{ t('home.dedicatedLanding.marketplace.simulator.stockBadge') }}
                  </span>
                </div>

                <div class="p-5 sm:p-6">
                  <!-- Sample Shelf Account Summary -->
                  <div class="flex items-center gap-3">
                    <div
                      class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg border border-gray-200 bg-gray-50 text-gray-800 dark:border-dark-700 dark:bg-dark-800 dark:text-dark-100"
                    >
                      <PlatformIcon platform="anthropic" size="lg" />
                    </div>
                    <div>
                      <div class="text-sm font-semibold text-gray-900 dark:text-white">
                        {{ t('home.dedicatedLanding.marketplace.simulator.sampleAlias') }}
                      </div>
                      <div class="mt-0.5 text-xs text-gray-500 dark:text-dark-400">
                        {{ t('home.dedicatedLanding.marketplace.simulator.sampleMeta') }}
                      </div>
                    </div>
                  </div>

                  <!-- Interactive Day Timeline Selector -->
                  <div class="mt-5">
                    <div class="mb-2 text-xs font-medium text-gray-500 dark:text-dark-400">
                      {{ t('home.dedicatedLanding.marketplace.simulator.daySelectorLabel') }}
                    </div>
                    <div class="grid grid-cols-3 gap-2">
                      <button
                        v-for="tier in decayTiers"
                        :key="tier.id"
                        :id="`decay-${tier.id}`"
                        type="button"
                        @click="selectedDecayTier = tier.id"
                        :class="[
                          'rounded-lg border px-3 py-2 text-center text-xs font-medium transition-colors',
                          selectedDecayTier === tier.id
                            ? 'border-primary-600 bg-primary-50/70 font-semibold text-primary-800 dark:border-primary-400 dark:bg-primary-950/40 dark:text-primary-200'
                            : 'border-gray-200 bg-white text-gray-600 hover:bg-gray-50 dark:border-dark-700 dark:bg-dark-800 dark:text-dark-300'
                        ]"
                      >
                        {{ t(tier.labelKey) }}
                      </button>
                    </div>
                  </div>

                  <!-- Live Calculated Price Display -->
                  <div
                    class="mt-4 flex flex-wrap items-baseline justify-between gap-2 rounded-lg bg-gray-50 px-4 py-3.5 dark:bg-dark-950/70"
                  >
                    <div>
                      <div class="text-xs text-gray-500 dark:text-dark-400">
                        {{ t('home.dedicatedLanding.marketplace.simulator.currentPriceLabel') }}
                      </div>
                      <div class="mt-0.5 text-[11px] text-gray-400 dark:text-dark-500">
                        {{ t('home.dedicatedLanding.marketplace.simulator.basePriceNote') }}
                      </div>
                    </div>
                    <div class="flex items-baseline gap-2 tabular-nums">
                      <span
                        v-if="currentDecayTier.price < 220"
                        class="font-mono text-sm text-gray-400 line-through dark:text-dark-500"
                      >
                        ¥220
                      </span>
                      <span class="font-mono text-2xl font-bold text-gray-900 dark:text-white">
                        ¥{{ currentDecayTier.price }}
                      </span>
                    </div>
                  </div>

                  <!-- Two Ownership Options (Managed vs Buyout) -->
                  <div class="mt-5 divide-y divide-gray-100 border-t border-gray-100 text-xs dark:divide-dark-800 dark:border-dark-800">
                    <div class="py-3.5">
                      <div class="flex items-center gap-2 font-semibold text-gray-900 dark:text-white">
                        <Icon name="shield" size="sm" class="text-primary-600 dark:text-primary-400" />
                        <span>{{ t('home.dedicatedLanding.marketplace.simulator.leaseModeTitle') }}</span>
                      </div>
                      <p class="mt-1 leading-relaxed text-gray-600 dark:text-dark-300">
                        {{ t('home.dedicatedLanding.marketplace.simulator.leaseModeDesc') }}
                      </p>
                    </div>

                    <div class="pt-3.5">
                      <div class="flex items-center gap-2 font-semibold text-gray-900 dark:text-white">
                        <Icon name="key" size="sm" class="text-amber-600 dark:text-amber-400" />
                        <span>{{ t('home.dedicatedLanding.marketplace.simulator.buyoutModeTitle') }}</span>
                      </div>
                      <p class="mt-1 leading-relaxed text-gray-600 dark:text-dark-300">
                        {{ t('home.dedicatedLanding.marketplace.simulator.buyoutModeDesc') }}
                      </p>
                    </div>
                  </div>

                  <router-link
                    :to="primaryDedicatedPath"
                    id="marketplace-simulator-cta"
                    class="btn btn-primary mt-5 w-full justify-center py-2.5 text-sm font-semibold"
                  >
                    <span>{{ t('home.dedicatedLanding.marketplace.simulator.actionCta') }}</span>
                    <Icon name="arrowRight" size="sm" />
                  </router-link>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      <!-- ==================== BAND 5: TRANSPARENT COMPARISON MATRIX ==================== -->
      <section
        id="comparison"
        class="scroll-mt-16 border-b border-gray-200/80 bg-white py-16 sm:py-24 dark:border-dark-800 dark:bg-dark-900/40"
      >
        <div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
          <div class="max-w-3xl">
            <p class="text-xs font-semibold uppercase tracking-[0.14em] text-primary-600 dark:text-primary-400">
              {{ t('home.dedicatedLanding.comparison.overline') }}
            </p>
            <h2 class="mt-2 text-2xl font-semibold tracking-tight text-gray-900 sm:text-3xl dark:text-white">
              {{ t('home.dedicatedLanding.comparison.title') }}
            </h2>
            <p class="mt-3 text-base leading-relaxed text-gray-600 dark:text-dark-300">
              {{ t('home.dedicatedLanding.comparison.subtitle') }}
            </p>
          </div>

          <div class="table-container mt-10 bg-white dark:bg-dark-900">
            <table class="table min-w-[720px]">
              <thead>
                <tr>
                  <th class="w-1/5">{{ t('home.dedicatedLanding.comparison.colDimension') }}</th>
                  <th class="w-1/4">{{ t('home.dedicatedLanding.comparison.colOfficial') }}</th>
                  <th class="w-1/4">{{ t('home.dedicatedLanding.comparison.colSharedRelay') }}</th>
                  <th class="w-[30%] bg-primary-50/60 text-primary-900 dark:bg-primary-950/40 dark:text-primary-200">
                    <div class="flex flex-wrap items-center gap-2">
                      <span>{{ t('home.dedicatedLanding.comparison.colDedicated') }}</span>
                      <span
                        class="rounded bg-primary-600 px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wide text-white dark:bg-primary-500"
                      >
                        {{ t('home.dedicatedLanding.comparison.recommendedTag') }}
                      </span>
                    </div>
                  </th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="rowKey in comparisonRowKeys" :key="rowKey">
                  <td class="font-semibold text-gray-900 dark:text-white">
                    {{ t(`home.dedicatedLanding.comparison.rows.${rowKey}.dim`) }}
                  </td>
                  <td class="text-gray-600 dark:text-dark-300">
                    {{ t(`home.dedicatedLanding.comparison.rows.${rowKey}.official`) }}
                  </td>
                  <td class="text-gray-500 dark:text-dark-400">
                    {{ t(`home.dedicatedLanding.comparison.rows.${rowKey}.shared`) }}
                  </td>
                  <td class="bg-primary-50/30 font-medium text-gray-900 dark:bg-primary-950/20 dark:text-dark-100">
                    <div class="flex items-start gap-2">
                      <Icon name="checkCircle" size="sm" class="mt-0.5 shrink-0 text-emerald-600 dark:text-emerald-400" />
                      <span>{{ t(`home.dedicatedLanding.comparison.rows.${rowKey}.dedicated`) }}</span>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </section>

      <!-- ==================== BAND 6: ECOSYSTEM & THREE-LINE QUICK START ==================== -->
      <section id="ecosystem" class="scroll-mt-16 border-b border-gray-200/80 py-16 sm:py-24 dark:border-dark-800">
        <div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
          <div class="grid items-start gap-12 lg:grid-cols-12">
            <!-- Left: Upstream Platforms & Coding Tool Matrix (5 Cols) -->
            <div class="lg:col-span-5">
              <p class="text-xs font-semibold uppercase tracking-[0.14em] text-primary-600 dark:text-primary-400">
                {{ t('home.dedicatedLanding.ecosystem.overline') }}
              </p>
              <h2 class="mt-2 text-2xl font-semibold tracking-tight text-gray-900 sm:text-3xl dark:text-white">
                {{ t('home.dedicatedLanding.ecosystem.title') }}
              </h2>
              <p class="mt-3 text-base leading-relaxed text-gray-600 dark:text-dark-300">
                {{ t('home.dedicatedLanding.ecosystem.subtitle') }}
              </p>

              <!-- Native Upstream Platforms -->
              <div class="mt-8">
                <h3 class="text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-dark-400">
                  {{ t('home.dedicatedLanding.ecosystem.providersTitle') }}
                </h3>
                <div class="mt-3 grid grid-cols-2 gap-2.5">
                  <div
                    v-for="provider in supportedProviders"
                    :key="provider.id"
                    class="flex items-center justify-between gap-2 rounded-lg border border-gray-200 bg-white px-3 py-2.5 dark:border-dark-700 dark:bg-dark-900"
                  >
                    <div class="flex min-w-0 items-center gap-2">
                      <PlatformIcon :platform="provider.platform" size="sm" class="shrink-0 text-gray-800 dark:text-dark-100" />
                      <span class="truncate text-xs font-semibold text-gray-900 dark:text-white">{{ provider.label }}</span>
                    </div>
                    <span
                      class="shrink-0 whitespace-nowrap rounded bg-emerald-50 px-1.5 py-0.5 text-[10px] font-semibold text-emerald-700 dark:bg-emerald-950/60 dark:text-emerald-400"
                    >
                      {{ t('home.providers.supported') }}
                    </span>
                  </div>
                </div>
              </div>

              <!-- Compatible Developer Tools -->
              <div class="mt-6">
                <h3 class="text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-dark-400">
                  {{ t('home.dedicatedLanding.ecosystem.clientsTitle') }}
                </h3>
                <div class="mt-3 flex flex-wrap gap-2">
                  <span
                    v-for="client in compatibleClients"
                    :key="client"
                    class="rounded-md border border-gray-200 bg-white px-3 py-1.5 font-mono text-xs font-medium text-gray-700 dark:border-dark-700 dark:bg-dark-900 dark:text-dark-200"
                  >
                    {{ client }}
                  </span>
                </div>
              </div>

              <!-- Self-Service Key Usage Callout -->
              <div class="mt-6 pt-2">
                <router-link
                  to="/key-usage"
                  id="ecosystem-key-usage-link"
                  class="inline-flex items-center gap-2 text-sm font-semibold text-primary-600 transition-colors hover:text-primary-700 dark:text-primary-400 dark:hover:text-primary-300"
                >
                  <Icon name="chart" size="sm" />
                  <span>{{ t('home.dedicatedLanding.ecosystem.selfServiceKeyCheck') }}</span>
                  <Icon name="arrowRight" size="xs" />
                </router-link>
              </div>
            </div>

            <!-- Right: Interactive CLI / SDK Quick-Start Panel (7 Cols) -->
            <div class="lg:col-span-7">
              <div class="overflow-hidden rounded-xl border border-gray-800 bg-[#0d1117] text-gray-100 shadow-card">
                <div class="flex flex-wrap items-center justify-between gap-2 border-b border-white/10 bg-[#161b22] px-4 py-3">
                  <div class="flex flex-wrap items-center gap-1.5">
                    <button
                      v-for="snip in quickStartSnippets"
                      :key="snip.id"
                      :id="`quickstart-tab-${snip.id}`"
                      type="button"
                      @click="activeQuickStart = snip.id"
                      :class="[
                        'rounded-md px-3 py-1.5 font-mono text-xs font-medium transition-colors',
                        activeQuickStart === snip.id
                          ? 'bg-primary-600 text-white'
                          : 'text-gray-400 hover:bg-white/5 hover:text-white'
                      ]"
                    >
                      {{ snip.label }}
                    </button>
                  </div>

                  <button
                    id="quickstart-copy-btn"
                    type="button"
                    @click="copyQuickStartSnippet"
                    class="inline-flex items-center gap-1.5 rounded-md border border-white/15 bg-white/5 px-2.5 py-1 text-xs font-medium text-gray-200 transition-colors hover:bg-white/10"
                  >
                    <Icon :name="snippetCopied ? 'check' : 'copy'" size="xs" />
                    <span>
                      {{
                        snippetCopied
                          ? t('home.dedicatedLanding.ecosystem.copiedBtn')
                          : t('home.dedicatedLanding.ecosystem.copyBtn')
                      }}
                    </span>
                  </button>
                </div>

                <pre
                  class="overflow-x-auto p-5 font-mono text-xs leading-6 text-gray-200 sm:p-6 sm:text-[13px]"
                ><code>{{ currentQuickStartCode }}</code></pre>
              </div>
            </div>
          </div>
        </div>
      </section>

      <!-- ==================== BAND 7: EXECUTIVE BOTTOM CTA ==================== -->
      <section class="bg-white py-16 sm:py-20 dark:bg-dark-900/60">
        <div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
          <div
            class="flex flex-col items-start justify-between gap-8 rounded-xl border border-gray-200 bg-gray-50/80 p-8 sm:p-10 lg:flex-row lg:items-center dark:border-dark-700 dark:bg-dark-900"
          >
            <div class="max-w-2xl">
              <h2 class="text-2xl font-semibold tracking-tight text-gray-900 sm:text-3xl dark:text-white">
                {{ t('home.dedicatedLanding.bottomCta.title') }}
              </h2>
              <p class="mt-2.5 text-sm leading-relaxed text-gray-600 sm:text-base dark:text-dark-300">
                {{ t('home.dedicatedLanding.bottomCta.subtitle') }}
              </p>
            </div>

            <div class="flex shrink-0 flex-wrap items-center gap-3">
              <router-link
                :to="primaryDedicatedPath"
                id="bottom-primary-cta"
                class="btn btn-primary px-6 py-3 text-sm font-semibold"
              >
                <span>
                  {{
                    isAuthenticated
                      ? t('home.dedicatedLanding.bottomCta.primaryAuthed')
                      : t('home.dedicatedLanding.bottomCta.primaryGuest')
                  }}
                </span>
                <Icon name="arrowRight" size="sm" />
              </router-link>

              <router-link
                v-if="!isAuthenticated && registrationEnabled"
                to="/register"
                id="bottom-register-cta"
                class="btn btn-secondary px-5 py-3 text-sm font-medium"
              >
                {{ t('home.dedicatedLanding.bottomCta.secondaryRegister') }}
              </router-link>
              <router-link
                v-else
                to="/key-usage"
                id="bottom-key-usage-cta"
                class="btn btn-secondary px-5 py-3 text-sm font-medium"
              >
                {{ t('home.dedicatedLanding.bottomCta.secondaryKeyUsage') }}
              </router-link>
            </div>
          </div>
        </div>
      </section>
    </main>

    <!-- Structured Enterprise Footer -->
    <footer class="border-t border-gray-200 bg-gray-50 dark:border-dark-800 dark:bg-dark-950">
      <div class="mx-auto max-w-7xl px-4 py-12 sm:px-6 lg:px-8">
        <div class="grid gap-8 md:grid-cols-12">
          <!-- Brand Column -->
          <div class="md:col-span-6">
            <div class="flex items-center gap-3">
              <div
                class="flex h-8 w-8 shrink-0 items-center justify-center overflow-hidden rounded-lg border border-gray-200 bg-white p-1 dark:border-dark-700 dark:bg-dark-900"
              >
                <img :src="siteLogo || '/logo.svg'" alt="Logo" class="h-full w-full object-contain" />
              </div>
              <span class="text-base font-semibold text-gray-900 dark:text-white">{{ siteName }}</span>
            </div>
            <p class="mt-3 max-w-md text-sm leading-relaxed text-gray-500 dark:text-dark-400">
              {{ t('home.dedicatedLanding.footerCols.brandTagline') }}
            </p>
            <p v-if="contactInfo" class="mt-3 text-xs text-gray-500 dark:text-dark-400">
              {{ contactInfo }}
            </p>
          </div>

          <!-- Architecture Links -->
          <div class="md:col-span-3">
            <h3 class="text-xs font-semibold uppercase tracking-wider text-gray-900 dark:text-white">
              {{ t('home.dedicatedLanding.footerCols.productTitle') }}
            </h3>
            <ul class="mt-3 space-y-2.5 text-sm">
              <li>
                <a
                  href="#capabilities"
                  class="text-gray-500 transition-colors hover:text-gray-900 dark:text-dark-400 dark:hover:text-white"
                >
                  {{ t('home.dedicatedLanding.nav.capabilities') }}
                </a>
              </li>
              <li>
                <a
                  href="#routing"
                  class="text-gray-500 transition-colors hover:text-gray-900 dark:text-dark-400 dark:hover:text-white"
                >
                  {{ t('home.dedicatedLanding.nav.routing') }}
                </a>
              </li>
              <li>
                <a
                  href="#marketplace"
                  class="text-gray-500 transition-colors hover:text-gray-900 dark:text-dark-400 dark:hover:text-white"
                >
                  {{ t('home.dedicatedLanding.nav.marketplace') }}
                </a>
              </li>
              <li>
                <a
                  href="#comparison"
                  class="text-gray-500 transition-colors hover:text-gray-900 dark:text-dark-400 dark:hover:text-white"
                >
                  {{ t('home.dedicatedLanding.nav.comparison') }}
                </a>
              </li>
            </ul>
          </div>

          <!-- Self-Service & Resources -->
          <div class="md:col-span-3">
            <h3 class="text-xs font-semibold uppercase tracking-wider text-gray-900 dark:text-white">
              {{ t('home.dedicatedLanding.footerCols.toolsTitle') }}
            </h3>
            <ul class="mt-3 space-y-2.5 text-sm">
              <li>
                <router-link
                  to="/key-usage"
                  class="text-gray-500 transition-colors hover:text-gray-900 dark:text-dark-400 dark:hover:text-white"
                >
                  {{ t('home.dedicatedLanding.nav.keyUsage') }}
                </router-link>
              </li>
              <li v-if="showModelPlazaEntry">
                <router-link
                  to="/model-plaza"
                  class="text-gray-500 transition-colors hover:text-gray-900 dark:text-dark-400 dark:hover:text-white"
                >
                  {{ t('nav.modelPlaza') }}
                </router-link>
              </li>
              <li v-if="docUrl">
                <a
                  :href="docUrl"
                  target="_blank"
                  rel="noopener noreferrer"
                  class="text-gray-500 transition-colors hover:text-gray-900 dark:text-dark-400 dark:hover:text-white"
                >
                  {{ t('home.docs') }}
                </a>
              </li>
              <li>
                <a
                  :href="githubUrl"
                  target="_blank"
                  rel="noopener noreferrer"
                  class="text-gray-500 transition-colors hover:text-gray-900 dark:text-dark-400 dark:hover:text-white"
                >
                  GitHub
                </a>
              </li>
            </ul>
          </div>
        </div>

        <div
          class="mt-10 flex flex-col items-center justify-between gap-3 border-t border-gray-200/80 pt-6 text-xs text-gray-500 sm:flex-row dark:border-dark-800 dark:text-dark-400"
        >
          <p>&copy; {{ currentYear }} {{ siteName }}. {{ t('home.footer.allRightsReserved') }}</p>
          <div class="flex items-center gap-4">
            <router-link
              :to="isAuthenticated ? '/dedicated-accounts' : '/login'"
              class="font-medium text-gray-600 transition-colors hover:text-gray-900 dark:text-dark-300 dark:hover:text-white"
            >
              {{
                isAuthenticated
                  ? t('home.dedicatedLanding.nav.dedicatedCenter')
                  : t('home.login')
              }}
            </router-link>
          </div>
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore, useAppStore } from '@/stores'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import PlatformIcon from '@/components/common/PlatformIcon.vue'
import Icon from '@/components/icons/Icon.vue'
import { sanitizeUrl } from '@/utils/url'
import { FeatureFlags, isFeatureFlagEnabled } from '@/utils/featureFlags'
import type { GroupPlatform } from '@/types'

const { t } = useI18n()

const authStore = useAuthStore()
const appStore = useAppStore()

// Site settings - directly from appStore (already initialized from injected config)
const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'Sub2API')
const siteLogo = computed(() =>
  sanitizeUrl(appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '', {
    allowRelative: true,
    allowDataUrl: true
  })
)
const defaultGenericSubtitle = 'AI API Gateway Platform'
const siteSubtitle = computed(() => appStore.cachedPublicSettings?.site_subtitle || defaultGenericSubtitle)
const hasCustomSubtitle = computed(
  () => Boolean(siteSubtitle.value.trim()) && siteSubtitle.value.trim() !== defaultGenericSubtitle
)
const docUrl = computed(() => sanitizeUrl(appStore.cachedPublicSettings?.doc_url || appStore.docUrl || ''))
const contactInfo = computed(() => (appStore.cachedPublicSettings?.contact_info || appStore.contactInfo || '').trim())
const homeContent = computed(() => appStore.cachedPublicSettings?.home_content || '')
const hasHomeContent = computed(() => homeContent.value.trim().length > 0)
const compactHomeEnabled = computed(() => appStore.cachedPublicSettings?.compact_home_enabled === true)
const registrationEnabled = computed(() => appStore.cachedPublicSettings?.registration_enabled !== false)
const modelPlazaEnabled = computed(() => isFeatureFlagEnabled(FeatureFlags.modelPlaza))

const effectiveApiBaseUrl = computed(() => {
  const configured = sanitizeUrl(appStore.cachedPublicSettings?.api_base_url || appStore.apiBaseUrl || '')
  if (configured) {
    return configured.replace(/\/+$/, '')
  }
  if (typeof window !== 'undefined' && window.location?.origin) {
    return window.location.origin.replace(/\/+$/, '')
  }
  return 'https://api.example.com'
})

// Check if homeContent is a URL (for iframe display)
const isHomeContentUrl = computed(() => {
  const content = homeContent.value.trim()
  return content.startsWith('http://') || content.startsWith('https://')
})

// Theme
const isDark = ref(document.documentElement.classList.contains('dark'))

// GitHub URL
const githubUrl = 'https://github.com/Wei-Shaw/sub2api'

// Auth state
const isAuthenticated = computed(() => authStore.isAuthenticated)
const modelPlazaRequiresAuth = computed(
  () => appStore.cachedPublicSettings?.model_plaza_require_auth === true
)
const showModelPlazaEntry = computed(
  () => modelPlazaEnabled.value && (isAuthenticated.value || !modelPlazaRequiresAuth.value)
)
const isAdmin = computed(() => authStore.isAdmin)
const dashboardPath = computed(() => (isAdmin.value ? '/admin/dashboard' : '/dashboard'))
const primaryDedicatedPath = computed(() => (isAuthenticated.value ? '/dedicated-accounts' : '/login'))
const userInitial = computed(() => {
  const user = authStore.user
  if (!user || !user.email) return ''
  return user.email.charAt(0).toUpperCase()
})

// Current year for footer
const currentYear = computed(() => new Date().getFullYear())

// ==================== Interactive Showcase State ====================
const activeShowcaseTab = ref<'console' | 'terminal'>('console')

interface ShowcaseAccount {
  id: 'claude' | 'openai' | 'gemini'
  shortLabel: string
  platform: GroupPlatform
  nameKey: string
  groupKey: string
  concurrencyKey: string
  fiveHourRemaining: number
  sevenDayRemaining: number
  maskedKey: string
  hasResetCredits: boolean
}

const showcaseAccounts: ShowcaseAccount[] = [
  {
    id: 'claude',
    shortLabel: 'Claude Max',
    platform: 'anthropic',
    nameKey: 'home.dedicatedLanding.showcase.accounts.claude.name',
    groupKey: 'home.dedicatedLanding.showcase.accounts.claude.group',
    concurrencyKey: 'home.dedicatedLanding.showcase.accounts.claude.concurrency',
    fiveHourRemaining: 86,
    sevenDayRemaining: 74,
    maskedKey: 'sk-ded-ant01••••••••9f2a',
    hasResetCredits: true
  },
  {
    id: 'openai',
    shortLabel: 'Codex Pro',
    platform: 'openai',
    nameKey: 'home.dedicatedLanding.showcase.accounts.openai.name',
    groupKey: 'home.dedicatedLanding.showcase.accounts.openai.group',
    concurrencyKey: 'home.dedicatedLanding.showcase.accounts.openai.concurrency',
    fiveHourRemaining: 92,
    sevenDayRemaining: 81,
    maskedKey: 'sk-ded-oai04••••••••4c1e',
    hasResetCredits: false
  },
  {
    id: 'gemini',
    shortLabel: 'Gemini Ultra',
    platform: 'gemini',
    nameKey: 'home.dedicatedLanding.showcase.accounts.gemini.name',
    groupKey: 'home.dedicatedLanding.showcase.accounts.gemini.group',
    concurrencyKey: 'home.dedicatedLanding.showcase.accounts.gemini.concurrency',
    fiveHourRemaining: 95,
    sevenDayRemaining: 88,
    maskedKey: 'sk-ded-gem02••••••••7b3d',
    hasResetCredits: false
  }
]

const selectedShowcaseAccountId = ref<ShowcaseAccount['id']>('claude')
const currentShowcaseAccount = computed(
  () => showcaseAccounts.find((a) => a.id === selectedShowcaseAccountId.value) ?? showcaseAccounts[0]
)

// ==================== Core Capabilities ====================
const capabilityItems = [
  { key: 'isolation', icon: 'lock' },
  { key: 'zeroOverflow', icon: 'shield' },
  { key: 'quotaWindow', icon: 'chart' },
  { key: 'buyoutShield', icon: 'key' }
] as const

// ==================== Interactive Tri-Mode Routing ====================
type RoutingModeId = 'pinned' | 'pool' | 'shared'
const routingModes: Array<{ id: RoutingModeId }> = [
  { id: 'pinned' },
  { id: 'pool' },
  { id: 'shared' }
]
const activeRoutingMode = ref<RoutingModeId>('pinned')

// ==================== Interactive Decay Marketplace Simulator ====================
type DecayTierId = 'day1' | 'day3' | 'day7'
const decayTiers: Array<{ id: DecayTierId; labelKey: string; price: number }> = [
  { id: 'day1', labelKey: 'home.dedicatedLanding.marketplace.simulator.day1', price: 220 },
  { id: 'day3', labelKey: 'home.dedicatedLanding.marketplace.simulator.day3', price: 190 },
  { id: 'day7', labelKey: 'home.dedicatedLanding.marketplace.simulator.day7', price: 130 }
]
const selectedDecayTier = ref<DecayTierId>('day3')
const currentDecayTier = computed(
  () => decayTiers.find((d) => d.id === selectedDecayTier.value) ?? decayTiers[0]
)

// ==================== Comparison Matrix ====================
const comparisonRowKeys = ['isolation', 'quota', 'billing', 'recovery', 'privacy'] as const

// ==================== Ecosystem & Quick-Start Snippets ====================
const supportedProviders: Array<{ id: string; label: string; platform: GroupPlatform }> = [
  { id: 'claude', label: 'Claude (Anthropic)', platform: 'anthropic' },
  { id: 'openai', label: 'Codex (OpenAI)', platform: 'openai' },
  { id: 'gemini', label: 'Gemini (Google)', platform: 'gemini' },
  { id: 'antigravity', label: 'Antigravity', platform: 'antigravity' }
]

const compatibleClients = [
  'Claude Code',
  'Codex CLI',
  'Cursor',
  'Cline / Roo Code',
  'OpenCode',
  'MCP Clients',
  'Anthropic & OpenAI SDKs'
]

type QuickStartId = 'claude' | 'codex' | 'curl'
const quickStartSnippets: Array<{ id: QuickStartId; label: string }> = [
  { id: 'claude', label: 'Claude Code' },
  { id: 'codex', label: 'Codex CLI' },
  { id: 'curl', label: 'cURL (HTTP)' }
]
const activeQuickStart = ref<QuickStartId>('claude')
const snippetCopied = ref(false)
let copyResetTimer: ReturnType<typeof setTimeout> | null = null

const currentQuickStartCode = computed(() => {
  const base = effectiveApiBaseUrl.value
  if (activeQuickStart.value === 'claude') {
    return [
      '# Bind your auto-issued 1:1 Pinned Dedicated API Key',
      `export ANTHROPIC_BASE_URL="${base}"`,
      'export ANTHROPIC_AUTH_TOKEN="sk-ded-your-1to1-pinned-key"',
      '',
      'claude'
    ].join('\n')
  }
  if (activeQuickStart.value === 'codex') {
    return [
      '# Connect OpenAI Codex CLI to your Dedicated Account slot',
      `export OPENAI_BASE_URL="${base}/v1"`,
      'export OPENAI_API_KEY="sk-ded-your-1to1-pinned-key"',
      '',
      'codex'
    ].join('\n')
  }
  return [
    `curl -X POST "${base}/v1/messages" \\`,
    '  -H "x-api-key: sk-ded-your-1to1-pinned-key" \\',
    '  -H "anthropic-version: 2023-06-01" \\',
    '  -H "content-type: application/json" \\',
    '  -d \'{"model":"claude-sonnet-4-5","max_tokens":1024,"messages":[{"role":"user","content":"Hello"}]}\''
  ].join('\n')
})

async function copyQuickStartSnippet() {
  try {
    await navigator.clipboard?.writeText(currentQuickStartCode.value)
    snippetCopied.value = true
    if (copyResetTimer) clearTimeout(copyResetTimer)
    copyResetTimer = setTimeout(() => {
      snippetCopied.value = false
    }, 2000)
  } catch {
    snippetCopied.value = false
  }
}

// Toggle theme
function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

// Initialize theme
function initTheme() {
  const savedTheme = localStorage.getItem('theme')
  if (
    savedTheme === 'dark' ||
    (!savedTheme && window.matchMedia('(prefers-color-scheme: dark)').matches)
  ) {
    isDark.value = true
    document.documentElement.classList.add('dark')
  }
}

onMounted(() => {
  initTheme()

  // Check auth state
  authStore.checkAuth()

  // Ensure public settings are loaded (will use cache if already loaded from injected config)
  if (!appStore.publicSettingsLoaded) {
    appStore.fetchPublicSettings()
  }
})

onBeforeUnmount(() => {
  if (copyResetTimer) {
    clearTimeout(copyResetTimer)
  }
})
</script>

<style scoped>
.landing-display-title {
  letter-spacing: -0.028em;
  line-height: 1.15;
}

.showcase-frame {
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.65),
    var(--shadow-card, 0 1px 3px rgba(20, 25, 31, 0.06));
}

:global(.dark) .showcase-frame {
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.04),
    var(--shadow-card, 0 2px 8px rgba(0, 0, 0, 0.35));
}

/* Terminal Container */
.terminal-container {
  position: relative;
  display: block;
}

.terminal-window {
  background: #0d1117;
  border-radius: 8px;
  border: 1px solid rgba(240, 246, 252, 0.12);
  overflow: hidden;
}

.terminal-header {
  display: flex;
  align-items: center;
  padding: 10px 14px;
  background: #161b22;
  border-bottom: 1px solid rgba(240, 246, 252, 0.08);
}

.terminal-buttons {
  display: flex;
  gap: 7px;
  flex-shrink: 0;
}

.terminal-buttons span {
  width: 10px;
  height: 10px;
  border-radius: 50%;
}

.btn-close {
  background: #ef4444;
}
.btn-minimize {
  background: #eab308;
}
.btn-maximize {
  background: #22c55e;
}

.terminal-title {
  flex: 1;
  text-align: center;
  font-size: 12px;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  color: #8b949e;
  padding: 0 8px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.terminal-body {
  padding: 18px 20px;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 12.5px;
  line-height: 1.9;
}

.code-line {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.code-prompt {
  color: #3fb950;
  font-weight: 600;
}
.code-cmd {
  color: #58a6ff;
}
.code-flag {
  color: #79c0ff;
}
.code-url {
  color: rgb(var(--color-primary-300));
  word-break: break-all;
}
.code-comment {
  color: #8b949e;
}
.code-success {
  color: #3fb950;
  background: rgba(63, 185, 80, 0.14);
  padding: 1px 7px;
  border-radius: 4px;
  font-weight: 600;
}
.code-response {
  color: #e3b341;
}
</style>
