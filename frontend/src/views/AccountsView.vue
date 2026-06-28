<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import Card from 'primevue/card'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import Select from 'primevue/select'
import InputText from 'primevue/inputtext'
import FloatLabel from 'primevue/floatlabel'
import ConfirmDialog from 'primevue/confirmdialog'
import { useConfirm } from 'primevue/useconfirm'
import Toast from 'primevue/toast'
import { useToast } from 'primevue/usetoast'
import api from '@/api'
import store from '@/store'
import type { Account, ListAccountResponse, CreateAccountResponse, UpdateAccountResponse } from '@/types/account'
import { getAmount, toDecimalValue } from '@/types/account'

const toast = useToast()
const confirm = useConfirm()

const accounts = ref<Account[]>([])
const isLoading = ref(true)

// ── Role ────────────────────────────────────────────
const isBanker = computed(() => store.state.user?.role === 'banker')

// ── Create dialog ───────────────────────────────────
const showCreateDialog = ref(false)
const newCurrency = ref('USD')
const isCreating = ref(false)

const currencies = [
  { label: 'USD — US Dollar', value: 'USD' },
  { label: 'EUR — Euro', value: 'EUR' },
  { label: 'CAD — Canadian Dollar', value: 'CAD' },
]

// ── Edit dialog (admin) ─────────────────────────────
const showEditDialog = ref(false)
const editingAccount = ref<Account | null>(null)
const editBalance = ref('')
const editCurrency = ref('USD')
const isUpdating = ref(false)

const hasAccounts = computed(() => accounts.value.length > 0)

function currencySymbol(currency: string): string {
  const map: Record<string, string> = { USD: '$', EUR: '€', CAD: 'CA$' }
  return map[currency] || currency
}

function formatBalance(raw: string): string {
  const n = parseFloat(raw)
  return n.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

// ── List ────────────────────────────────────────────
async function fetchAccounts() {
  isLoading.value = true
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

// ── Create ──────────────────────────────────────────
async function createAccount() {
  isCreating.value = true
  try {
    const response = await api.post<CreateAccountResponse>('/v1/create_account', {
      currency: newCurrency.value,
    })
    accounts.value.push(response.data.account)
    showCreateDialog.value = false
    toast.add({
      severity: 'success',
      summary: 'Account opened',
      detail: `Your ${newCurrency.value} account is ready.`,
      life: 3000,
    })
  } catch (err: unknown) {
    const axiosErr = err as { response?: { status?: number; data?: { message?: string } } }
    const detail =
      axiosErr.response?.status === 409
        ? 'You already have an account in this currency.'
        : axiosErr.response?.data?.message ?? 'Could not open account.'
    toast.add({
      severity: 'error',
      summary: 'Failed',
      detail,
      life: 4000,
    })
  } finally {
    isCreating.value = false
  }
}

// ── Edit (admin) ────────────────────────────────────
function openEditDialog(acct: Account) {
  editingAccount.value = acct
  editBalance.value = parseFloat(getAmount(acct)).toFixed(2)
  editCurrency.value = acct.currency
  showEditDialog.value = true
}

async function updateAccount() {
  if (!editingAccount.value) return
  isUpdating.value = true
  try {
    const response = await api.patch<UpdateAccountResponse>(
      `/v1/update_account/${editingAccount.value.id}`,
      {
        balance: toDecimalValue(editBalance.value),
        currency: editCurrency.value,
      },
    )
    // Replace the updated account in the list
    const idx = accounts.value.findIndex((a) => a.id === editingAccount.value!.id)
    if (idx !== -1) accounts.value[idx] = response.data.account
    showEditDialog.value = false
    toast.add({
      severity: 'success',
      summary: 'Account updated',
      detail: `Account ····${String(editingAccount.value.id).slice(-4)} has been updated.`,
      life: 3000,
    })
  } catch (err: unknown) {
    const axiosErr = err as { response?: { status?: number; data?: { message?: string } } }
    toast.add({
      severity: 'error',
      summary: 'Update failed',
      detail: axiosErr.response?.data?.message ?? 'Could not update account.',
      life: 4000,
    })
  } finally {
    isUpdating.value = false
  }
}

// ── Delete (admin) ──────────────────────────────────
function confirmDelete(acct: Account) {
  confirm.require({
    message: `Delete ${acct.currency} account ····${String(acct.id).slice(-4)}?`,
    header: 'Delete account',
    icon: 'pi pi-exclamation-triangle',
    rejectLabel: 'Cancel',
    rejectProps: { severity: 'secondary', variant: 'outlined', label: 'Cancel' },
    acceptLabel: 'Delete',
    acceptProps: { severity: 'danger', label: 'Delete' },
    accept: () => deleteAccount(acct),
  })
}

async function deleteAccount(acct: Account) {
  try {
    await api.delete(`/v1/delete_account/${acct.id}`)
    accounts.value = accounts.value.filter((a) => a.id !== acct.id)
    toast.add({
      severity: 'success',
      summary: 'Account deleted',
      detail: `${acct.currency} account ····${String(acct.id).slice(-4)} has been removed.`,
      life: 3000,
    })
  } catch (err: unknown) {
    const axiosErr = err as { response?: { status?: number; data?: { message?: string } } }
    toast.add({
      severity: 'error',
      summary: 'Delete failed',
      detail: axiosErr.response?.data?.message ?? 'Could not delete account.',
      life: 4000,
    })
  }
}

onMounted(fetchAccounts)
</script>

<template>
  <Toast />
  <ConfirmDialog />

  <div>
    <header class="flex flex-row align-items-center justify-content-between" style="margin-bottom: 1.5rem;">
      <div class="flex flex-row align-items-center" style="gap: 0.75rem;">
        <h1 class="ledger-heading" style="margin-bottom: 0;">Accounts</h1>
        <span v-if="isBanker" class="admin-badge">admin</span>
      </div>
      <Button label="Open new" icon="pi pi-plus" @click="showCreateDialog = true" />
    </header>

    <!-- Loading skeleton -->
    <div v-if="isLoading" class="flex flex-column row-gap-3">
      <Card v-for="i in 3" :key="i">
        <template #content><div style="height: 3.5rem;" /></template>
      </Card>
    </div>

    <!-- Empty state -->
    <Card v-else-if="!hasAccounts">
      <template #content>
        <div style="text-align: center; padding: 2rem 0;">
          <i class="pi pi-wallet" style="font-size: 2.5rem; color: var(--color-gold-reserve); margin-bottom: 1rem;" />
          <p style="margin: 0 0 0.5rem; font-size: 1.1rem;">No accounts yet</p>
          <p class="text-muted" style="margin: 0 0 1.25rem;">
            Open your first account — it takes seconds.
          </p>
          <Button label="Open an account" icon="pi pi-plus" @click="showCreateDialog = true" />
        </div>
      </template>
    </Card>

    <!-- Account list -->
    <div v-else class="flex flex-column row-gap-3">
      <Card v-for="acct in accounts" :key="acct.id">
        <template #content>
          <div class="flex flex-row align-items-center justify-content-between">
            <div>
              <div class="flex flex-row align-items-center" style="gap: 0.5rem; margin-bottom: 0.25rem;">
                <span style="font-weight: 600;">{{ acct.currency }}</span>
                <span class="text-muted" style="font-size: 0.8rem;">
                  ····{{ String(acct.id).slice(-4) }}
                </span>
              </div>
              <p class="text-mono" style="margin: 0; font-size: 1.35rem; font-weight: 500;">
                {{ currencySymbol(acct.currency) }}{{ formatBalance(getAmount(acct)) }}
              </p>
              <p class="text-muted" style="margin: 0.25rem 0 0; font-size: 0.75rem;">
                {{ acct.owner }}
              </p>
            </div>

            <!-- Admin actions -->
            <div v-if="isBanker" class="flex flex-row" style="gap: 0.5rem;">
              <Button
                icon="pi pi-pencil"
                severity="secondary"
                variant="outlined"
                aria-label="Edit account"
                @click="openEditDialog(acct)"
              />
              <Button
                icon="pi pi-trash"
                severity="danger"
                variant="outlined"
                aria-label="Delete account"
                @click="confirmDelete(acct)"
              />
            </div>
          </div>
        </template>
      </Card>
    </div>

    <!-- Create dialog -->
    <Dialog
      v-model:visible="showCreateDialog"
      header="Open a new account"
      :modal="true"
      :style="{ width: '380px' }"
    >
      <div class="flex flex-column row-gap-4" style="padding-top: 0.5rem;">
        <FloatLabel>
          <Select
            id="currency"
            v-model="newCurrency"
            :options="currencies"
            option-label="label"
            option-value="value"
            class="w-full"
          />
          <label for="currency">Currency</label>
        </FloatLabel>
        <div class="flex flex-row justify-content-end" style="gap: 0.75rem;">
          <Button label="Cancel" severity="secondary" variant="outlined" @click="showCreateDialog = false" />
          <Button label="Open account" :loading="isCreating" @click="createAccount" />
        </div>
      </div>
    </Dialog>

    <!-- Edit dialog (admin) -->
    <Dialog
      v-model:visible="showEditDialog"
      header="Edit account"
      :modal="true"
      :style="{ width: '380px' }"
    >
      <div v-if="editingAccount" class="flex flex-column row-gap-4" style="padding-top: 0.5rem;">
        <div>
          <p class="text-muted" style="margin: 0 0 0.25rem; font-size: 0.8rem; text-transform: uppercase; letter-spacing: 0.05em;">
            Account
          </p>
          <p style="margin: 0;">
            {{ editingAccount.currency }} ····{{ String(editingAccount.id).slice(-4) }}
          </p>
        </div>

        <FloatLabel>
          <InputText
            id="editBalance"
            v-model="editBalance"
            type="number"
            step="0.01"
            class="w-full"
          />
          <label for="editBalance">Balance</label>
        </FloatLabel>

        <FloatLabel>
          <Select
            id="editCurrency"
            v-model="editCurrency"
            :options="currencies"
            option-label="label"
            option-value="value"
            class="w-full"
          />
          <label for="editCurrency">Currency</label>
        </FloatLabel>

        <div class="flex flex-row justify-content-end" style="gap: 0.75rem;">
          <Button label="Cancel" severity="secondary" variant="outlined" @click="showEditDialog = false" />
          <Button label="Save changes" :loading="isUpdating" @click="updateAccount" />
        </div>
      </div>
    </Dialog>
  </div>
</template>

<style scoped>
.admin-badge {
  font-size: 0.7rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: #ffffff;
  background: var(--color-gold-reserve);
  padding: 0.15rem 0.5rem;
  border-radius: 3px;
}
</style>
