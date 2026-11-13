/**
 * Local Skin Manager (CSS-only skin switching without modifying component DOM)
 *
 * Supported skins:
 * - 'default': Original Teal & glassmorphism theme (:root defaults in style.css)
 * - 'pro': Concise Stripe / GitHub Business Slate-Blue skin (styles/skins/professional.css)
 *
 * Local switching entry points:
 * 1. URL query parameter: `?skin=pro` or `?skin=default` (persists to localStorage)
 * 2. Keyboard shortcut: `Alt + Shift + S` anywhere outside text inputs
 * 3. Browser console / script API: `window.__SUB2API_SKIN__.setSkin('pro' | 'default')` or `.toggleSkin()`
 * 4. Build-time env default: `VITE_DEFAULT_SKIN=default|pro`
 */

export type SkinName = 'default' | 'pro'

export interface SkinController {
  getSkin: () => SkinName
  setSkin: (skin: SkinName) => SkinName
  toggleSkin: () => SkinName
}

export const SKIN_STORAGE_KEY = 'skin'
export const DEFAULT_SKIN: SkinName = 'pro'

let shortcutBound = false

export function isValidSkin(value: unknown): value is SkinName {
  return value === 'default' || value === 'pro'
}

function readStoredSkin(): SkinName | null {
  try {
    const saved = window.localStorage.getItem(SKIN_STORAGE_KEY)
    return isValidSkin(saved) ? saved : null
  } catch {
    return null
  }
}

function writeStoredSkin(skin: SkinName): void {
  try {
    window.localStorage.setItem(SKIN_STORAGE_KEY, skin)
  } catch {
    // Ignore storage quota / privacy mode errors
  }
}

function resolveEnvDefaultSkin(): SkinName {
  const envSkin = import.meta.env?.VITE_DEFAULT_SKIN
  return isValidSkin(envSkin) ? envSkin : DEFAULT_SKIN
}

export function resolveInitialSkin(search?: string): { skin: SkinName; fromQuery: boolean } {
  const queryString = search ?? (typeof window !== 'undefined' ? window.location.search : '')
  if (queryString) {
    const params = new URLSearchParams(queryString)
    const querySkin = params.get('skin')
    if (isValidSkin(querySkin)) {
      return { skin: querySkin, fromQuery: true }
    }
  }

  if (typeof window !== 'undefined') {
    const stored = readStoredSkin()
    if (stored) {
      return { skin: stored, fromQuery: false }
    }
  }

  return { skin: resolveEnvDefaultSkin(), fromQuery: false }
}

export function applySkinToDocument(skin: SkinName): void {
  if (typeof document === 'undefined') return
  document.documentElement.dataset.skin = skin
}

export function getSkin(): SkinName {
  if (typeof document !== 'undefined') {
    const current = document.documentElement.dataset.skin
    if (isValidSkin(current)) {
      return current
    }
  }
  return resolveInitialSkin().skin
}

export function setSkin(skin: SkinName): SkinName {
  const target: SkinName = isValidSkin(skin) ? skin : DEFAULT_SKIN
  applySkinToDocument(target)
  if (typeof window !== 'undefined') {
    writeStoredSkin(target)
    window.dispatchEvent(new CustomEvent('sub2api:skin-change', { detail: { skin: target } }))
  }
  return target
}

export function toggleSkin(): SkinName {
  const next: SkinName = getSkin() === 'pro' ? 'default' : 'pro'
  return setSkin(next)
}

function isEditableTarget(target: EventTarget | null): boolean {
  if (!(target instanceof HTMLElement)) return false
  if (target.isContentEditable) return true
  const tag = target.tagName
  return tag === 'INPUT' || tag === 'TEXTAREA' || tag === 'SELECT'
}

function handleSkinShortcut(event: KeyboardEvent): void {
  if (!event.altKey || !event.shiftKey || event.ctrlKey || event.metaKey) return
  if (event.code !== 'KeyS' && event.key.toLowerCase() !== 's') return
  if (isEditableTarget(event.target)) return

  event.preventDefault()
  toggleSkin()
}

export function initSkin(): SkinName {
  const { skin, fromQuery } = resolveInitialSkin()
  applySkinToDocument(skin)

  if (typeof window !== 'undefined') {
    if (fromQuery) {
      writeStoredSkin(skin)
    }

    window.__SUB2API_SKIN__ = {
      getSkin,
      setSkin,
      toggleSkin
    }

    if (!shortcutBound) {
      window.addEventListener('keydown', handleSkinShortcut)
      shortcutBound = true
    }
  }

  return skin
}

/** Internal test helper to reset keydown listener state between unit tests */
export function _resetSkinShortcutForTests(): void {
  if (typeof window !== 'undefined' && shortcutBound) {
    window.removeEventListener('keydown', handleSkinShortcut)
  }
  shortcutBound = false
}
