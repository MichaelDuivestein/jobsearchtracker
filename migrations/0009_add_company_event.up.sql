CREATE TABLE IF NOT EXISTS company_event
(
    company_id      UUID        NOT NULL,
    event_id        UUID        NOT NULL,
    created_date    DATETIME    NOT NULL    DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (company_id, event_id),
    FOREIGN KEY (company_id) REFERENCES company(id),
    FOREIGN KEY (event_id) REFERENCES event(id)
);
