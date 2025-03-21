# Golang Gin Gorm Starter
You can join in the development (Open Source). **Let's Go!!!**

## Introduction 👋
> Clean Architecture is an approach to organizing code in an application that focuses on separating responsibilities and dependencies between components. In the context of Golang, Clean Architecture refers to the application of Clean Architecture principles in developing applications using the Go programming language.


![image](https://github.com/user-attachments/assets/0b011bcc-f9c6-466e-a9da-964cce47a8bc)

## Logs Feature 📋

The application includes a built-in logging system that allows you to monitor and track system queries. You can access the logs through a modern, user-friendly interface.

### Accessing Logs
To view the logs:
1. Make sure the application is running
2. Open your browser and navigate to:
```bash
http://your-domain/logs
```

### Features
- **Monthly Filtering**: Filter logs by selecting different months
- **Real-time Refresh**: Instantly refresh logs with the refresh button
- **Expandable Entries**: Click on any log entry to view its full content
- **Modern UI**: Clean and responsive interface with glass-morphism design

![Logs Interface](https://github.com/user-attachments/assets/adda0afb-a1e4-4e05-b44e-87225fe63309)


## Prerequisite 🏆
- Go Version `>= go 1.20`
- PostgreSQL Version `>= version 15.0`

## How To Use
1. Clone the repository or **Use This Template**
  ```bash
  git clone https://github.com/Caknoooo/go-gin-clean-starter.git
  ```
2. Navigate to the project directory:
  ```bash
  cd go-gin-clean-starter
  ```
3. Copy the example environment file and configure it:
  ```bash 
  cp .env.example .env
  ```
There are 2 ways to do running
### With Docker
1. Build Docker
  ```bash
  make up
  ```
2. Run Initial UUID V4 for Auto Generate UUID
  ```bash
  make init-uuid
  ```
3. Run Migration and Seeder
  ```bash
  make migrate-seed
  ```

### Without Docker
1. Configure `.env` with your PostgreSQL credentials:
  ```bash
  DB_HOST=localhost
  DB_USER=postgres
  DB_PASS=
  DB_NAME=
  DB_PORT=5432
  ```
2. Open the terminal and follow these steps:
  - If you haven't downloaded PostgreSQL, download it first.
  - Run:
    ```bash
    psql -U postgres
    ```
  - Create the database according to what you put in `.env` => if using uuid-ossp or auto generate (check file **/entity/user.go**):
    ```bash
    CREATE DATABASE your_database;
    \c your_database
    CREATE EXTENSION IF NOT EXISTS "uuid-ossp"; // remove default:uuid_generate_v4() if you not use you can uncomment code in user_entity.go
    \q
    ``` 
3. Run the application:
  ```bash
  go run main.go
  ```

## Run Migrations, Seeder, and Script
To run migrations, seed the database, and execute a script while keeping the application running, use the following command:

```bash
go run main.go --migrate --seed --run --script:example_script
```

- ``--migrate`` will apply all pending migrations.
- ``--seed`` will seed the database with initial data.
- ``--script:example_script`` will run the specified script (replace ``example_script`` with your script name).
- ``--run`` will ensure the application continues running after executing the commands above.

#### Migrate Database 
To migrate the database schema 
```bash
go run main.go --migrate
```
This command will apply all pending migrations to your PostgreSQL database specified in `.env`

#### Seeder Database 
To seed the database with initial data:
```bash
go run main.go --seed
```
This command will populate the database with initial data using the seeders defined in your application.

#### Script Run
To run a specific script:
```bash
go run main.go --script:example_script
```
Replace ``example_script`` with the actual script name in **script.go** at script folder

If you need the application to continue running after performing migrations, seeding, or executing a script, always append the ``--run`` option.

## What did you get?
By using this template, you get a ready-to-go architecture with pre-configured endpoints. The template provides a structured foundation for building your application using Golang with Clean Architecture principles.

### Postman Documentation
You can explore the available endpoints and their usage in the [Postman Documentation](https://documenter.getpostman.com/view/29665461/2s9YJaZQCG). This documentation provides a comprehensive overview of the API endpoints, including request and response examples, making it easier to understand how to interact with the API.

### Issue / Pull Request Template

The repository includes templates for issues and pull requests to standardize contributions and improve the quality of discussions and code reviews.

- **Issue Template**: Helps in reporting bugs or suggesting features by providing a structured format to capture all necessary information.
- **Pull Request Template**: Guides contributors to provide a clear description of changes, related issues, and testing steps, ensuring smooth and efficient code reviews.