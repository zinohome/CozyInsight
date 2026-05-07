import request from './request'
import type { OperationLog } from '@/types/log'

export const logAPI = {
  list: (limit?: number) => request.get<OperationLog[]>('/operation-log', { params: { limit } }),
}
