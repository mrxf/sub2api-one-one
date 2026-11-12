import type { Account, CreateAccountRequest, UpdateAccountRequest } from '@/types'

export type DedicatedSaleFormMode = 'none' | 'listed' | 'sold'

export interface DedicatedSaleFormState {
  sale_status: DedicatedSaleFormMode
  display_alias: string
  sale_price: number | null
  daily_price_decay: number | null
  min_sale_price: number | null
  sale_description: string
  owner_user_id: number | null
  owner_user_email: string
  buyout_price: number | null
  buyout_info: string
  bought_out_at: string | null
}

const PUBLIC_SENSITIVE_PATTERNS: RegExp[] = [
  /[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}/,
  /[a-zA-Z0-9._%+-]+(?:%40|\\u0040)[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}/i,
  /\borg-[A-Za-z0-9]{12,}\b/,
  /\bproj_[A-Za-z0-9]{12,}\b/,
  /\buser-[A-Za-z0-9]{12,}\b/,
  /\bauth0\|[A-Za-z0-9_-]{8,}\b/,
  /\beyJ[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}\b/,
  /\brt\.[A-Za-z0-9._-]{16,}\b/,
  /\bsk-[A-Za-z0-9_-]{16,}\b/
]

export function containsPublicSensitiveText(raw: string): boolean {
  const trimmed = raw.trim()
  if (!trimmed) return false
  return PUBLIC_SENSITIVE_PATTERNS.some((re) => re.test(trimmed))
}

export function createDefaultDedicatedSaleFormState(): DedicatedSaleFormState {
  return {
    sale_status: 'none',
    display_alias: '',
    sale_price: 200,
    daily_price_decay: 0,
    min_sale_price: null,
    sale_description: '',
    owner_user_id: null,
    owner_user_email: '',
    buyout_price: null,
    buyout_info: '',
    bought_out_at: null
  }
}

export function initDedicatedSaleFormFromAccount(account?: Account | null): DedicatedSaleFormState {
  if (!account) {
    return createDefaultDedicatedSaleFormState()
  }

  let mode: DedicatedSaleFormMode = 'none'
  if (account.sale_status === 'listed' || account.sale_status === 'locked') {
    mode = 'listed'
  } else if (account.sale_status === 'sold' || (account.owner_user_id && account.owner_user_id > 0)) {
    mode = 'sold'
  }

  return {
    sale_status: mode,
    display_alias: account.display_alias ?? '',
    sale_price: typeof account.sale_price === 'number' && account.sale_price > 0 ? account.sale_price : 200,
    daily_price_decay: typeof account.daily_price_decay === 'number' ? account.daily_price_decay : 0,
    min_sale_price:
      typeof account.min_sale_price === 'number' && account.min_sale_price > 0 ? account.min_sale_price : null,
    sale_description: account.sale_description ?? '',
    owner_user_id: account.owner_user_id && account.owner_user_id > 0 ? account.owner_user_id : null,
    owner_user_email: account.owner_user_email ?? '',
    buyout_price:
      typeof account.buyout_price === 'number' && account.buyout_price > 0 ? account.buyout_price : null,
    buyout_info: account.buyout_info ?? '',
    bought_out_at: account.bought_out_at ?? null
  }
}

export function validateDedicatedSaleForm(
  state: DedicatedSaleFormState,
  t: (key: string) => string
): string | null {
  if (state.sale_status !== 'none') {
    if (
      containsPublicSensitiveText(state.display_alias) ||
      containsPublicSensitiveText(state.sale_description)
    ) {
      return t('dedicatedAccounts.admin.validation.publicSensitiveForbidden')
    }
  }

  if (state.sale_status === 'listed') {
    const price = Number(state.sale_price)
    if (!Number.isFinite(price) || price <= 0) {
      return t('dedicatedAccounts.admin.validation.salePriceRequired')
    }
    const decay = state.daily_price_decay == null ? 0 : Number(state.daily_price_decay)
    if (!Number.isFinite(decay) || decay < 0) {
      return t('dedicatedAccounts.admin.validation.decayNegative')
    }
    if (state.min_sale_price != null && state.min_sale_price > 0) {
      const minPrice = Number(state.min_sale_price)
      if (!Number.isFinite(minPrice) || minPrice > price) {
        return t('dedicatedAccounts.admin.validation.minPriceInvalid')
      }
    }
  } else if (state.sale_status === 'sold') {
    const uid = Number(state.owner_user_id)
    if (!Number.isInteger(uid) || uid <= 0) {
      return t('dedicatedAccounts.admin.validation.ownerRequired')
    }
  }

  if (state.sale_status !== 'none' && state.buyout_price != null) {
    const buyoutPrice = Number(state.buyout_price)
    if (!Number.isFinite(buyoutPrice) || buyoutPrice < 0) {
      return t('dedicatedAccounts.admin.validation.buyoutPriceNegative')
    }
    if (buyoutPrice > 0 && !state.buyout_info.trim()) {
      return t('dedicatedAccounts.admin.validation.buyoutInfoRequired')
    }
  }

  return null
}

export function buildDedicatedSaleCreatePatch(
  state: DedicatedSaleFormState
): Partial<CreateAccountRequest> {
  const buyoutPrice =
    state.buyout_price != null && Number(state.buyout_price) > 0 ? Number(state.buyout_price) : 0
  const buyoutInfo = buyoutPrice > 0 ? state.buyout_info.trim() : null

  if (state.sale_status === 'listed') {
    const price = Number(state.sale_price ?? 0)
    const decay = Math.max(0, Number(state.daily_price_decay ?? 0) || 0)
    const minPrice =
      state.min_sale_price != null && Number(state.min_sale_price) > 0
        ? Number(state.min_sale_price)
        : 0.01
    return {
      sale_status: 'listed',
      display_alias: state.display_alias.trim() || null,
      sale_price: price,
      daily_price_decay: decay,
      min_sale_price: minPrice,
      sale_description: state.sale_description.trim() || null,
      buyout_price: buyoutPrice,
      buyout_info: buyoutInfo
    }
  }

  if (state.sale_status === 'sold') {
    return {
      sale_status: 'sold',
      owner_user_id: Number(state.owner_user_id),
      display_alias: state.display_alias.trim() || null,
      buyout_price: buyoutPrice,
      buyout_info: buyoutInfo
    }
  }

  return {}
}

export function buildDedicatedSaleUpdatePatch(
  state: DedicatedSaleFormState,
  options?: { includeAlias?: boolean }
): Partial<UpdateAccountRequest> {
  const includeAlias = options?.includeAlias !== false
  const buyoutPrice =
    state.buyout_price != null && Number(state.buyout_price) > 0 ? Number(state.buyout_price) : 0

  if (state.sale_status === 'listed') {
    const price = Number(state.sale_price ?? 0)
    const decay = Math.max(0, Number(state.daily_price_decay ?? 0) || 0)
    const minPrice =
      state.min_sale_price != null && Number(state.min_sale_price) > 0
        ? Number(state.min_sale_price)
        : 0.01
    const patch: Partial<UpdateAccountRequest> = {
      sale_status: 'listed',
      clear_owner_user_id: true,
      sale_price: price,
      daily_price_decay: decay,
      min_sale_price: minPrice,
      sale_description: state.sale_description.trim()
    }
    if (includeAlias) {
      patch.display_alias = state.display_alias.trim()
      patch.buyout_price = buyoutPrice
      patch.buyout_info = state.buyout_info.trim()
    }
    return patch
  }

  if (state.sale_status === 'sold') {
    const patch: Partial<UpdateAccountRequest> = {
      sale_status: 'sold',
      owner_user_id: Number(state.owner_user_id),
      clear_owner_user_id: false
    }
    if (includeAlias) {
      patch.display_alias = state.display_alias.trim()
      patch.buyout_price = buyoutPrice
      patch.buyout_info = state.buyout_info.trim()
    }
    return patch
  }

  const patch: Partial<UpdateAccountRequest> = {
    sale_status: 'none',
    clear_owner_user_id: true,
    sale_price: 0,
    daily_price_decay: 0,
    min_sale_price: 0,
    sale_description: ''
  }
  if (includeAlias) {
    patch.display_alias = ''
    patch.buyout_price = 0
  }
  return patch
}
