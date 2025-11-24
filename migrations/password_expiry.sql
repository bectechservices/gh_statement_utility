CREATE TABLE password_expiries (
    id char(36) PRIMARY KEY,
    days integer NOT NULL,
    remind_in integer NOT NULL,
    created_at datetime default getdate(),
    updated_at datetime NULL
);