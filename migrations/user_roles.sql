CREATE TABLE user_roles (
    id char(36) PRIMARY KEY,
    user_id char(36) NOT NULL FOREIGN KEY REFERENCES users(id),
    role_id char(36) NOT NULL FOREIGN KEY REFERENCES roles(id),
    created_at datetime default getdate(),
    updated_at datetime NULL
);