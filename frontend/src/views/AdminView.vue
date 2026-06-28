<script setup lang="ts">
import { ref } from 'vue'
import Card from 'primevue/card'
import InputText from 'primevue/inputtext'
import FloatLabel from 'primevue/floatlabel'
import Select from 'primevue/select'
import Button from 'primevue/button'
import Toast from 'primevue/toast'
import { useToast } from 'primevue/usetoast'
import api from '@/api'
import type { UpdateUserResponse } from '@/types/user'

const toast = useToast()

const username = ref('')
const selectedRole = ref<string>('depositor')
const isLoading = ref(false)

const roles = [
  { label: 'Depositor — standard user', value: 'depositor' },
  { label: 'Banker — administrator', value: 'banker' },
]

async function updateRole() {
  if (!username.value) return
  isLoading.value = true
  try {
    const response = await api.patch<UpdateUserResponse>('/v1/update_user', {
      username: username.value,
      role: selectedRole.value,
    })
    toast.add({
      severity: 'success',
      summary: 'Role updated',
      detail: `${response.data.user.full_name} is now a ${selectedRole.value}.`,
      life: 3000,
    })
  } catch (err: unknown) {
    const axiosErr = err as { response?: { status?: number; data?: { message?: string } } }
    const detail =
      axiosErr.response?.status === 404
        ? `User "${username.value}" not found.`
        : axiosErr.response?.data?.message ?? 'Failed to update role.'
    toast.add({
      severity: 'error',
      summary: 'Update failed',
      detail,
      life: 4000,
    })
  } finally {
    isLoading.value = false
  }
}
</script>

<template>
  <Toast />

  <div>
    <h1 class="ledger-heading">Admin</h1>
    <p class="text-muted" style="margin-bottom: 2rem;">
      Manage user roles and permissions.
    </p>

    <Card>
      <template #content>
        <h2 style="font-family: var(--font-display); font-size: 1.15rem; margin: 0 0 1.25rem;">
          Change user role
        </h2>
        <form class="flex flex-column row-gap-4" @submit.prevent="updateRole">
          <FloatLabel>
            <InputText
              id="username"
              v-model="username"
              class="w-full"
              autocomplete="off"
            />
            <label for="username">Username</label>
          </FloatLabel>

          <FloatLabel>
            <Select
              id="role"
              v-model="selectedRole"
              :options="roles"
              option-label="label"
              option-value="value"
              class="w-full"
            />
            <label for="role">Role</label>
          </FloatLabel>

          <Button
            type="submit"
            label="Update role"
            icon="pi pi-check"
            :loading="isLoading"
            :disabled="!username"
          />
        </form>
      </template>
    </Card>

    <Card style="margin-top: 1.5rem;">
      <template #content>
        <h2 style="font-family: var(--font-display); font-size: 1.15rem; margin: 0 0 0.75rem;">
          Role reference
        </h2>
        <div class="flex flex-column row-gap-3">
          <div>
            <p style="margin: 0; font-weight: 600;">Depositor</p>
            <p class="text-muted" style="margin: 0; font-size: 0.85rem;">
              Create accounts, make transfers, manage own profile.
            </p>
          </div>
          <div>
            <p style="margin: 0; font-weight: 600;">Banker</p>
            <p class="text-muted" style="margin: 0; font-size: 0.85rem;">
              Full access — create/update/delete any account, manage user roles.
            </p>
          </div>
        </div>
      </template>
    </Card>
  </div>
</template>
