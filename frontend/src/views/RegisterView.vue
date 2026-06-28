<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import InputText from 'primevue/inputtext'
import FloatLabel from 'primevue/floatlabel'
import Button from 'primevue/button'
import Toast from 'primevue/toast'
import { useToast } from 'primevue/usetoast'
import axios, { AxiosError } from 'axios'

const router = useRouter()
const toast = useToast()

const username = ref('')
const fullName = ref('')
const email = ref('')
const password = ref('')
const isLoading = ref(false)

const isIncomplete = computed(
  () => !username.value || !fullName.value || !email.value || !password.value,
)

const passwordHint = computed(() =>
  password.value.length > 0 && password.value.length < 6
    ? 'Password must be at least 6 characters.'
    : '',
)

async function handleRegister() {
  isLoading.value = true
  try {
    await axios.post('http://localhost:8080/v1/create_user', {
      username: username.value,
      full_name: fullName.value,
      email: email.value,
      password: password.value,
    })
    toast.add({
      severity: 'success',
      summary: 'Account created',
      detail: 'You can now sign in with your new account.',
      life: 5000,
    })
    router.push({ name: 'login' })
  } catch (error) {
    const axiosError = error as AxiosError<{ message: string }>
    const message =
      axiosError.response?.status === 409
        ? 'That username or email is already in use.'
        : 'Something went wrong. Please try again.'
    toast.add({
      severity: 'error',
      summary: 'Registration failed',
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
    <h1 class="ledger-heading">Create your account</h1>
    <p class="text-muted" style="margin-bottom: 2rem;">
      Open a Simple Bank account in under a minute.
    </p>

    <form class="flex flex-column row-gap-4" @submit.prevent="handleRegister">
      <FloatLabel>
        <InputText id="username" v-model="username" class="w-full" autocomplete="username" />
        <label for="username">Username</label>
      </FloatLabel>

      <FloatLabel>
        <InputText id="fullName" v-model="fullName" class="w-full" autocomplete="name" />
        <label for="fullName">Full name</label>
      </FloatLabel>

      <FloatLabel>
        <InputText id="email" v-model="email" type="email" class="w-full" autocomplete="email" />
        <label for="email">Email address</label>
      </FloatLabel>

      <FloatLabel>
        <InputText
          id="password"
          v-model="password"
          type="password"
          class="w-full"
          autocomplete="new-password"
          :invalid="password.length > 0 && password.length < 6"
        />
        <label for="password">Password</label>
      </FloatLabel>
      <small v-if="passwordHint" style="color: #dc2626; margin-top: -0.5rem;">{{ passwordHint }}</small>

      <Button
        type="submit"
        label="Create account"
        :loading="isLoading"
        :disabled="isIncomplete || password.length < 6"
        class="w-full"
      />
    </form>

    <p style="margin-top: 1.5rem; text-align: center;">
      <router-link :to="{ name: 'login' }">Already have an account? Sign in</router-link>
    </p>
  </div>
</template>

<style scoped>
.auth-card {
  margin-top: 6vh;
}
</style>
