CREATE TABLE IF NOT EXISTS person
(
    id           UUID   PRIMARY KEY,
    name         TEXT   NOT NULL,
    person_type  TEXT CHECK (person_type IN ('CEO', 'CTO', 'developer', 'externalRecruiter', 'internalRecruiter', 'HR',
                                             'jobAdvertiser', 'jobContact', 'other', 'unknown'))    NOT NULL,
    email        TEXT   NULLABLE UNIQUE,
    phone        TEXT   NULLABLE,
    notes        TEXT   NULLABLE,
    created_date DATETIME   NOT NULL    DEFAULT CURRENT_TIMESTAMP,
    updated_date DATETIME NULLABLE
);