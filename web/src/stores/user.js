import { defineStore } from 'pinia'
import { login, getCurrentUser, logout } from '@/api/auth'
import router from '@/router'

export const useUserStore = defineStore('user', {
  state: () => ({
    token: localStorage.getItem('token') || '',
    user: JSON.parse(localStorage.getItem('user') || '{}'),
    roles: []
  }),

  getters: {
    isLoggedIn: (state) => !!state.token,
    roleCode: (state) => state.user.role?.code || '',
    userId: (state) => state.user.id || 0,
    
    // 组织角色判断
    isAdmin: (state) => state.user.role?.code === 'admin',
    isDeptManager: (state) => state.user.role?.code === 'dept_manager',
    isTeamLeader: (state) => state.user.role?.code === 'team_leader',
    isTeamMember: (state) => state.user.role?.code === 'team_member',
    
    // 权限判断
    canManageUsers: (state) => state.user.role?.code === 'admin',
    canCreateProject: (state) => ['team_leader', 'team_member'].includes(state.user.role?.code),
    canViewLogs: (state) => state.user.role?.code === 'admin',
    canManageKnowledge: (state) => ['admin', 'dept_manager', 'team_leader', 'team_member'].includes(state.user.role?.code),
    canManageProject: (state) => ['admin', 'dept_manager'].includes(state.user.role?.code),
  },

  actions: {
    async doLogin(username, password) {
      try {
        const res = await login({ username, password })
        this.token = res.data.token
        this.user = res.data.user
        localStorage.setItem('token', res.data.token)
        localStorage.setItem('user', JSON.stringify(res.data.user))
        return res
      } catch (error) {
        throw error
      }
    },

    async fetchCurrentUser() {
      try {
        const res = await getCurrentUser()
        this.user = res.data
        localStorage.setItem('user', JSON.stringify(res.data))
        return res
      } catch (error) {
        throw error
      }
    },

    async doLogout() {
      try {
        await logout()
      } finally {
        this.token = ''
        this.user = {}
        localStorage.removeItem('token')
        localStorage.removeItem('user')
        router.push('/login')
      }
    },

    clearAuth() {
      this.token = ''
      this.user = {}
      localStorage.removeItem('token')
      localStorage.removeItem('user')
    }
  }
})
