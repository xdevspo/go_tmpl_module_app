DELETE
FROM public.roles
WHERE role_name IN (
                    'admin',
                    'sales',
                    'customer',
                    'contractor'
    );