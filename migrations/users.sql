CREATE TABLE users (
    id char(36) PRIMARY KEY,
    ab_number varchar(255) NOT NULL,
    first_name varchar(255) NOT NULL,
    last_name varchar(255) NOT NULL,
    email varchar(255) NOT NULL,
    password varchar(255) NOT NULL,
    branch_id char(36) NOT NULL FOREIGN KEY REFERENCES branches(id),
    password_last_changed datetime NULL,
    must_reset_password bit default 1,
    locked bit default 0,
    deleted datetime NULL,
    last_login datetime NULL,
    created_at datetime default getdate(),
    updated_at datetime NULL
);

