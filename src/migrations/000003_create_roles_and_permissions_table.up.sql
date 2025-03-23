CREATE TABLE IF NOT EXISTS permissions
(
    id              SERIAL PRIMARY KEY,
    permission_name VARCHAR(255) NOT NULL UNIQUE,
    description     TEXT
);

CREATE TABLE IF NOT EXISTS roles
(
    id          SERIAL PRIMARY KEY,
    role_name   VARCHAR(255) NOT NULL UNIQUE,
    description TEXT
);

CREATE TABLE IF NOT EXISTS role_permissions
(
    role_id       INTEGER REFERENCES roles (id) ON DELETE CASCADE,
    permission_id INTEGER REFERENCES permissions (id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

CREATE TABLE IF NOT EXISTS user_permissions
(
    user_id       UUID REFERENCES users (id) ON DELETE CASCADE,
    permission_id INTEGER REFERENCES permissions (id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, permission_id)
);

CREATE TABLE IF NOT EXISTS user_roles
(
    user_id UUID REFERENCES users (id) ON DELETE CASCADE,
    role_id INTEGER REFERENCES roles (id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_id)
);

CREATE INDEX IF NOT EXISTS idx_role_permissions_role_id ON role_permissions (role_id);
CREATE INDEX IF NOT EXISTS idx_role_permissions_permission_id ON role_permissions (permission_id);
CREATE INDEX IF NOT EXISTS idx_user_permissions_user_id ON user_permissions (user_id);
CREATE INDEX IF NOT EXISTS idx_user_permissions_permission_id ON user_permissions (permission_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_user_id ON user_roles (user_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_role_id ON user_roles (role_id);