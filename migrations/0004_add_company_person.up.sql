CREATE TABLE IF NOT EXISTS company_person
(
    company_id      UUID        NOT NULL,
    person_id       UUID        NOT NULL,
    created_date    DATETIME    NOT NULL    DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (company_id, person_id),
    FOREIGN KEY (company_id) REFERENCES company(id),
    FOREIGN KEY (person_id) REFERENCES person(id)
);
