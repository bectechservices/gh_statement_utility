CREATE TABLE user_account_audits (
    id char(36) PRIMARY KEY,
    activity varchar(255) NOT NULL,
    description varchar(255) NOT NULL,
    activity_by char(36) NOT NULL FOREIGN KEY REFERENCES users(id),
    activity_for char(36) NOT NULL FOREIGN KEY REFERENCES users(id),
    created_at datetime default getdate(),
    updated_at datetime default getdate()
);
