/**
 * Dedicated Accounts & Buyout Marketplace API endpoints (User-facing)
 */

import { apiClient } from './client'
import type {
  DedicatedBuyoutInfoResponse,
  DedicatedBuyoutResult,
  DedicatedMarketItem,
  DedicatedPurchaseResult,
  DedicatedQuotaSnapshot,
  DedicatedResetCreditRedeemResult,
  DedicatedResetCreditStatus,
  UserDedicatedAccountView
} from '@/types'

function buildIdempotencyKey(prefix: string, id: number): string {
  const rand = globalThis.crypto?.randomUUID?.() ?? `${Date.now()}-${Math.random().toString(36).slice(2)}`
  return `${prefix}-${id}-${rand}`
}

export const dedicatedAccountsAPI = {
  /** List pre-listed dedicated accounts on the buyout marketplace */
  async listMarket(platform?: string): Promise<DedicatedMarketItem[]> {
    const params: Record<string, string> = {}
    if (platform && platform !== 'all') {
      params.platform = platform
    }
    const { data } = await apiClient.get<{ items: DedicatedMarketItem[] }>('/dedicated-accounts/market', {
      params
    })
    return data?.items ?? []
  },

  /** Buy a pre-listed dedicated account immediately using platform wallet balance */
  async buyWithBalance(accountId: number): Promise<DedicatedPurchaseResult> {
    const { data } = await apiClient.post<DedicatedPurchaseResult>(
      `/dedicated-accounts/market/${accountId}/buy-with-balance`,
      undefined,
      {
        headers: {
          'Idempotency-Key': buildIdempotencyKey('dedicated-buy-balance', accountId)
        }
      }
    )
    return data
  },

  /** Buy out a dedicated account (shelf or owned) using platform wallet balance and reveal plaintext credentials */
  async buyoutWithBalance(accountId: number): Promise<DedicatedBuyoutResult> {
    const { data } = await apiClient.post<DedicatedBuyoutResult>(
      `/dedicated-accounts/${accountId}/buyout-with-balance`,
      undefined,
      {
        headers: {
          'Idempotency-Key': buildIdempotencyKey('dedicated-buyout-balance', accountId)
        }
      }
    )
    return data
  },

  /** Fetch plaintext buyout info for an owned dedicated account that has already been bought out */
  async getBuyoutInfo(accountId: number): Promise<DedicatedBuyoutInfoResponse> {
    const { data } = await apiClient.get<DedicatedBuyoutInfoResponse>(
      `/dedicated-accounts/${accountId}/buyout-info`
    )
    return data
  },

  /** List user's owned dedicated accounts (strictly read-only view) */
  async listMyAccounts(): Promise<UserDedicatedAccountView[]> {
    const { data } = await apiClient.get<{ items: UserDedicatedAccountView[] }>('/dedicated-accounts')
    return data?.items ?? []
  },

  /** Batch fetch real-time 5H / 7D quota usage for owned dedicated accounts */
  async getMyAccountsUsage(
    accountIds: number[],
    force: boolean = false
  ): Promise<Record<string, DedicatedQuotaSnapshot>> {
    const { data } = await apiClient.post<{ usage: Record<string, DedicatedQuotaSnapshot> }>(
      '/dedicated-accounts/usage',
      {
        account_ids: accountIds,
        force
      }
    )
    return data?.usage ?? {}
  },

  /** Query official reset credit availability for an owned dedicated account */
  async getResetCredits(accountId: number, refresh: boolean = false): Promise<DedicatedResetCreditStatus> {
    const { data } = await apiClient.get<DedicatedResetCreditStatus>(
      `/dedicated-accounts/${accountId}/reset-credits`,
      {
        params: refresh ? { refresh: '1' } : undefined
      }
    )
    return data
  },

  /** Redeem 1 official reset credit on an owned dedicated account */
  async redeemResetCredit(accountId: number): Promise<DedicatedResetCreditRedeemResult> {
    const { data } = await apiClient.post<DedicatedResetCreditRedeemResult>(
      `/dedicated-accounts/${accountId}/reset-credits/redeem`,
      undefined,
      {
        headers: {
          'Idempotency-Key': buildIdempotencyKey('dedicated-redeem-credit', accountId)
        }
      }
    )
    return data
  }
}

export default dedicatedAccountsAPI

