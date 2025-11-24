CREATE TABLE records (
    id char(36) PRIMARY KEY,
    client_info_id char(36) NULL FOREIGN KEY REFERENCES client_info(id),
    user_id char(36) NOT NULL FOREIGN KEY REFERENCES users(id),
    national_id varchar(255) NOT NULL,
    success tinyint not null,
    created_at datetime default getdate()
);