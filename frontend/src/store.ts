import { reactive, readonly } from 'vue'
import type { AuthState } from './types/auth_state'
import type { User } from './types/user'

const state = reactive<AuthState>({
  user: null,
  accessToken: null,
  refreshToken: null,
})

function setUser(user: User, accessToken: string, refreshToken: string) {
  state.user = user
  state.accessToken = accessToken
  state.refreshToken = refreshToken
  persistAuth()
}

function updateTokens(accessToken: string, refreshToken: string) {
  state.accessToken = accessToken
  state.refreshToken = refreshToken
  persistAuth()
}

function clearUser() {
  state.user = null
  state.accessToken = null
  state.refreshToken = null
  localStorage.removeItem('sb_auth')
}

// Persist session to localStorage so it survives page reloads
function persistAuth() {
  localStorage.setItem(
    'sb_auth',
    JSON.stringify({
      user: state.user,
      accessToken: state.accessToken,
      refreshToken: state.refreshToken,
    }),
  )
}

function loadPersistedAuth() {
  const stored = localStorage.getItem('sb_auth')
  if (stored) {
    try {
      const parsed = JSON.parse(stored)
      state.user = parsed.user
      state.accessToken = parsed.accessToken
      state.refreshToken = parsed.refreshToken
    } catch {
      localStorage.removeItem('sb_auth')
    }
  }
}

export default {
  state: readonly(state),
  setUser,
  updateTokens,
  clearUser,
  loadPersistedAuth,
}
