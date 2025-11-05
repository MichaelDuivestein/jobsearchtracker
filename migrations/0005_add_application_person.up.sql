CREATE TABLE IF NOT EXISTS application_person
(
    application_id  UUID        NOT NULL,
    person_id       UUID        NOT NULL,
    created_date    DATETIME    NOT NULL    DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (application_id, person_id),
    FOREIGN KEY (application_id) REFERENCES application(id),
    FOREIGN KEY (person_id) REFERENCES person(id)
);
