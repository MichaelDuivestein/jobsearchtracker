CREATE TABLE IF NOT EXISTS event_person
(
    event_id      UUID        NOT NULL,
    person_id       UUID        NOT NULL,
    created_date    DATETIME    NOT NULL    DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (event_id, person_id),
    FOREIGN KEY (event_id) REFERENCES event(id),
    FOREIGN KEY (person_id) REFERENCES person(id)
);
