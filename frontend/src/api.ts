import axios from 'axios'
import store from '@/store'

const api = axios.create({
  baseURL: 'http://localhost:8080',
  headers: {
    'Content-Type': 'application/json',
  },
})

// Attach access token to every authenticated request
api.interceptors.request.use((config) => {
  if (store.state.accessToken) {
    config.headers.Authorization = `Bearer ${store.state.accessToken}`
  }
  return config
})

// On 401, clear auth and redirect to login
// (Token refresh via /tokens/renew_access not yet available in gRPC-Gateway)
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      store.clearUser()
    }
    return Promise.reject(error)
  },
)

export default api
