CREATE TABLE client_info (
    id char(36) PRIMARY KEY,
    national_id varchar(255) NOT NULL,
    surname varchar(255) NOT NULL,
    forenames varchar(255) NOT NULL,
    email varchar(255) NOT NULL,
    phone varchar(255) NOT NULL,
    picture varchar(255) NOT NULL,
    nationality varchar(255) NOT NULL,
    gender varchar(15) NOT NULL,
    card_id varchar(255) NOT NULL,
    birth_date date NOT NULL,
    card_valid_to date NOT NULL,
);