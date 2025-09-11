CREATE TABLE IF NOT EXISTS application
(
    id                      UUID        PRIMARY KEY,
    company_id              UUID        NULLABLE,
    recruiter_id            UUID        NULLABLE,
    job_title               TEXT        NULLABLE,
    job_ad_url              TEXT        NULLABLE,
    country                 TEXT        NULLABLE,
    area                    TEXT        NULLABLE,
    remote_status_type      TEXT        NOT NULL    CHECK (remote_status_type IN ('hybrid', 'office', 'remote', 'unknown')),
    weekdays_in_office      INT         NULLABLE,
    estimated_cycle_time    INT         NULLABLE,
    estimated_commute_time  INT         NULLABLE,
    application_date        DATETIME    NULLABLE,
    created_date            DATETIME    NOT NULL    DEFAULT CURRENT_TIMESTAMP,
    updated_date            DATETIME    NULLABLE,
    CONSTRAINT fk_application_company FOREIGN KEY(company_id) REFERENCES company(id),
    CONSTRAINT fk_application_recruiter FOREIGN KEY(recruiter_id) REFERENCES company(id),
    CONSTRAINT company_reference_not_null CHECK (company_id IS NOT NULL OR recruiter_id IS NOT NULL),
    CONSTRAINT job_title_job_url_not_null CHECK (job_title IS NOT NULL OR job_ad_url IS NOT NULL)
);
