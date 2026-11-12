<template>
  <div
    class="rounded-xl border border-gray-200 bg-gray-50/70 p-4 dark:border-dark-600 dark:bg-dark-800/60"
    data-testid="account-dedicated-sale-section"
  >
    <div v-if="!hideHeader" class="mb-3">
      <div class="flex items-center gap-2">
        <Icon name="shoppingBag" size="sm" class="text-primary-600 dark:text-primary-400" />
        <h4 class="text-sm font-semibold text-gray-900 dark:text-white">
          {{ t('dedicatedAccounts.admin.sectionTitle') }}
        </h4>
      </div>
      <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
        {{ t('dedicatedAccounts.admin.sectionDesc') }}
      </p>
    </div>

    <!-- 3-Mode Selector -->
    <div class="grid grid-cols-1 gap-2.5 sm:grid-cols-3">
      <button
        v-for="option in modeOptions"
        :key="option.value"
        type="button"
        :disabled="disabled"
        :data-testid="`dedicated-mode-${option.value}`"
        :class="[
          'flex flex-col items-start rounded-lg border p-3 text-left transition-all',
          modelValue.sale_status === option.value
            ? option.activeClass
            : 'border-gray-200 bg-white hover:border-gray-300 dark:border-dark-600 dark:bg-dark-700/60 dark:hover:border-dark-500',
          disabled && 'cursor-not-allowed opacity-60'
        ]"
        @click="updateField('sale_status', option.value)"
      >
        <div class="flex w-full items-center justify-between gap-2">
          <span class="text-xs font-semibold text-gray-900 dark:text-white">
            {{ option.label }}
          </span>
          <span
            :class="[
              'h-2.5 w-2.5 flex-shrink-0 rounded-full',
              modelValue.sale_status === option.value
                ? option.dotClass
                : 'bg-gray-300 dark:bg-dark-500'
            ]"
          />
        </div>
        <p class="mt-1 text-[11px] leading-relaxed text-gray-500 dark:text-gray-400">
          {{ option.hint }}
        </p>
      </button>
    </div>

    <!-- Group Required Warning -->
    <div
      v-if="showGroupWarning && !authStore.isSimpleMode && modelValue.sale_status !== 'none' && !hasGroupsSelected"
      class="mt-3 flex items-start gap-2 rounded-lg border border-amber-200 bg-amber-50 px-3 py-2.5 text-xs text-amber-800 dark:border-amber-800/60 dark:bg-amber-900/20 dark:text-amber-300"
    >
      <Icon name="exclamationTriangle" size="sm" class="mt-0.5 flex-shrink-0 text-amber-500" />
      <span>{{ t('dedicatedAccounts.admin.groupRequiredWarning') }}</span>
    </div>

    <!-- Listed Mode Fields -->
    <div v-if="modelValue.sale_status === 'listed'" class="mt-4 space-y-3.5">
      <div v-if="!hideAlias">
        <label class="input-label">
          {{ t('dedicatedAccounts.admin.displayAlias') }}
        </label>
        <input
          :value="modelValue.display_alias"
          :disabled="disabled"
          type="text"
          class="input"
          :placeholder="t('dedicatedAccounts.admin.displayAliasPlaceholder')"
          @input="updateField('display_alias', ($event.target as HTMLInputElement).value)"
        />
        <p class="input-hint">
          {{ t('dedicatedAccounts.admin.displayAliasHint') }}
        </p>
      </div>

      <div class="grid grid-cols-1 gap-3 sm:grid-cols-3">
        <div>
          <label class="input-label">
            {{ t('dedicatedAccounts.admin.salePrice') }}
            <span class="text-red-500">*</span>
          </label>
          <input
            :value="modelValue.sale_price ?? ''"
            :disabled="disabled"
            type="number"
            min="0.01"
            step="0.01"
            class="input font-mono"
            :placeholder="t('dedicatedAccounts.admin.salePricePlaceholder')"
            @input="updateNumberField('sale_price', ($event.target as HTMLInputElement).value)"
          />
        </div>

        <div>
          <label class="input-label">
            {{ t('dedicatedAccounts.admin.dailyPriceDecay') }}
          </label>
          <input
            :value="modelValue.daily_price_decay ?? ''"
            :disabled="disabled"
            type="number"
            min="0"
            step="0.01"
            class="input font-mono"
            :placeholder="t('dedicatedAccounts.admin.dailyPriceDecayPlaceholder')"
            @input="updateNumberField('daily_price_decay', ($event.target as HTMLInputElement).value)"
          />
        </div>

        <div>
          <label class="input-label">
            {{ t('dedicatedAccounts.admin.minSalePrice') }}
          </label>
          <input
            :value="modelValue.min_sale_price ?? ''"
            :disabled="disabled"
            type="number"
            min="0"
            step="0.01"
            class="input font-mono"
            :placeholder="t('dedicatedAccounts.admin.minSalePricePlaceholder')"
            @input="updateNumberField('min_sale_price', ($event.target as HTMLInputElement).value)"
          />
        </div>
      </div>

      <!-- Live Price Decay Preview -->
      <div
        v-if="livePreview"
        class="flex items-center gap-2 rounded-lg border border-amber-200/80 bg-amber-50/70 px-3 py-2 text-xs text-amber-800 dark:border-amber-800/40 dark:bg-amber-900/20 dark:text-amber-300"
      >
        <Icon name="sparkles" size="xs" class="flex-shrink-0 text-amber-500" />
        <span>
          {{
            t('dedicatedAccounts.admin.livePricePreview', {
              day0: livePreview.day0,
              day3: livePreview.day3,
              day7: livePreview.day7
            })
          }}
        </span>
      </div>

      <div>
        <label class="input-label">
          {{ t('dedicatedAccounts.admin.saleDescription') }}
        </label>
        <textarea
          :value="modelValue.sale_description"
          :disabled="disabled"
          rows="2"
          class="input"
          :placeholder="t('dedicatedAccounts.admin.saleDescriptionPlaceholder')"
          @input="updateField('sale_description', ($event.target as HTMLTextAreaElement).value)"
        />
      </div>
    </div>

    <!-- Sold / Assigned Owner Mode Fields -->
    <div v-else-if="modelValue.sale_status === 'sold'" class="mt-4 space-y-3.5">
      <div v-if="!hideAlias">
        <label class="input-label">
          {{ t('dedicatedAccounts.admin.displayAlias') }}
        </label>
        <input
          :value="modelValue.display_alias"
          :disabled="disabled"
          type="text"
          class="input"
          :placeholder="t('dedicatedAccounts.admin.displayAliasPlaceholder')"
          @input="updateField('display_alias', ($event.target as HTMLInputElement).value)"
        />
        <p class="input-hint">
          {{ t('dedicatedAccounts.admin.displayAliasHint') }}
        </p>
      </div>

      <div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
        <div>
          <label class="input-label">
            {{ t('dedicatedAccounts.admin.ownerUserId') }}
            <span class="text-red-500">*</span>
          </label>
          <input
            :value="modelValue.owner_user_id ?? ''"
            :disabled="disabled"
            type="number"
            min="1"
            step="1"
            class="input font-mono"
            :placeholder="t('dedicatedAccounts.admin.ownerUserPlaceholder')"
            @input="handleManualOwnerIdInput(($event.target as HTMLInputElement).value)"
          />
        </div>

        <div class="relative">
          <label class="input-label">
            {{ t('common.search') }}
          </label>
          <input
            v-model="userSearchQuery"
            :disabled="disabled"
            type="text"
            class="input"
            :placeholder="t('dedicatedAccounts.admin.searchUserPlaceholder')"
            @input="handleUserSearchInput"
            @focus="showUserDropdown = userCandidates.length > 0"
          />

          <div
            v-if="showUserDropdown && userCandidates.length > 0"
            class="absolute z-30 mt-1 max-h-48 w-full overflow-y-auto rounded-lg border border-gray-200 bg-white py-1 shadow-lg dark:border-dark-600 dark:bg-dark-800"
          >
            <button
              v-for="candidate in userCandidates"
              :key="candidate.id"
              type="button"
              class="flex w-full items-center justify-between px-3 py-2 text-left text-xs hover:bg-primary-50 dark:hover:bg-dark-700"
              @click="selectOwnerCandidate(candidate)"
            >
              <div class="min-w-0 flex-1">
                <div class="truncate font-medium text-gray-900 dark:text-white">
                  {{ candidate.email }}
                </div>
                <div v-if="candidate.username" class="truncate text-[11px] text-gray-500 dark:text-gray-400">
                  {{ candidate.username }}
                </div>
              </div>
              <span class="ml-2 rounded bg-emerald-100 px-1.5 py-0.5 font-mono text-[11px] text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-300">
                UID #{{ candidate.id }}
              </span>
            </button>
          </div>
        </div>
      </div>

      <div
        v-if="modelValue.owner_user_id && modelValue.owner_user_id > 0"
        class="flex items-center justify-between rounded-lg border border-emerald-200 bg-emerald-50/70 px-3 py-2 text-xs text-emerald-800 dark:border-emerald-800/50 dark:bg-emerald-900/20 dark:text-emerald-300"
      >
        <span>
          {{
            t('dedicatedAccounts.admin.selectedOwner', {
              email: modelValue.owner_user_email || `#${modelValue.owner_user_id}`,
              id: modelValue.owner_user_id
            })
          }}
        </span>
        <span v-if="searchingUsers" class="text-[11px] opacity-75">{{ t('common.loading') }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { adminAPI } from '@/api/admin'
import { useAuthStore } from '@/stores'
import type { AdminUser } from '@/types'
import type { DedicatedSaleFormMode, DedicatedSaleFormState } from './dedicatedSaleForm'

const props = withDefaults(
  defineProps<{
    modelValue: DedicatedSaleFormState
    hasGroupsSelected?: boolean
    showGroupWarning?: boolean
    hideAlias?: boolean
    hideHeader?: boolean
    disabled?: boolean
  }>(),
  {
    hasGroupsSelected: true,
    showGroupWarning: true,
    hideAlias: false,
    hideHeader: false,
    disabled: false
  }
)

const emit = defineEmits<{
  'update:modelValue': [value: DedicatedSaleFormState]
}>()

const { t } = useI18n()
const authStore = useAuthStore()

const modeOptions = computed<
  Array<{
    value: DedicatedSaleFormMode
    label: string
    hint: string
    activeClass: string
    dotClass: string
  }>
>(() => [
  {
    value: 'none',
    label: t('dedicatedAccounts.admin.modeNone'),
    hint: t('dedicatedAccounts.admin.modeNoneHint'),
    activeClass:
      'border-primary-500 bg-primary-50/60 ring-1 ring-primary-500/40 dark:border-primary-500 dark:bg-primary-900/20',
    dotClass: 'bg-primary-500'
  },
  {
    value: 'listed',
    label: t('dedicatedAccounts.admin.modeListed'),
    hint: t('dedicatedAccounts.admin.modeListedHint'),
    activeClass:
      'border-amber-500 bg-amber-50/60 ring-1 ring-amber-500/40 dark:border-amber-500 dark:bg-amber-900/20',
    dotClass: 'bg-amber-500'
  },
  {
    value: 'sold',
    label: t('dedicatedAccounts.admin.modeSold'),
    hint: t('dedicatedAccounts.admin.modeSoldHint'),
    activeClass:
      'border-emerald-500 bg-emerald-50/60 ring-1 ring-emerald-500/40 dark:border-emerald-500 dark:bg-emerald-900/20',
    dotClass: 'bg-emerald-500'
  }
])

const updateField = <K extends keyof DedicatedSaleFormState>(
  key: K,
  value: DedicatedSaleFormState[K]
) => {
  emit('update:modelValue', {
    ...props.modelValue,
    [key]: value
  })
}

const updateNumberField = (
  key: 'sale_price' | 'daily_price_decay' | 'min_sale_price',
  raw: string
) => {
  const trimmed = raw.trim()
  if (!trimmed) {
    updateField(key, null)
    return
  }
  const parsed = Number(trimmed)
  updateField(key, Number.isFinite(parsed) ? parsed : null)
}

const handleManualOwnerIdInput = (raw: string) => {
  const trimmed = raw.trim()
  if (!trimmed) {
    emit('update:modelValue', {
      ...props.modelValue,
      owner_user_id: null,
      owner_user_email: ''
    })
    return
  }
  const parsed = Number.parseInt(trimmed, 10)
  emit('update:modelValue', {
    ...props.modelValue,
    owner_user_id: Number.isFinite(parsed) && parsed > 0 ? parsed : null,
    owner_user_email: ''
  })
}

const computePriceAtDay = (base: number, decay: number, floor: number, elapsedDays: number) => {
  const raw = Math.max(floor, base - elapsedDays * decay)
  return Math.round(raw * 100) / 100
}

const livePreview = computed(() => {
  const base = Number(props.modelValue.sale_price ?? 0)
  if (!Number.isFinite(base) || base <= 0) return null
  const decay = Math.max(0, Number(props.modelValue.daily_price_decay ?? 0) || 0)
  const rawMin = Number(props.modelValue.min_sale_price ?? 0)
  const floor = Number.isFinite(rawMin) && rawMin > 0 ? Math.min(rawMin, base) : 0.01
  return {
    day0: computePriceAtDay(base, decay, floor, 0).toFixed(2),
    day3: computePriceAtDay(base, decay, floor, 2).toFixed(2),
    day7: computePriceAtDay(base, decay, floor, 6).toFixed(2)
  }
})

// Debounced Admin User Search
const userSearchQuery = ref('')
const userCandidates = ref<AdminUser[]>([])
const showUserDropdown = ref(false)
const searchingUsers = ref(false)
let searchTimer: ReturnType<typeof setTimeout> | null = null

const runUserSearch = async (keyword: string) => {
  searchingUsers.value = true
  try {
    const res = await adminAPI.users.list(1, 8, { search: keyword })
    userCandidates.value = res.items || []
    showUserDropdown.value = userCandidates.value.length > 0
  } catch {
    userCandidates.value = []
    showUserDropdown.value = false
  } finally {
    searchingUsers.value = false
  }
}

const handleUserSearchInput = () => {
  if (searchTimer) {
    clearTimeout(searchTimer)
  }
  const keyword = userSearchQuery.value.trim()
  if (!keyword) {
    userCandidates.value = []
    showUserDropdown.value = false
    return
  }
  searchTimer = setTimeout(() => {
    void runUserSearch(keyword)
  }, 260)
}

const selectOwnerCandidate = (candidate: AdminUser) => {
  emit('update:modelValue', {
    ...props.modelValue,
    owner_user_id: candidate.id,
    owner_user_email: candidate.email
  })
  userSearchQuery.value = candidate.email
  showUserDropdown.value = false
}

onBeforeUnmount(() => {
  if (searchTimer) {
    clearTimeout(searchTimer)
  }
})
</script>
