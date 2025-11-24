CREATE TABLE role_permissions (
    id char(36) PRIMARY KEY,
    role_id char(36) NOT NULL FOREIGN KEY REFERENCES roles(id),
    permission_id char(36) NOT NULL FOREIGN KEY REFERENCES permissions(id),
    created_at datetime default getdate(),
    updated_at datetime NULL
);