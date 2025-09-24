# Golang Gin Clean Starter

You can join in the development (Open Source). **Let's Go!!!**

[![Go Report Card](https://goreportcard.com/badge/github.com/Caknoooo/go-gin-clean-starter)](https://goreportcard.com/report/github.com/Caknoooo/go-gin-clean-starter) [![Go Reference](https://pkg.go.dev/badge/github.com/Caknoooo/go-gin-clean-starter.svg)](https://pkg.go.dev/github.com/Caknoooo/go-gin-clean-starter) [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT) [![Release](https://img.shields.io/badge/release-v2.1.0-green.svg)](https://github.com/Caknoooo/go-gin-clean-starter/releases) <img align="right" width="200" height="200" alt="Go Gin Clean Architecture" src="https://github.com/user-attachments/assets/b7e2f353-bb6b-4ef1-88e9-6ab9bf2b8327" />

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.20-blue.svg)](https://golang.org/) [![PostgreSQL](https://img.shields.io/badge/PostgreSQL-%3E%3D%2015.0-blue.svg)](https://www.postgresql.org/) [![Docker](https://img.shields.io/badge/Docker-Supported-blue.svg)](https://www.docker.com/) [![Gin](https://img.shields.io/badge/Gin-Web%20Framework-red.svg)](https://gin-gonic.com/) [![GORM](https://img.shields.io/badge/GORM-ORM-green.svg)](https://gorm.io/)

## Introduction 👋
> This project implements **Clean Architecture** principles with the Controller–Service–Repository pattern. This approach emphasizes clear separation of responsibilities across different layers in Golang applications. The architecture helps keep the codebase clean, testable, and scalable by dividing application logic into distinct modules with well-defined boundaries.

<img width="1485" height="610" alt="Image" src="https://github.com/user-attachments/assets/918adf6d-9dc4-47fa-b9a6-3a10ca1e5242" />

## Logs Feature 📋

The application includes a built-in logging system that allows you to monitor and track system queries. You can access the logs through a modern, user-friendly interface.

### Accessing Logs
To view the logs:
1. Make sure the application is running
2. Open your browser and navigate to:
```bash
http://your-domain/logs
```

![Logs Interface](https://github.com/user-attachments/assets/adda0afb-a1e4-4e05-b44e-87225fe63309)

### Features
- **Monthly Filtering**: Filter logs by selecting different months
- **Real-time Refresh**: Instantly refresh logs with the refresh button
- **Expandable Entries**: Click on any log entry to view its full content
- **Modern UI**: Clean and responsive interface with glass-morphism design

## Quick Start 🚀

### Prerequisites
- Go Version `>= go 1.20`
- PostgreSQL Version `>= version 15.0`

### Installation
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
4. Install dependencies:
   ```bash
   make dep
   ```

## Available Make Commands 🚀
The project includes a comprehensive Makefile with the following commands:

### Development Commands
```bash
make dep          # Install and tidy dependencies
make run          # Run the application locally
make build        # Build the application binary
make run-build    # Build and run the application
```

### Module Generation Commands
```bash
make module name=<module_name>  # Generate a new module with all necessary files
```

**Example:**
```bash
make module name=product
```

This command will automatically create a complete module structure including:
- Controller (`product_controller.go`)
- Service (`product_service.go`) 
- Repository (`product_repository.go`)
- DTO (`product_dto.go`)
- Validation (`product_validation.go`)
- Routes (`routes.go`)
- Test files for all components
- Query directory (for custom queries)

The generated module follows Clean Architecture principles and is ready to use with proper dependency injection setup.

### Testing Commands
```bash
make test-auth      # Run auth module tests only
make test-user      # Run user module tests only
make test-all       # Run tests for all modules
make test-coverage  # Run tests with coverage report
```

### Local Database Commands (without Docker)
```bash
make migrate-local      # Run migrations locally
make seed-local        # Run seeders locally  
make migrate-seed-local # Run migrations + seeders locally
```

### Docker Commands
```bash
make init-docker       # Initialize and build Docker containers
make up               # Start Docker containers
make down             # Stop Docker containers
make logs             # View Docker logs
```

### Docker Database Commands
```bash
make migrate          # Run migrations in Docker
make seed            # Run seeders in Docker
make migrate-seed    # Run migrations + seeders in Docker
make container-go    # Access Go container shell
make container-postgres # Access PostgreSQL container
```

## Running the Application 🏃‍♂️

There are two ways to run the application:

### Option 1: With Docker
1. Build and start Docker containers:
   ```bash
   make init-docker
   ```
2. Initialize UUID V4 extension for auto-generated UUIDs:
   ```bash
   make init-uuid
   ```
3. Run migrations and seeders:
   ```bash
   make migrate-seed
   ```
4. The application will be available at `http://localhost:8080`

### Option 2: Without Docker
1. Configure `.env` with your PostgreSQL credentials:
   ```bash
   DB_HOST=localhost
   DB_USER=postgres
   DB_PASS=your_password
   DB_NAME=your_database
   DB_PORT=5432
   ```
2. Set up PostgreSQL:
   - Download and install PostgreSQL if you haven't already
   - Create a database:
     ```bash
     psql -U postgres
     CREATE DATABASE your_database;
     \c your_database
     CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
     \q
     ```
3. Run the application:
   ```bash
   make migrate-local    # Run migrations
   make seed-local       # Run seeders (optional)
   make run              # Start the application
   ```

## Advanced Usage 🔧

### Running Migrations, Seeders, and Scripts
You can run migrations, seed the database, and execute scripts while keeping the application running:

```bash
go run cmd/main.go --migrate --seed --run --script:example_script
```

**Available flags:**
- `--migrate`: Apply all pending migrations
- `--seed`: Seed the database with initial data
- `--script:example_script`: Run the specified script (replace `example_script` with your script name)
- `--run`: Keep the application running after executing the commands above

### Individual Commands

#### Database Migration
```bash
go run cmd/main.go --migrate
```
This command will apply all pending migrations to your PostgreSQL database specified in `.env`

#### Database Seeding
```bash
go run cmd/main.go --seed
```
This command will populate the database with initial data using the seeders defined in your application.

#### Script Execution
```bash
go run cmd/main.go --script:example_script
```
Replace `example_script` with the actual script name in **script.go** at the script folder.

> **Note:** If you need the application to continue running after performing migrations, seeding, or executing a script, always append the `--run` option.

## What You Get 🎁

By using this template, you get a production-ready architecture with:

### 🏗️ Clean Architecture Implementation
- **Controller-Service-Repository pattern** with clear separation of concerns
- **Dependency injection** using samber/do
- **Modular structure** for easy maintenance and testing
- **Consistent code organization** across all modules

### 🚀 Pre-configured Features
- **Authentication system** with JWT tokens
- **User management** with email verification
- **Password reset** functionality
- **Database migrations** and seeders
- **Comprehensive logging** system with web interface
- **CORS middleware** for cross-origin requests
- **Input validation** with go-playground/validator

### 📚 Documentation & Testing
- **Postman collection** for API testing
- **Comprehensive test suite** for all modules
- **Code coverage** reporting
- **Issue and PR templates** for better collaboration

### 🔧 Developer Experience
- **Hot reload** with Air for development
- **Docker support** for easy deployment
- **Make commands** for common tasks
- **Module generator** for rapid development
- **Structured logging** with query tracking

## 📖 Documentation

### API Documentation
Explore the available endpoints and their usage in the [Postman Documentation](https://documenter.getpostman.com/view/29665461/2s9YJaZQCG). This documentation provides a comprehensive overview of the API endpoints, including request and response examples.

### Contributing
We welcome contributions! The repository includes templates for issues and pull requests to standardize contributions and improve the quality of discussions and code reviews.

- **Issue Template**: Helps in reporting bugs or suggesting features by providing a structured format
- **Pull Request Template**: Guides contributors to provide clear descriptions of changes and testing steps

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Gin Web Framework](https://gin-gonic.com/)
- [GORM](https://gorm.io/)
- [Samber/do](https://github.com/samber/do) for dependency injection
- [Go Playground Validator](https://github.com/go-playground/validator)
- [Testify](https://github.com/stretchr/testify) for testing