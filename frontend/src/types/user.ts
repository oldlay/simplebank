export interface User {
  username: string
  full_name: string
  email: string
  password_changed_at?: string
  created_at?: string
  role?: string // 'depositor' | 'banker'
  is_email_verified?: boolean
}

/** Response from PATCH /v1/update_user */
export interface UpdateUserResponse {
  user: User
}
