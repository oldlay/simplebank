<script setup lang="ts">
import { computed } from 'vue'
import { RouterView, useRoute } from 'vue-router'
import store from '@/store'

const route = useRoute()

const isAuthenticated = computed(() => !!store.state.accessToken)
const isGuestPage = computed(() => route.meta.guest === true)
const isBanker = computed(() => store.state.user?.role === 'banker')

const navItems = computed(() => {
  const items = [
    { label: 'Home', icon: 'pi pi-home', route: { name: 'home' } },
    { label: 'Accounts', icon: 'pi pi-wallet', route: { name: 'accounts' } },
    { label: 'Transfer', icon: 'pi pi-send', route: { name: 'transfer' } },
    { label: 'Profile', icon: 'pi pi-user', route: { name: 'profile' } },
  ]
  if (isBanker.value) {
    items.push({ label: 'Admin', icon: 'pi pi-cog', route: { name: 'admin' } })
  }
  return items
})

type NavItem = { label: string; icon: string; route: { name: string } }

function isActive(item: NavItem) {
  return route.name === item.route.name
}
</script>

<template>
  <div class="app-shell">
    <!-- Navigation bar for authenticated users -->
    <nav v-if="isAuthenticated && !isGuestPage" class="app-nav">
      <div class="nav-inner">
        <router-link :to="{ name: 'home' }" class="nav-brand">
          Simple Bank
        </router-link>

        <div class="nav-links">
          <router-link
            v-for="item in navItems"
            :key="item.route.name"
            :to="item.route"
            class="nav-link"
            :class="{ active: isActive(item) }"
          >
            <i :class="item.icon" style="margin-right: 0.375rem;" />
            {{ item.label }}
          </router-link>
        </div>
      </div>
    </nav>

    <!-- Page content -->
    <main class="app-main">
      <RouterView />
    </main>
  </div>
</template>

<style scoped>
.app-shell {
  min-height: 100vh;
}

/* ── Navigation ─────────────────────────────────── */
.app-nav {
  border-bottom: 1px solid var(--color-statement-border);
  background: #ffffff;
  position: sticky;
  top: 0;
  z-index: 10;
}

.nav-inner {
  max-width: 480px;
  margin: 0 auto;
  padding: 0 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 0;
}

.nav-brand {
  font-family: var(--font-display);
  font-size: 1.15rem;
  color: var(--color-vault-navy);
  text-decoration: none;
  border-bottom: none;
  padding: 0.75rem 0 0.5rem;
}

.nav-brand:hover {
  border-bottom: none;
  color: var(--color-vault-navy);
}

.nav-links {
  display: flex;
  justify-content: space-around;
  padding-bottom: 0.5rem;
}

.nav-link {
  display: inline-flex;
  flex-direction: column;
  align-items: center;
  gap: 0.15rem;
  font-size: 0.85rem;
  font-weight: 500;
  color: var(--color-muted-bronze);
  text-decoration: none;
  border-bottom: none;
  padding: 0.375rem 0;
  border-bottom: 2px solid transparent;
  transition: color 0.15s ease, border-color 0.15s ease;
}

.nav-link:hover {
  color: var(--color-vault-navy);
  border-bottom: 2px solid transparent;
}

.nav-link.active {
  color: var(--color-vault-navy);
  border-bottom-color: var(--color-gold-reserve);
}

/* ── Main content ───────────────────────────────── */
.app-main {
  padding-top: 1.5rem;
}

/* ── Reset #app padding when shell is active ────── */
:deep(#app) {
  padding-top: 0;
}
</style>
