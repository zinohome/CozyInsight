import request from './request'
import type { Message } from '@/types/message'

export const messageAPI = {
  list: (unreadOnly?: boolean) =>
    request.get<Message[]>(`/messages${unreadOnly ? '?unreadOnly=true' : ''}`),
  countUnread: () => request.get<number>('/messages/unread-count'),
  markAsRead: (id: number) => request.post(`/messages/${id}/read`),
  markAllAsRead: () => request.post('/messages/read-all'),
  remove: (id: number) => request.delete(`/messages/${id}`),
}
