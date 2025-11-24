CREATE TABLE user_permissions (
    id char(36) PRIMARY KEY,
    user_id char(36) NOT NULL FOREIGN KEY REFERENCES users(id),
    permission_id char(36) NOT NULL FOREIGN KEY REFERENCES permissions(id),
    created_at datetime default getdate(),
    updated_at datetime NULL
);