CREATE TABLE branches (
    id char(36) PRIMARY KEY,
    name varchar(255) NOT NULL,
    code varchar(255) NOT NULL,
    created_at datetime default getdate(),
    updated_at datetime default getdate(),
);