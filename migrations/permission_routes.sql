CREATE TABLE permission_routes (
    id char(36) PRIMARY KEY,
    permission_id char(36) NOT NULL FOREIGN KEY REFERENCES permissions(id),
    path varchar(255) NOT NULL,
    alias varchar(255) NOT NULL,
    created_at datetime default getdate(),
    updated_at datetime NULL
);