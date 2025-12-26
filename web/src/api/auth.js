import request from '@/utils/request'

// 登录
export function login(data) {
  return request.post('/login', data)
}

// 获取当前用户信息
export function getCurrentUser() {
  return request.get('/user/current')
}

// 修改密码
export function changePassword(data) {
  return request.post('/user/change-password', data)
}

// 退出登录
export function logout() {
  return request.post('/logout')
}

// 获取角色列表
export function getRoles() {
  return request.get('/roles')
}
