CREATE TABLE password_resets (
    id char(36) PRIMARY KEY,
    user_id char(36) NOT NULL FOREIGN KEY REFERENCES users(id),
    token varchar(255) NOT NULL,
    created_at datetime default getdate(),
    updated_at datetime NULL
);