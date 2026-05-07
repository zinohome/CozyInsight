export interface Role {
  id: number
  name: string
  code: string
  description: string
  status: number
  createdAt: string
}

export interface Menu {
  id: number
  parentId: number
  name: string
  path: string
  component: string
  icon: string
  sort: number
  status: number
}

export interface CreateRoleRequest {
  name: string
  code: string
  description?: string
}
