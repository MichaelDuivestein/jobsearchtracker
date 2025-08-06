# jobsearchtracker

GoLang learning project. Record job applications and interviews.

This project was made in order to learn proper Go project structure, database management and CRUD, dependency injection, REST services, and Testing frameworks.

At present the MVP contains one table, `company`. The endpoints and database contains a `create` and a `getByID` function. More functionality should be added in the near future.

Features demonstrated:
- Dependency Injection via `Dig`.
- SQL via `SQLite` and Go's `database/sql` package.
- Schema Migration via `golang-migrate`.
- REST APIs Via `Mux` and Go's `net/http` package.
- JSON.
- Unit and Integration testing via `stretchr/testify`.
