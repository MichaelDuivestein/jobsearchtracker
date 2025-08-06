CREATE TABLE IF NOT EXISTS company (
    id              UUID                                                                    PRIMARY KEY,
    name            TEXT                                                                    NOT NULL,
    company_type    TEXT CHECK(company_type IN ('employer', 'recruiter', 'consultancy'))    NOT NULL,
    notes           TEXT                                                                    NULLABLE,
    last_contact    DATETIME                                                                NULLABLE,
    created_date    DATETIME                                                                NOT NULL    DEFAULT CURRENT_TIMESTAMP,
    updated_date    DATETIME                                                                NULLABLE
);

CREATE INDEX index_company_name ON company(name);
CREATE INDEX index_created_date ON company(created_date);
CREATE INDEX index_last_contact ON company(last_contact);
