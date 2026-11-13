import type { PublicSettings } from '@/types'
import type { SkinController } from '@/utils/skin'

declare global {
  interface Window {
    __APP_CONFIG__?: PublicSettings
    __SUB2API_SKIN__?: SkinController
  }
}

export {}
