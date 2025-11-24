CREATE TABLE statement_print_audits (
    id char(36) PRIMARY KEY,
    user_id  char(36) NOT NULL,
    customer_branch_id  char(36) NOT NULL,
    branch_requested_id  char(36) NOT NULL,
    account_name varchar(255) NOT NULL,
    account_number char(13) NOT NULL,
    pages int(10) NOT NULL,
    print_type int(10) NOT NULL,
    activity_by char(36) NOT NULL FOREIGN KEY REFERENCES users(id),
    created_at datetime default getdate(),
    updated_at datetime default getdate()
);


