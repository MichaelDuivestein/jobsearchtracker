CREATE TABLE IF NOT EXISTS application_event
(
    application_id  UUID        NOT NULL,
    event_id       UUID        NOT NULL,
    created_date    DATETIME    NOT NULL    DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (application_id, event_id),
    FOREIGN KEY (application_id) REFERENCES application(id),
    FOREIGN KEY (event_id) REFERENCES event(id)
);
