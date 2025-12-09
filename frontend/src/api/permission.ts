import request from '../utils/request';

export interface Role {
    id: string;
    name: string;
    description: string;
    type: string;
    createTime: number;
    updateTime: number;
    createBy: string;
}

export interface Permission {
    id: string;
    name: string;
    resource: string;
    resourceId: string;
    action: string;
    description: string;
    createTime: number;
}

export interface ResourcePermission {
    id: string;
    resourceType: string;
    resourceId: string;
    targetType: string;
    targetId: string;
    permission: string;
    createTime: number;
    createBy: string;
}

export const roleAPI = {
    // 角色管理
    list: () => request.get<Role[]>('/role'),
    get: (id: string) => request.get<Role>(`/role/${id}`),
    create: (data: Partial<Role>) => request.post<Role>('/role', data),
    update: (id: string, data: Partial<Role>) => request.put<Role>(`/role/${id}`, data),
    delete: (id: string) => request.delete(`/role/${id}`),

    // 用户角色
    assignToUser: (roleId: string, userId: string) =>
        request.post(`/role/${roleId}/assign`, { userId }),
    removeFromUser: (roleId: string, userId: string) =>
        request.delete(`/role/${roleId}/remove?userId=${userId}`),
    getUserRoles: (userId: string) =>
        request.get<Role[]>(`/role/user?userId=${userId}`),
};

export const permissionAPI = {
    // 权限管理
    list: () => request.get<Permission[]>('/permission'),
    getRolePermissions: (roleId: string) =>
        request.get<Permission[]>(`/permission/role/${roleId}`),
    grantToRole: (roleId: string, permissionIds: string[]) =>
        request.post(`/permission/role/${roleId}/grant`, { permissionIds }),
    revokeFromRole: (roleId: string, permissionIds: string[]) =>
        request.post(`/permission/role/${roleId}/revoke`, { permissionIds }),

    // 资源权限
    grantResourcePermission: (data: {
        resourceType: string;
        resourceId: string;
        targetType: string;
        targetId: string;
        permission: string;
    }) => request.post('/permission/resource/grant', data),

    getResourcePermissions: (resourceType: string, resourceId: string) =>
        request.get<ResourcePermission[]>(
            `/permission/resource?resourceType=${resourceType}&resourceId=${resourceId}`
        ),

    checkPermission: (resourceType: string, resourceId: string, action: string) =>
        request.get<{ hasPermission: boolean }>(
            `/permission/check?resourceType=${resourceType}&resourceId=${resourceId}&action=${action}`
        ),
};
