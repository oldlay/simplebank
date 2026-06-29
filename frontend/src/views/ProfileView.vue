<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import Card from 'primevue/card'
import Divider from 'primevue/divider'
import Button from 'primevue/button'
import Toast from 'primevue/toast'
import { useToast } from 'primevue/usetoast'
import store from '@/store'
import type { User } from '@/types/user'

const router = useRouter()
const toast = useToast()

const user = store.state.user as User

function handleLogout() {
  store.clearUser()
  toast.add({
    severity: 'success',
    summary: `Goodbye, ${user.full_name}`,
    detail: 'You have signed out.',
    life: 3000,
  })
  router.push({ name: 'login' })
}

function formatDate(dateStr: string | undefined) {
  if (!dateStr) return '—'
  return new Date(dateStr).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  })
}
</script>

<template>
  <Toast />

  <div>
    <h1 class="ledger-heading">Profile</h1>
    <p class="text-muted" style="margin-bottom: 2rem;">
      Your account details and settings.
    </p>

    <Card>
      <template #content>
        <div class="flex flex-column row-gap-3">
          <div>
            <p class="text-muted" style="margin: 0; font-size: 0.8rem; text-transform: uppercase; letter-spacing: 0.05em;">
              Full name
            </p>
            <p style="margin: 0.125rem 0 0; font-size: 1.1rem;">{{ user.full_name }}</p>
          </div>

          <div>
            <p class="text-muted" style="margin: 0; font-size: 0.8rem; text-transform: uppercase; letter-spacing: 0.05em;">
              Username
            </p>
            <p class="text-mono" style="margin: 0.125rem 0 0;">{{ user.username }}</p>
          </div>

          <div>
            <p class="text-muted" style="margin: 0; font-size: 0.8rem; text-transform: uppercase; letter-spacing: 0.05em;">
              Email
            </p>
            <p style="margin: 0.125rem 0 0;">
              {{ user.email }}
              <span
                v-if="user.is_email_verified"
                style="color: #16a34a; font-size: 0.8rem; margin-left: 0.5rem;"
              >
                Verified
              </span>
              <span
                v-else
                style="color: #8b7355; font-size: 0.8rem; margin-left: 0.5rem;"
              >
                Unverified
              </span>
            </p>
          </div>

          <div>
            <p class="text-muted" style="margin: 0; font-size: 0.8rem; text-transform: uppercase; letter-spacing: 0.05em;">
              Role
            </p>
            <p style="margin: 0.125rem 0 0; text-transform: capitalize;">
              {{ user.role ?? 'depositor' }}
            </p>
          </div>

          <div>
            <p class="text-muted" style="margin: 0; font-size: 0.8rem; text-transform: uppercase; letter-spacing: 0.05em;">
              Member since
            </p>
            <p style="margin: 0.125rem 0 0;">
              {{ formatDate(user.created_at) }}
            </p>
          </div>
        </div>

        <Divider />

        <Button
          label="Sign out"
          icon="pi pi-sign-out"
          severity="danger"
          variant="outlined"
          @click="handleLogout"
        />
      </template>
    </Card>
  </div>
</template>
