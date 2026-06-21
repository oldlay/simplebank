<script setup lang="ts">
import InputGroup from "primevue/inputgroup";
import InputGroupAddon from "primevue/inputgroupaddon";
import { InputText } from "primevue";
import FloatLabel from "primevue/floatlabel";
import Button from "primevue/button";
import axios, { AxiosError } from "axios";

import { computed, ref } from "vue";
import type { User } from "@/types/user";
import store from "@/store";
import { useToast } from "primevue/usetoast";

interface LoginResponse {
  user: User;
  access_token: string;
  refresh_token: string;
}

const username = ref<string>("");
const password = ref<string>("");
const isLoginDisabled = computed(() => !username.value || !password.value);

const errorMessage = ref<string>("");

const toast = useToast();

const handleLogin = async () => {
  try {
    const response = await axios.post<LoginResponse>("http://localhost:8080/v1/login_user", {
      username: username.value,
      password: password.value,
    });

    store.setUSer(response.data.user, response.data.access_token, response.data.refresh_token);
    toast.add({
      severity: "success",
      summary: `Hello, ${response.data.user.full_name}`,
      detail: "You have successfully logged in.",
      life: 3000,
    });
  } catch (error) {
    const axiosError = error as AxiosError<{ message: string }>;
    if (axiosError.response && axiosError.response.status === 404) {
      errorMessage.value = axiosError.response.data.message;
    } else {
      errorMessage.value = "An error occured, Please try again later";
    }
    toast.add({
      severity: "error",
      summary: "Login failed",
      detail: errorMessage.value,
      life: 3000,
    });
  }
};
</script>

<template>
  <div class="flex flex-column row-gap-5">
    <InputGroup>
      <InputGroupAddon>
        <i class="pi pi-user"></i>
      </InputGroupAddon>
      <FloatLabel>
        <InputText id="username" v-model="username" />
        <label for="username">Username</label>
      </FloatLabel>
    </InputGroup>

    <InputGroup>
      <InputGroupAddon>
        <i class="pi pi-lock"></i>
      </InputGroupAddon>
      <FloatLabel>
        <InputText type="password" id="Password" v-model="password" />
        <label for="password">Password</label>
      </FloatLabel>
    </InputGroup>

    <Button label="Login" :disabled="isLoginDisabled" @click="handleLogin" />
  </div>
</template>
