CREATE TABLE permissions (
    id char(36) PRIMARY KEY,
    name varchar(255) NOT NULL UNIQUE,
    description varchar(255) NULL,
    created_at datetime default getdate(),
    updated_at datetime NULL
);