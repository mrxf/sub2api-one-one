import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import {
  DEFAULT_SKIN,
  SKIN_STORAGE_KEY,
  _resetSkinShortcutForTests,
  getSkin,
  initSkin,
  resolveInitialSkin,
  setSkin,
  toggleSkin
} from '@/utils/skin'

describe('skin manager', () => {
  const originalLocation = window.location.href

  beforeEach(() => {
    window.localStorage.clear()
    delete document.documentElement.dataset.skin
    delete window.__SUB2API_SKIN__
    _resetSkinShortcutForTests()
    window.history.replaceState({}, '', '/')
  })

  afterEach(() => {
    _resetSkinShortcutForTests()
    window.history.replaceState({}, '', originalLocation)
  })

  it('defaults to the professional skin when no query or storage is present', () => {
    const applied = initSkin()

    expect(applied).toBe(DEFAULT_SKIN)
    expect(applied).toBe('pro')
    expect(document.documentElement.dataset.skin).toBe('pro')
    expect(getSkin()).toBe('pro')
  })

  it('reads saved skin preference from localStorage even when <html> starts with static data-skin="pro"', () => {
    document.documentElement.dataset.skin = 'pro'
    expect(getSkin()).toBe('pro')

    window.localStorage.setItem(SKIN_STORAGE_KEY, 'default')

    const applied = initSkin()

    expect(applied).toBe('default')
    expect(document.documentElement.dataset.skin).toBe('default')
  })

  it('prioritizes ?skin= query parameter and persists it to localStorage', () => {
    window.localStorage.setItem(SKIN_STORAGE_KEY, 'pro')
    window.history.replaceState({}, '', '/login?skin=default')

    const applied = initSkin()

    expect(applied).toBe('default')
    expect(document.documentElement.dataset.skin).toBe('default')
    expect(window.localStorage.getItem(SKIN_STORAGE_KEY)).toBe('default')
  })

  it('ignores invalid skin values from query or storage', () => {
    window.localStorage.setItem(SKIN_STORAGE_KEY, 'neon-cyberpunk')
    expect(resolveInitialSkin('?skin=unknown').skin).toBe('pro')
  })

  it('toggles between pro and default via API and Alt+Shift+S shortcut', () => {
    initSkin()
    expect(getSkin()).toBe('pro')

    expect(toggleSkin()).toBe('default')
    expect(document.documentElement.dataset.skin).toBe('default')
    expect(window.localStorage.getItem(SKIN_STORAGE_KEY)).toBe('default')

    window.dispatchEvent(
      new KeyboardEvent('keydown', {
        key: 'S',
        code: 'KeyS',
        altKey: true,
        shiftKey: true,
        bubbles: true
      })
    )
    expect(getSkin()).toBe('pro')
    expect(document.documentElement.dataset.skin).toBe('pro')

    // Does not trigger when typing inside an input element
    const input = document.createElement('input')
    document.body.appendChild(input)
    input.dispatchEvent(
      new KeyboardEvent('keydown', {
        key: 'S',
        code: 'KeyS',
        altKey: true,
        shiftKey: true,
        bubbles: true
      })
    )
    expect(getSkin()).toBe('pro')
    input.remove()

    expect(window.__SUB2API_SKIN__?.setSkin('default')).toBe('default')
    expect(setSkin('pro')).toBe('pro')
  })
})
