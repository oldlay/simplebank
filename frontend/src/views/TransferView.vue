<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import Card from 'primevue/card'
import Button from 'primevue/button'
import Select from 'primevue/select'
import InputText from 'primevue/inputtext'
import FloatLabel from 'primevue/floatlabel'
import Toast from 'primevue/toast'
import { useToast } from 'primevue/usetoast'
import api from '@/api'
import type { Account, ListAccountResponse } from '@/types/account'
import { getAmount, toDecimalValue } from '@/types/account'
import type { CreateTransferResponse } from '@/types/transfer'

const toast = useToast()

const accounts = ref<Account[]>([])
const isLoading = ref(true)
const isSending = ref(false)

const fromAccountId = ref<number | null>(null)
const toOwner = ref('')
const toCurrency = ref('USD')
const amount = ref('')

// ── Derived ────────────────────────────────────────
const fromAccount = computed(() =>
  accounts.value.find((a) => a.id === fromAccountId.value) ?? null,
)

//const sourceCurrency = computed(() => fromAccount.value?.currency ?? 'USD')

const currencySymbol = computed(() => {
  const map: Record<string, string> = { USD: '$', EUR: '€', CAD: 'CA$' }
  return map[toCurrency.value] || toCurrency.value
})

const availableBalance = computed(() => {
  if (!fromAccount.value) return '0.00'
  return parseFloat(getAmount(fromAccount.value)).toLocaleString('en-US', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  })
})

const isInsufficient = computed(() => {
  if (!fromAccount.value || !amount.value) return false
  return parseFloat(amount.value) > parseFloat(getAmount(fromAccount.value))
})

const isSelfTransfer = computed(
  () => fromAccount.value && toOwner.value === fromAccount.value.owner && toCurrency.value === fromAccount.value.currency,
)

const isFormValid = computed(
  () =>
    fromAccountId.value !== null &&
    toOwner.value.trim().length > 0 &&
    toCurrency.value.length > 0 &&
    !isSelfTransfer.value &&
    parseFloat(amount.value) > 0 &&
    !isInsufficient.value,
)

const accountOptions = computed(() =>
  accounts.value.map((a) => ({
    label: `${a.currency} ····${String(a.id).slice(-4)} — ${currencyPrefix(a.currency)}${parseFloat(getAmount(a)).toFixed(2)}`,
    value: a.id,
  })),
)

const currencies = [
  { label: 'USD — US Dollar', value: 'USD' },
  { label: 'EUR — Euro', value: 'EUR' },
  { label: 'CAD — Canadian Dollar', value: 'CAD' },
]

function currencyPrefix(c: string): string {
  const map: Record<string, string> = { USD: '$', EUR: '€', CAD: 'CA$' }
  return map[c] || c
}

// When source account changes, default the transfer currency
watch(fromAccountId, (newId) => {
  const acct = accounts.value.find((a) => a.id === newId)
  if (acct) toCurrency.value = acct.currency
  toOwner.value = ''
  amount.value = ''
})

// ── API calls ──────────────────────────────────────
async function fetchAccounts() {
  try {
    const response = await api.get<ListAccountResponse>('/v1/list_account', {
      params: { page_id: 1, page_size: 10 },
    })
    accounts.value = response.data.accounts ?? []
  } catch {
    toast.add({
      severity: 'error',
      summary: 'Unable to load accounts',
      detail: 'Please try again later.',
      life: 4000,
    })
  } finally {
    isLoading.value = false
  }
}

async function sendTransfer() {
  if (!isFormValid.value || !fromAccountId.value) return

  isSending.value = true
  try {
    const response = await api.post<CreateTransferResponse>('/v1/create_transfer', {
      from_account_id: fromAccountId.value,
      to_owner: toOwner.value.trim(),
      to_currency: toCurrency.value,
      amount: toDecimalValue(amount.value),
      currency: toCurrency.value,
    })

    const data = response.data
    const amt = amount.value
    toast.add({
      severity: 'success',
      summary: 'Transfer completed',
      detail: `${currencySymbol.value}${parseFloat(amt).toFixed(2)} sent to ${data.to_account.owner}'s ${data.to_account.currency} account.`,
      life: 5000,
    })

    fromAccountId.value = null
    toOwner.value = ''
    amount.value = ''
    await fetchAccounts()
  } catch (err: unknown) {
    const axiosErr = err as { response?: { status?: number; data?: { message?: string } } }
    let detail = 'Unable to complete the transfer.'
    if (axiosErr.response?.status === 404) {
      detail = `No ${toCurrency.value} account found for "${toOwner.value}".`
    }
    toast.add({
      severity: 'error',
      summary: 'Transfer failed',
      detail,
      life: 5000,
    })
  } finally {
    isSending.value = false
  }
}

onMounted(fetchAccounts)
</script>

<template>
  <Toast />

  <div>
    <h1 class="ledger-heading">Transfer money</h1>
    <p class="text-muted" style="margin-bottom: 2rem;">
      Send money to any account by username.
    </p>

    <!-- No accounts -->
    <Card v-if="!isLoading && accounts.length === 0">
      <template #content>
        <div style="text-align: center; padding: 2rem 0;">
          <i class="pi pi-send" style="font-size: 2.5rem; color: var(--color-gold-reserve); margin-bottom: 1rem;" />
          <p style="margin: 0 0 0.5rem; font-size: 1.1rem;">No accounts available</p>
          <p class="text-muted" style="margin: 0;">
            You need an account before you can send money.
          </p>
        </div>
      </template>
    </Card>

    <!-- Transfer form -->
    <Card v-else>
      <template #content>
        <form class="flex flex-column row-gap-4" @submit.prevent="sendTransfer">
          <!-- From account -->
          <FloatLabel>
            <Select
              id="fromAccount"
              v-model="fromAccountId"
              :options="accountOptions"
              option-label="label"
              option-value="value"
              class="w-full"
              :disabled="isLoading"
            />
            <label for="fromAccount">From account</label>
          </FloatLabel>

          <p
            v-if="fromAccount"
            class="text-muted"
            style="margin: -0.75rem 0 0; font-size: 0.85rem;"
          >
            Available: {{ currencySymbol }}{{ availableBalance }}
          </p>

          <!-- Recipient owner -->
          <FloatLabel>
            <InputText
              id="toOwner"
              v-model="toOwner"
              class="w-full"
              :disabled="!fromAccountId"
              autocomplete="off"
            />
            <label for="toOwner">Recipient username</label>
          </FloatLabel>

          <!-- Recipient currency -->
          <FloatLabel>
            <Select
              id="toCurrency"
              v-model="toCurrency"
              :options="currencies"
              option-label="label"
              option-value="value"
              class="w-full"
              :disabled="!fromAccountId"
            />
            <label for="toCurrency">Currency</label>
          </FloatLabel>

          <p
            v-if="isSelfTransfer"
            style="color: #dc2626; margin: -0.75rem 0 0; font-size: 0.85rem;"
          >
            This is your own {{ toCurrency }} account. Send to someone else.
          </p>

          <!-- Amount -->
          <FloatLabel>
            <InputText
              id="amount"
              v-model="amount"
              type="number"
              step="0.01"
              min="0.01"
              class="w-full"
              :disabled="!fromAccountId"
              :invalid="isInsufficient"
            />
            <label for="amount">Amount</label>
          </FloatLabel>

          <p
            v-if="isInsufficient"
            style="color: #dc2626; margin: -0.75rem 0 0; font-size: 0.85rem;"
          >
            This exceeds your available balance of {{ currencySymbol }}{{ availableBalance }}.
          </p>

          <Button
            type="submit"
            label="Send transfer"
            icon="pi pi-send"
            :loading="isSending"
            :disabled="!isFormValid"
            class="w-full"
          />
        </form>
      </template>
    </Card>
  </div>
</template>
