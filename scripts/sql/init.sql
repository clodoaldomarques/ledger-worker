CREATE TABLE IF NOT EXISTS event(
    event_id varchar(50) primary key,
    correlation_id varchar(50) not null,
    org_id varchar(60) not null,
    processing_code varchar(100),
    program_id bigint not null,
    account_id bigint not null,
    description varchar(200) not null,
    producer varchar(100) not null,
    created_at timestamp,
    updated_at timestamp
);

CREATE TABLE IF NOT EXISTS entry(
    event_id varchar(50) not null,
    entry_type_id bigint not null,
    amount decimal(20, 15) not null,
    debit_account varchar(60),
    credit_account varchar(60),
    description varchar(200) not null,
    primary key(event_id, entry_type_id)
);