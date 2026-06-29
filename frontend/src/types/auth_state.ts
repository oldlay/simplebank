import type { User } from './user'

export interface AuthState {
  user: User | null
  accessToken: string | null
  refreshToken: string | null
}

export interface LoginResponse {
  session_id: string
  access_token: string
  access_token_expires_at: string
  refresh_token: string
  refresh_token_expires_at: string
  user: User
}
