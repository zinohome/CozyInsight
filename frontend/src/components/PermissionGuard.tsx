import React from 'react';
import { Navigate } from 'react-router-dom';
import { usePermission } from '../contexts/PermissionContext';

interface PermissionGuardProps {
    children: React.ReactElement;
    resourceType: string;
    resourceId: string;
    action: string;
    fallback?: React.ReactElement;
}

export const PermissionGuard: React.FC<PermissionGuardProps> = ({
    children,
    resourceType,
    resourceId,
    action,
    fallback,
}) => {
    const { hasPermission } = usePermission();

    const allowed = hasPermission(resourceType, resourceId, action);

    if (!allowed) {
        return fallback || <Navigate to="/403" replace />;
    }

    return children;
};

// 按钮级权限控制
interface PermissionButtonProps {
    children: React.ReactElement;
    resourceType: string;
    resourceId: string;
    action: string;
}

export const PermissionButton: React.FC<PermissionButtonProps> = ({
    children,
    resourceType,
    resourceId,
    action,
}) => {
    const { hasPermission } = usePermission();

    const allowed = hasPermission(resourceType, resourceId, action);

    if (!allowed) {
        return null;
    }

    return children;
};
