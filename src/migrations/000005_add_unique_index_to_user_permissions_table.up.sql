ALTER TABLE user_permissions
    ADD CONSTRAINT user_permissions_user_id_permission_id_unique
        UNIQUE (user_id, permission_id);