# Golang Clean Architecture

## Introduction 👋
> Clean Architecture is an approach to organizing code in an application that focuses on separating responsibilities and dependencies between components. In the context of Golang, Clean Architecture refers to the application of Clean Architecture principles in developing applications using the Go programming language.

Clean Architecture proposes a structured application design with several layers that have clear and isolated responsibilities. Each layer has a specific role and boundaries. Here are some common layers in Golang Clean Architecture

## Layers 🔥
- **Domain Layer:** This layer contains the core business definitions of the application. It is the innermost layer and does not depend on any other layers. It includes entities, business rules, and repository interfaces that will be implemented in the infrastructure layer.
- **Use Case Layer:** This layer holds the business logic specific to use cases in the application. Use cases provide operations and interactions between entities in the domain layer. Use cases do not depend on implementation details in the infrastructure layer.
- **Delivery Layer:** This layer is responsible for receiving and delivering data to and from the application. It typically consists of APIs, controllers, and presenters. This layer acts as the interface to interact with the outside world and can take input from users or deliver output to users.
- **Repository Layer:** This layer is responsible for implementing the repository interfaces defined in the domain layer. Repositories are used to access and store data from the storage (database, cache, APIs, etc.). This layer serves as a bridge between the domain layer and the infrastructure layer.
- **Infrastructure Layer:** This layer contains the technical details and implementation of the technologies used in the application, such as databases, networking, data storage, and external APIs. This layer depends on other layers and is used to implement the technical components required by the application.

## How To Use?
```
1. git clone https://github.com/Caknoooo/golang-clean-template.git
2. cd golang-clean-template
3. cp .env.example .env
4. configure .env with your postgres
DB_HOST = localhost
DB_USER = postgres
DB_PASS = 
DB_NAME = 
DB_PORT = 5432
5. Open terminal, follow the steps below:
- if you haven't downloaded postgres, you can download it first
- Run -> psql -U postgres
- Run -> Create database according to what you put in .env
- \c (your database)
- Run -> CREATE EXTENSION IF NOT EXISTS "uuid-ossp"
- Quit from postgres
6. go run main.go
```
![image](https://github.com/Caknoooo/golang-clean-template/assets/92671053/ceaf74a2-4015-42c4-a94a-9ddd6480b73b)