import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { permissionAPI } from '../api/permission';

interface PermissionContextType {
    checkPermission: (resourceType: string, resourceId: string, action: string) => Promise<boolean>;
    hasPermission: (resourceType: string, resourceId: string, action: string) => boolean;
    permissions: Map<string, boolean>;
}

const PermissionContext = createContext<PermissionContextType | undefined>(undefined);

export const PermissionProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
    const [permissions, setPermissions] = useState<Map<string, boolean>>(new Map());

    const checkPermission = async (resourceType: string, resourceId: string, action: string): Promise<boolean> => {
        const key = `${resourceType}:${resourceId}:${action}`;

        // 如果已缓存,直接返回
        if (permissions.has(key)) {
            return permissions.get(key)!;
        }

        try {
            const result = await permissionAPI.checkPermission(resourceType, resourceId, action);
            const hasPermission = result.hasPermission;

            // 缓存结果
            setPermissions(prev => new Map(prev).set(key, hasPermission));

            return hasPermission;
        } catch (error) {
            console.error('Permission check failed:', error);
            return false;
        }
    };

    const hasPermission = (resourceType: string, resourceId: string, action: string): boolean => {
        const key = `${resourceType}:${resourceId}:${action}`;
        return permissions.get(key) || false;
    };

    return (
        <PermissionContext.Provider value={{ checkPermission, hasPermission, permissions }}>
            {children}
        </PermissionContext.Provider>
    );
};

export const usePermission = () => {
    const context = useContext(PermissionContext);
    if (!context) {
        throw new Error('usePermission must be used within PermissionProvider');
    }
    return context;
};

// 权限守卫HOC
export function withPermission<P extends object>(
    Component: React.ComponentType<P>,
    resourceType: string,
    resourceId: string,
    action: string
) {
    return (props: P) => {
        const { checkPermission } = usePermission();
        const [hasAccess, setHasAccess] = useState<boolean | null>(null);

        useEffect(() => {
            checkPermission(resourceType, resourceId, action).then(setHasAccess);
        }, [resourceType, resourceId, action]);

        if (hasAccess === null) {
            return <div>Loading...</div>;
        }

        if (!hasAccess) {
            return <div>无权限访问</div>;
        }

        return <Component {...props} />;
    };
}
