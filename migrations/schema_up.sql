-- Adding Application Upgrade Version tables in mssql database
-- Input the Database you want to add the below
---USE NG_DB;
GO

/*****  branches Table **************/
CREATE TABLE branches (
    id char(36) PRIMARY KEY,
    name varchar(255) NOT NULL,
    code varchar(255) NOT NULL,
    created_at datetime default getdate(),
    updated_at datetime default getdate(),
);
/*****  password_expiries Table **************/
CREATE TABLE password_expiries (
    id char(36) PRIMARY KEY,
    days integer NOT NULL,
    remind_in integer NOT NULL,
    length integer NOT NULL,
    created_at datetime default getdate(),
    updated_at datetime NULL
);
--- User table
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
/*****  user_activity_audits Table **************/
CREATE TABLE user_activity_audits (
    id char(36) PRIMARY KEY,
    user_id char(36) NOT NULL FOREIGN KEY REFERENCES users(id),
    branch_id char(36) NOT NULL FOREIGN KEY REFERENCES branches(id),
    activity varchar(255) NOT NULL,
    created_at datetime default getdate()
);
/*****  user_account_audits Table **************/
CREATE TABLE user_account_audits (
    id char(36) PRIMARY KEY,
    activity varchar(255) NOT NULL,
    description varchar(255) NOT NULL,
    activity_by char(36) NOT NULL FOREIGN KEY REFERENCES users(id),
    activity_for char(36) NOT NULL FOREIGN KEY REFERENCES users(id),
    created_at datetime default getdate(),
    updated_at datetime default getdate()
);


--- password_resets table
CREATE TABLE password_resets (
    id char(36) PRIMARY KEY,
    user_id char(36) NOT NULL FOREIGN KEY REFERENCES users(id),
    token varchar(255) NOT NULL,
    created_at datetime default getdate(),
    updated_at datetime NULL
);

--- roles table
CREATE TABLE roles (
    id char(36) PRIMARY KEY,
    name varchar(255) NOT NULL UNIQUE,
    description varchar(255) NULL,
    created_at datetime default getdate(),
    updated_at datetime NULL
);
--- permissions table
CREATE TABLE permissions (
    id char(36) PRIMARY KEY,
    name varchar(255) NOT NULL UNIQUE,
    description varchar(255) NULL,
    created_at datetime default getdate(),
    updated_at datetime NULL
);
--- permission_routes table
CREATE TABLE permission_routes (
    id char(36) PRIMARY KEY,
    permission_id char(36) NOT NULL FOREIGN KEY REFERENCES permissions(id),
    path varchar(255) NOT NULL,
    alias varchar(255) NOT NULL,
    created_at datetime default getdate(),
    updated_at datetime NULL
);
--- role_permissions table
CREATE TABLE role_permissions (
    id char(36) PRIMARY KEY,
    role_id char(36) NOT NULL FOREIGN KEY REFERENCES roles(id),
    permission_id char(36) NOT NULL FOREIGN KEY REFERENCES permissions(id),
    created_at datetime default getdate(),
    updated_at datetime NULL
);
--- user_roles table
CREATE TABLE user_roles (
    id char(36) PRIMARY KEY,
    user_id char(36) NOT NULL FOREIGN KEY REFERENCES users(id),
    role_id char(36) NOT NULL FOREIGN KEY REFERENCES roles(id),
    created_at datetime default getdate(),
    updated_at datetime NULL
);
--- user_permissions table
CREATE TABLE user_permissions (
    id char(36) PRIMARY KEY,
    user_id char(36) NOT NULL FOREIGN KEY REFERENCES users(id),
    permission_id char(36) NOT NULL FOREIGN KEY REFERENCES permissions(id),
    created_at datetime default getdate(),
    updated_at datetime NULL
);
--- failed_login_attempts table
CREATE TABLE failed_login_attempts (
    id char(36) PRIMARY KEY,
    user_id char(36) NOT NULL FOREIGN KEY REFERENCES users(id),
    created_at datetime default getdate(),
    updated_at datetime NULL
);

--- statement_print_audits table
CREATE TABLE statement_print_audits (
    id char(36) PRIMARY KEY,
    account_name varchar(255) NOT NULL,
    account_number char(13) NOT NULL,
    query_date_from datetime Not Null,
    query_date_to datetime Not Null,
    pages int NULL,
    print_type varchar(255) NULL,
    requested_by varchar(255) NOT NULL,
    requester_branch varchar(255) NOT NULL,
    created_at datetime default getdate(),
    updated_at datetime default getdate()
);

----inserting data for login access support services valid for 90days
insert into branches (id, name, code) values ( 'f9c33c6e-7cb9-487f-a7a6-f42b50355c86','head Office Branch', '020100');
insert into users (id, ab_number, first_name, last_name, email, password, branch_id) values('bc431588-44aa-2c4d-a4fa-28e29276a5bd','1018624', 'Support', 'Vendor','supporth@bectechservices.com', '$2a$12$M0jS/w/d4f/5mvVbnI1vzesn9dCOXE4wPYwbAvFhRVQ2yKXjGIVRi', 'f9c33c6e-7cb9-487f-a7a6-f42b50355c86');
insert into password_expiries (id, days, remind_in, length) values ('6efbc319-909b-4395-ab57-8f0c8fb77785',30, 45,12)







