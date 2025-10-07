# jobsearchtracker

GoLang learning project. Record job applications and interviews.

This project was made in order to learn proper Go project structure, database management and CRUD, dependency 
  injection, REST services, and Testing frameworks.

At present the MVP contains CRUD methods for `company`, `person`, and `application`.

Features demonstrated:
- Dependency Injection via `Dig`.
- SQL via `SQLite` and Go's `database/sql` package.
- Schema Migration via `golang-migrate`.
- REST APIs Via `Mux` and Go's `net/http` package.
- JSON.
- Unit and Integration testing via `stretchr/testify`.
- OpenAPI documentation via `swaggo`.

## Before running
Due to the way that Swaggo behaves, OpenAPI documentation needs to be generated before anything will run:

`go:generate go run github.com/swaggo/swag/cmd/swag@latest init`

An IDE, such as Goland, can run this directly from `main.go`,

## Running tests
When running tests from the root directory, use `go test ./...`. Some integration tests required a package name change 
  in order to avoid import conflicts 

## OpenAPI Documentation
When the service is running, OpenAPI Documentation can be found at the `/swagger/` endpoint. Alternatively, it can be found in the `docs` folder.