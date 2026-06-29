<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import Card from 'primevue/card'
import Button from 'primevue/button'
import Toast from 'primevue/toast'
import { useToast } from 'primevue/usetoast'
import api from '@/api'
import store from '@/store'
import type { Account, ListAccountResponse } from '@/types/account'
import { getAmount } from '@/types/account'
import type { User } from '@/types/user'

interface ExchangeRates {
  result: string
  date: string
  rates: Record<string, number> // e.g. { "USD": 1, "EUR": 0.92, "CAD": 1.35 }
}

const router = useRouter()
const toast = useToast()
const user = store.state.user as User

const accounts = ref<Account[]>([])
const rates = ref<Record<string, number>>({ USD: 1 })
const ratesDate = ref('')
const isLoading = ref(true)

const hasAccounts = computed(() => accounts.value.length > 0)

// Convert each account's balance to USD, then sum
const totalUsd = computed(() => {
  const sum = accounts.value.reduce((total, a) => {
    const balance = parseFloat(getAmount(a))
    const rate = rates.value[a.currency] ?? 1
    return total + balance / rate
  }, 0)
  return sum.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
})

function currencyPrefix(currency: string): string {
  const map: Record<string, string> = { USD: '$', EUR: '€', CAD: 'CA$' }
  return map[currency] || currency
}

function formatBalance(raw: string): string {
  const n = parseFloat(raw)
  return n.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

async function fetchData() {
  try {
    const [acctRes, rateRes] = await Promise.all([
      api.get<ListAccountResponse>('/v1/list_account', { params: { page_id: 1, page_size: 10 } }),
      api.get<ExchangeRates>('/v1/exchange_rate'),
    ])
    accounts.value = acctRes.data.accounts ?? []
    rates.value = rateRes.data.rates
    ratesDate.value = rateRes.data.date
  } catch {
    // Silent — accounts may be empty, rates may fail temporarily
  } finally {
    isLoading.value = false
  }
}

function onLogout() {
  store.clearUser()
  toast.add({
    severity: 'success',
    summary: `Goodbye, ${user.full_name}`,
    detail: 'You have signed out.',
    life: 3000,
  })
  router.push({ name: 'login' })
}

onMounted(fetchData)
</script>

<template>
  <Toast />

  <div>
    <header class="flex flex-row align-items-center justify-content-between" style="margin-bottom: 2rem;">
      <div>
        <h1 class="ledger-heading" style="margin-bottom: 0.25rem;">Simple Bank</h1>
        <p class="text-muted" style="margin: 0;">
          Welcome, {{ user?.full_name ?? '—' }}
        </p>
      </div>
      <Button
        label="Sign out"
        icon="pi pi-sign-out"
        severity="secondary"
        variant="outlined"
        @click="onLogout"
      />
    </header>

    <!-- Total balance (USD) -->
    <Card v-if="hasAccounts" style="margin-bottom: 1.5rem;">
      <template #content>
        <p class="text-muted" style="margin: 0 0 0.25rem; font-size: 0.8rem; text-transform: uppercase; letter-spacing: 0.05em;">
          Total balance
        </p>
        <p class="text-mono" style="margin: 0; font-size: 2rem; font-weight: 500;">
          ${{ totalUsd }}
        </p>
        <p class="text-muted" style="margin: 0.25rem 0 0; font-size: 0.7rem;">
          USD equivalent · rates as of {{ ratesDate || '—' }}
        </p>
      </template>
    </Card>

    <!-- Loading -->
    <div v-if="isLoading" class="flex flex-column row-gap-3">
      <Card v-for="i in 2" :key="i">
        <template #content>
          <div style="height: 3rem;" />
        </template>
      </Card>
    </div>

    <!-- Account cards -->
    <div v-else-if="hasAccounts" class="flex flex-column row-gap-3">
      <Card v-for="acct in accounts" :key="acct.id">
        <template #content>
          <div class="flex flex-row align-items-center justify-content-between">
            <div>
              <div class="flex flex-row align-items-center" style="gap: 0.5rem; margin-bottom: 0.25rem;">
                <span style="font-weight: 600;">{{ acct.currency }}</span>
                <span class="text-muted" style="font-size: 0.8rem;">{{ acct.owner }}</span>
              </div>
              <p class="text-mono" style="margin: 0; font-size: 1.35rem; font-weight: 500;">
                {{ currencyPrefix(acct.currency) }}{{ formatBalance(getAmount(acct)) }}
              </p>
            </div>
            <Button
              icon="pi pi-chevron-right"
              severity="secondary"
              variant="outlined"
              aria-label="Go to accounts"
              @click="router.push({ name: 'accounts' })"
            />
          </div>
        </template>
      </Card>

      <!-- Quick actions -->
      <div class="flex flex-row" style="gap: 0.75rem; margin-top: 0.25rem;">
        <Button
          label="New transfer"
          icon="pi pi-send"
          severity="secondary"
          variant="outlined"
          class="flex-1"
          @click="router.push({ name: 'transfer' })"
        />
        <Button
          label="New account"
          icon="pi pi-plus"
          severity="secondary"
          variant="outlined"
          class="flex-1"
          @click="router.push({ name: 'accounts' })"
        />
      </div>
    </div>

    <!-- No accounts yet -->
    <div v-else class="flex flex-column row-gap-3">
      <Card class="nav-card" @click="router.push({ name: 'accounts' })">
        <template #content>
          <div class="flex flex-row align-items-center" style="gap: 1rem;">
            <div class="nav-icon">
              <i class="pi pi-wallet" style="font-size: 1.5rem;" />
            </div>
            <div style="flex: 1;">
              <p style="margin: 0; font-weight: 600;">Open an account</p>
              <p class="text-muted" style="margin: 0; font-size: 0.85rem;">
                Start by creating your first account
              </p>
            </div>
            <i class="pi pi-chevron-right text-muted" />
          </div>
        </template>
      </Card>

      <Card class="nav-card" @click="router.push({ name: 'profile' })">
        <template #content>
          <div class="flex flex-row align-items-center" style="gap: 1rem;">
            <div class="nav-icon">
              <i class="pi pi-user" style="font-size: 1.5rem;" />
            </div>
            <div style="flex: 1;">
              <p style="margin: 0; font-weight: 600;">Profile</p>
              <p class="text-muted" style="margin: 0; font-size: 0.85rem;">
                Manage your account settings
              </p>
            </div>
            <i class="pi pi-chevron-right text-muted" />
          </div>
        </template>
      </Card>
    </div>
  </div>
</template>

<style scoped>
.nav-card {
  cursor: pointer;
  transition: border-color 0.15s ease;
}

.nav-card:hover {
  border-color: var(--color-gold-reserve);
}

.nav-icon {
  width: 2.5rem;
  height: 2.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-vault-navy);
  background: var(--color-ledger-white);
  border-radius: var(--radius-card);
}
</style>
