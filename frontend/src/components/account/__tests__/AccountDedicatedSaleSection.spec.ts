import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia } from 'pinia'
import { defineComponent, ref } from 'vue'
import { createI18n } from 'vue-i18n'
import en from '@/i18n/locales/en'
import zh from '@/i18n/locales/zh'
import AccountDedicatedSaleSection from '../AccountDedicatedSaleSection.vue'
import {
  createDefaultDedicatedSaleFormState,
  validateDedicatedSaleForm,
  type DedicatedSaleFormState
} from '../dedicatedSaleForm'

vi.mock('@/api/admin', () => ({
  adminAPI: {
    users: {
      list: vi.fn().mockResolvedValue({ items: [] })
    }
  }
}))

function createRealI18n(locale: 'zh' | 'en' = 'zh') {
  return createI18n({
    legacy: false,
    locale,
    fallbackLocale: 'en',
    messages: { zh, en },
    warnHtmlMessage: false
  })
}

describe('AccountDedicatedSaleSection with real vue-i18n compiler', () => {
  it('renders listed and sold configuration fields without vue-i18n linked-message (@) syntax crash in zh and en', async () => {
    for (const locale of ['zh', 'en'] as const) {
      const i18n = createRealI18n(locale)
      const Host = defineComponent({
        components: { AccountDedicatedSaleSection },
        setup() {
          const state = ref<DedicatedSaleFormState>(createDefaultDedicatedSaleFormState())
          return { state }
        },
        template: `<AccountDedicatedSaleSection v-model="state" :has-groups-selected="false" />`
      })

      const wrapper = mount(Host, {
        global: {
          plugins: [createPinia(), i18n]
        }
      })

      // Click option 2: Listed on Marketplace
      await wrapper.get('[data-testid="dedicated-mode-listed"]').trigger('click')
      expect(wrapper.vm.state.sale_status).toBe('listed')
      expect(wrapper.find('[data-testid="account-buyout-config-section"]').exists()).toBe(true)
      expect(wrapper.find('[data-testid="buyout-price-input"]').exists()).toBe(true)
      expect(wrapper.find('[data-testid="buyout-info-textarea"]').exists()).toBe(true)

      // Click option 3: Assigned to Dedicated User (1:1 / N:1)
      await wrapper.get('[data-testid="dedicated-mode-sold"]').trigger('click')
      expect(wrapper.vm.state.sale_status).toBe('sold')
      expect(wrapper.find('input[type="number"][min="1"]').exists()).toBe(true)
      expect(wrapper.find('[data-testid="account-buyout-config-section"]').exists()).toBe(true)

      // Verify publicSensitiveWarning and validation messages also compile cleanly via real t()
      wrapper.vm.state.display_alias = 'leak@example.com'
      await wrapper.vm.$nextTick()
      expect(wrapper.text()).toContain(
        locale === 'zh' ? '公开展示别名和商城卖点说明中严禁包含邮箱' : 'Security Guard'
      )
      const validationMsg = validateDedicatedSaleForm(wrapper.vm.state, (key) =>
        i18n.global.t(key)
      )
      expect(validationMsg).toContain('@')
    }
  })
})
