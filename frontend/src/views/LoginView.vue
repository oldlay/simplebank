<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import InputText from 'primevue/inputtext'
import FloatLabel from 'primevue/floatlabel'
import Button from 'primevue/button'
import Toast from 'primevue/toast'
import { useToast } from 'primevue/usetoast'
import axios, { AxiosError } from 'axios'
import store from '@/store'
import type { LoginResponse } from '@/types/auth_state'

const router = useRouter()
const toast = useToast()

const username = ref('')
const password = ref('')
const isLoading = ref(false)
const isEmpty = computed(() => !username.value || !password.value)

async function handleLogin() {
  isLoading.value = true
  try {
    const response = await axios.post<LoginResponse>('http://localhost:8080/v1/login_user', {
      username: username.value,
      password: password.value,
    })
    const data = response.data
    store.setUser(data.user, data.access_token, data.refresh_token)
    toast.add({
      severity: 'success',
      summary: `Welcome back, ${data.user.full_name}`,
      detail: 'You have successfully signed in.',
      life: 3000,
    })
    router.push({ name: 'home' })
  } catch (error) {
    const axiosError = error as AxiosError<{ message: string }>
    const message =
      axiosError.response?.status === 404
        ? axiosError.response.data.message
        : 'Unable to sign in. Please check your credentials and try again.'
    toast.add({
      severity: 'error',
      summary: 'Sign in failed',
      detail: message,
      life: 4000,
    })
  } finally {
    isLoading.value = false
  }
}
</script>

<template>
  <Toast />
  <div class="auth-card">
    <h1 class="ledger-heading">Sign in</h1>
    <p class="text-muted" style="margin-bottom: 2rem;">
      Enter your credentials to access your accounts.
    </p>

    <form class="flex flex-column row-gap-4" @submit.prevent="handleLogin">
      <FloatLabel>
        <InputText id="username" v-model="username" class="w-full" autocomplete="username" />
        <label for="username">Username</label>
      </FloatLabel>

      <FloatLabel>
        <InputText
          id="password"
          v-model="password"
          type="password"
          class="w-full"
          autocomplete="current-password"
        />
        <label for="password">Password</label>
      </FloatLabel>

      <Button
        type="submit"
        label="Sign in"
        :loading="isLoading"
        :disabled="isEmpty"
        class="w-full"
      />
    </form>

    <p style="margin-top: 1.5rem; text-align: center;">
      <router-link :to="{ name: 'register' }">New to Simple Bank? Create an account</router-link>
    </p>
  </div>
</template>

<style scoped>
.auth-card {
  margin-top: 10vh;
}
</style>
