CREATE TABLE user_activity_audits (
    id char(36) PRIMARY KEY,
    user_id char(36) NOT NULL FOREIGN KEY REFERENCES users(id),
    branch_id char(36) NOT NULL FOREIGN KEY REFERENCES branches(id),
    activity varchar(255) NOT NULL,
    created_at datetime default getdate()
);