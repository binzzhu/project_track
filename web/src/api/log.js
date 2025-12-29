import request from '@/utils/request'

// 获取操作日志列表
export function getLogs(params) {
  return request.get('/logs', { params })
}

// 获取操作类型列表
export function getLogActions() {
  return request.get('/logs/actions')
}

// 获取模块列表
export function getLogModules() {
  return request.get('/logs/modules')
}

// 获取日志统计
export function getLogStatistics() {
  return request.get('/logs/statistics')
}
