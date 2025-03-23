DELETE
FROM public.permissions
WHERE permission_name IN (
        'admin',
        'full',
        'users:full',
        'users:create',
        'users:view',
        'users:update',
        'users:delete',
        'users:assign-role',
        'users:revoke-role',
        'users:view-roles',
        'users:assign-permission',
        'users:revoke-permission',
        'users:view-permissions',
        'roles:full',
        'roles:create',
        'roles:view',
        'roles:delete',
        'permissions:full',
        'permissions:create',
        'permissions:view',
        'permissions:delete'

    );