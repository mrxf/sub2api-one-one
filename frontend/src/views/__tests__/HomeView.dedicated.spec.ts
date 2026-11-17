import { beforeEach, describe, expect, it, vi } from 'vitest'
import { mount, RouterLinkStub } from '@vue/test-utils'

import HomeView from '../HomeView.vue'

const { appStore, authStore } = vi.hoisted(() => ({
  appStore: {
    cachedPublicSettings: {} as Record<string, unknown>,
    siteName: 'Sub2API',
    siteLogo: '',
    docUrl: '',
    apiBaseUrl: '',
    contactInfo: '',
    publicSettingsLoaded: true,
    fetchPublicSettings: vi.fn()
  },
  authStore: {
    isAuthenticated: false,
    isAdmin: false,
    user: null as { email?: string } | null,
    checkAuth: vi.fn()
  }
}))

vi.mock('@/stores', () => ({
  useAppStore: () => appStore,
  useAuthStore: () => authStore
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => appStore
}))

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key })
  }
})

function mountDedicatedHome(settings: Record<string, unknown> = {}) {
  appStore.cachedPublicSettings = {
    site_name: 'Sub2API Dedicated',
    site_subtitle: 'AI API Gateway Platform',
    registration_enabled: true,
    api_base_url: 'https://gateway.example.com/',
    ...settings
  }

  return mount(HomeView, {
    global: {
      stubs: {
        RouterLink: RouterLinkStub,
        LocaleSwitcher: { template: '<div data-testid="locale-switcher" />' },
        Icon: { template: '<span data-testid="icon" />' }
      }
    }
  })
}

describe('HomeView dedicated accounts landing page', () => {
  beforeEach(() => {
    authStore.isAuthenticated = false
    authStore.isAdmin = false
    authStore.user = null
    authStore.checkAuth.mockClear()
    appStore.fetchPublicSettings.mockClear()
    localStorage.clear()
    vi.spyOn(window, 'matchMedia').mockReturnValue({ matches: false } as MediaQueryList)
  })

  it('renders a single h1 and all core dedicated account sections', () => {
    const wrapper = mountDedicatedHome()

    expect(wrapper.find('[data-testid="dedicated-home"]').exists()).toBe(true)
    expect(wrapper.findAll('h1')).toHaveLength(1)
    expect(wrapper.find('#capabilities').exists()).toBe(true)
    expect(wrapper.find('#routing').exists()).toBe(true)
    expect(wrapper.find('#marketplace').exists()).toBe(true)
    expect(wrapper.find('#comparison').exists()).toBe(true)
    expect(wrapper.find('#ecosystem').exists()).toBe(true)
  })

  it('switches showcase account slots and toggles between live console and gateway terminal', async () => {
    const wrapper = mountDedicatedHome()

    expect(wrapper.text()).toContain('sk-ded-ant01••••••••9f2a')

    await wrapper.get('#showcase-account-openai').trigger('click')
    expect(wrapper.text()).toContain('sk-ded-oai04••••••••4c1e')

    await wrapper.get('#showcase-account-gemini').trigger('click')
    expect(wrapper.text()).toContain('sk-ded-gem02••••••••7b3d')

    await wrapper.get('#showcase-tab-terminal').trigger('click')
    expect(wrapper.get('.terminal-container').text()).toContain(
      'https://gateway.example.com/v1/messages'
    )
  })

  it('switches tri-mode routing contracts and blocks public pool overflow for dedicated modes', async () => {
    const wrapper = mountDedicatedHome()

    expect(wrapper.get('#routing').text()).toContain('mode = pinned')
    expect(wrapper.get('#routing').text()).toContain(
      'home.dedicatedLanding.routing.publicPoolBlocked'
    )

    await wrapper.get('#routing-tab-pool').trigger('click')
    expect(wrapper.get('#routing').text()).toContain('mode = pool')
    expect(wrapper.get('#routing').text()).toContain(
      'home.dedicatedLanding.routing.publicPoolBlocked'
    )

    await wrapper.get('#routing-tab-shared').trigger('click')
    expect(wrapper.get('#routing').text()).toContain('mode = shared')
    expect(wrapper.get('#routing').text()).not.toContain(
      'home.dedicatedLanding.routing.publicPoolBlocked'
    )
  })

  it('updates live decay marketplace pricing across Day 1, Day 3, and Day 7 floor tiers', async () => {
    const wrapper = mountDedicatedHome()

    // Default preview is Day 3 (¥190 with ¥220 strikethrough)
    expect(wrapper.get('#marketplace').text()).toContain('¥190')

    await wrapper.get('#decay-day1').trigger('click')
    expect(wrapper.get('#marketplace').text()).toContain('¥220')

    await wrapper.get('#decay-day7').trigger('click')
    expect(wrapper.get('#marketplace').text()).toContain('¥130')
  })

  it('updates quick-start code snippets with sanitized api_base_url across CLI tools', async () => {
    const wrapper = mountDedicatedHome()

    expect(wrapper.get('#ecosystem pre').text()).toContain(
      'export ANTHROPIC_BASE_URL="https://gateway.example.com"'
    )

    await wrapper.get('#quickstart-tab-codex').trigger('click')
    expect(wrapper.get('#ecosystem pre').text()).toContain(
      'export OPENAI_BASE_URL="https://gateway.example.com/v1"'
    )

    await wrapper.get('#quickstart-tab-curl').trigger('click')
    expect(wrapper.get('#ecosystem pre').text()).toContain(
      'curl -X POST "https://gateway.example.com/v1/messages"'
    )
  })

  it('adapts navigation and hero CTAs for guests, disabled registration, and authenticated users', () => {
    const guestWrapper = mountDedicatedHome({ registration_enabled: true })
    expect(guestWrapper.get('#hero-primary-cta').attributes('to') ?? guestWrapper.findComponent('#hero-primary-cta').props('to')).toBe('/login')
    expect(guestWrapper.find('#nav-register-btn').exists()).toBe(true)

    const closedRegWrapper = mountDedicatedHome({ registration_enabled: false })
    expect(closedRegWrapper.find('#nav-register-btn').exists()).toBe(false)

    authStore.isAuthenticated = true
    authStore.user = { email: 'alice@example.com' }
    const authedWrapper = mountDedicatedHome()
    expect(authedWrapper.findComponent('#hero-primary-cta').props('to')).toBe('/dedicated-accounts')
    expect(authedWrapper.findComponent('#nav-dedicated-center-btn').props('to')).toBe(
      '/dedicated-accounts'
    )
    expect(authedWrapper.findComponent('#nav-dashboard-btn').props('to')).toBe('/dashboard')
  })
})
