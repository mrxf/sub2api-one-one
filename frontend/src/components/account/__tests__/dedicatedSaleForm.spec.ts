import { describe, expect, it } from 'vitest'
import {
  buildDedicatedSaleCreatePatch,
  buildDedicatedSaleUpdatePatch,
  containsPublicSensitiveText,
  createDefaultDedicatedSaleFormState,
  initDedicatedSaleFormFromAccount,
  validateDedicatedSaleForm
} from '../dedicatedSaleForm'
import type { Account } from '@/types'

const t = (key: string) => key

describe('dedicatedSaleForm', () => {
  it('detects sensitive emails, encoded emails, tokens, and upstream org/user IDs in public fields', () => {
    expect(containsPublicSensitiveText('Claude Max 20x 独享 #08')).toBe(false)
    expect(containsPublicSensitiveText('满额 5H 窗口，稳定低延迟')).toBe(false)
    expect(containsPublicSensitiveText('账号 owner@example.com 独享')).toBe(true)
    expect(containsPublicSensitiveText('owner%40example.com')).toBe(true)
    expect(containsPublicSensitiveText('org-SyntheticOrgId12345678')).toBe(true)
    expect(containsPublicSensitiveText('user-SyntheticUserId12345678')).toBe(true)
    expect(containsPublicSensitiveText('rt.1.SyntheticRefreshToken1234567890')).toBe(true)
    expect(containsPublicSensitiveText('sk-ant-api03-abcdef1234567890')).toBe(true)
  })

  it('rejects public alias/description leaks while allowing sensitive credentials inside buyout_info', () => {
    const state = createDefaultDedicatedSaleFormState()
    state.sale_status = 'listed'
    state.display_alias = 'secret@gmail.com'
    state.sale_price = 200
    expect(validateDedicatedSaleForm(state, t)).toBe(
      'dedicatedAccounts.admin.validation.publicSensitiveForbidden'
    )

    state.display_alias = 'OpenAI Pro 独享 #01'
    state.buyout_price = 580
    state.buyout_info = ''
    expect(validateDedicatedSaleForm(state, t)).toBe(
      'dedicatedAccounts.admin.validation.buyoutInfoRequired'
    )

    state.buyout_info = '邮箱: secret@gmail.com\n密码: P@ssw0rd123\n2FA: JBSWY3DPEHPK3PXP'
    expect(validateDedicatedSaleForm(state, t)).toBeNull()
  })

  it('hydrates and builds create/update payloads with buyout fields and preserves buyout in bulk mode', () => {
    const account = {
      id: 42,
      sale_status: 'sold',
      owner_user_id: 9,
      owner_user_email: 'buyer@site.test',
      display_alias: 'Claude Max 独享 #42',
      buyout_price: 699,
      buyout_info: 'account: raw@upstream.test / pass: 123456',
      bought_out_at: '2026-11-11T12:00:00Z'
    } as unknown as Account

    const hydrated = initDedicatedSaleFormFromAccount(account)
    expect(hydrated.sale_status).toBe('sold')
    expect(hydrated.buyout_price).toBe(699)
    expect(hydrated.buyout_info).toContain('raw@upstream.test')
    expect(hydrated.bought_out_at).toBe('2026-11-11T12:00:00Z')

    const createPatch = buildDedicatedSaleCreatePatch(hydrated)
    expect(createPatch).toMatchObject({
      sale_status: 'sold',
      owner_user_id: 9,
      display_alias: 'Claude Max 独享 #42',
      buyout_price: 699,
      buyout_info: 'account: raw@upstream.test / pass: 123456'
    })

    const singleUpdate = buildDedicatedSaleUpdatePatch(hydrated, { includeAlias: true })
    expect(singleUpdate.buyout_price).toBe(699)
    expect(singleUpdate.buyout_info).toBe('account: raw@upstream.test / pass: 123456')

    // Bulk mode (includeAlias: false) must NOT overwrite per-account alias or buyout secrets
    const bulkUpdate = buildDedicatedSaleUpdatePatch(hydrated, { includeAlias: false })
    expect(bulkUpdate.display_alias).toBeUndefined()
    expect(bulkUpdate.buyout_price).toBeUndefined()
    expect(bulkUpdate.buyout_info).toBeUndefined()
  })
})
