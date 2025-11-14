CREATE TABLE IF NOT EXISTS event
(
    id           UUID       PRIMARY KEY,
    event_type   TEXT CHECK (event_type IN ('applied', 'callBooked', 'callCompleted', 'codeTestCompleted',
                                           'codeTestReceived', 'interviewBooked', 'interviewCompleted', 'paused',
                                           'offer', 'other', 'recruiterInterviewBooked', 'recruiterInterviewCompleted',
                                           'rejected', 'signed', 'withdrew'))    NOT NULL,
    description  TEXT       NULLABLE,
    notes        TEXT       NULLABLE,
    event_date   DATETIME   NOT NULL,
    created_date DATETIME   NOT NULL    DEFAULT CURRENT_TIMESTAMP,
    updated_date DATETIME   NULLABLE
);
