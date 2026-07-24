// 持久化认证状态 (localStorage 防止刷新丢失)
import { reactive } from 'vue'

function loadAuth() {
  try {
    const saved = localStorage.getItem('auth')
    if (saved) return JSON.parse(saved)
  } catch {}
  return { isLoggedIn: false, username: '' }
}

function saveAuth(state) {
  localStorage.setItem('auth', JSON.stringify({
    isLoggedIn: state.isLoggedIn,
    username: state.username
  }))
}

const initial = loadAuth()

export const auth = reactive({
  isLoggedIn: initial.isLoggedIn,
  username: initial.username
})

export const VALID_CREDENTIALS = {
  username: 'chengdu',
  password: '2024'
}

export function login(username, password) {
  if (username === VALID_CREDENTIALS.username && password === VALID_CREDENTIALS.password) {
    auth.isLoggedIn = true
    auth.username = username
    saveAuth(auth)
    return true
  }
  return false
}

export function logout() {
  auth.isLoggedIn = false
  auth.username = ''
  saveAuth(auth)
}
