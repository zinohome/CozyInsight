export interface OperationLog {
  id: number
  userId: number
  username: string
  method: string
  path: string
  query: string
  body: string
  ip: string
  userAgent: string
  statusCode: number
  duration: number
  errorMessage: string
  createdAt: string
}
