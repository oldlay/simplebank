<script setup lang="ts">
import UserInfo from "@/components/UserInfo.vue";
import LoginUser from "../components/LoginUser.vue";

import Toast from "primevue/toast";
import store from "@/store.ts";
import { useToast } from "primevue/usetoast";
import type { User } from "@/types/user.ts";

const toast = useToast()

const onLogout = (user: User) => {
  store.clearUser()
  toast.add({
    severity: "success",
      summary: `Goodbye, ${user.full_name}`,
      detail: "You have successfully logged out.",
      life: 3000,
  })
}
</script>

<template>
  <main>
    <Toast />
    <h1 class="green">Welcome to Simple Bank!</h1>
    <UserInfo v-if="store.state.user" :user="store.state.user" @logout="onLogout" />
    <LoginUser v-else />

  </main>
</template>

<style>
h1 {
  font-weight: 500;
  font-size: 2.6rem;
}
</style>