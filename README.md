# Golang Gin Clean Starter

You can join in the development (Open Source). **Let's Go!!!**

[![Go Report Card](https://goreportcard.com/badge/github.com/Caknoooo/go-gin-clean-starter)](https://goreportcard.com/report/github.com/Caknoooo/go-gin-clean-starter) [![Go Reference](https://pkg.go.dev/badge/github.com/Caknoooo/go-gin-clean-starter.svg)](https://pkg.go.dev/github.com/Caknoooo/go-gin-clean-starter) [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT) [![Release](https://img.shields.io/badge/release-v2.1.0-green.svg)](https://github.com/Caknoooo/go-gin-clean-starter/releases) <img align="right" width="200" height="200" alt="Go Gin Clean Architecture" src="https://github.com/user-attachments/assets/b7e2f353-bb6b-4ef1-88e9-6ab9bf2b8327" />

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.20-blue.svg)](https://golang.org/) [![PostgreSQL](https://img.shields.io/badge/PostgreSQL-%3E%3D%2015.0-blue.svg)](https://www.postgresql.org/) [![Docker](https://img.shields.io/badge/Docker-Supported-blue.svg)](https://www.docker.com/) [![Gin](https://img.shields.io/badge/Gin-Web%20Framework-red.svg)](https://gin-gonic.com/) [![GORM](https://img.shields.io/badge/GORM-ORM-green.svg)](https://gorm.io/)

## Introduction 👋
> This project implements **Clean Architecture** principles with the Controller–Service–Repository pattern. This approach emphasizes clear separation of responsibilities across different layers in Golang applications. The architecture helps keep the codebase clean, testable, and scalable by dividing application logic into distinct modules with well-defined boundaries.

<img width="1485" height="610" alt="Image" src="https://github.com/user-attachments/assets/918adf6d-9dc4-47fa-b9a6-3a10ca1e5242" />

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

## Running the Application 🏃‍♂️

There are two ways to run the application:

### Option 1: With Docker
1. Configure `.env` with your PostgreSQL credentials:
   ```bash
   DB_HOST=localhost
   DB_USER=postgres
   DB_PASS=your_password
   DB_NAME=your_database
   DB_PORT=5432
   ```
2. Build and start Docker containers:
   ```bash
   make init-docker
   ```
3. Run migrations and seeders:
   ```bash
   make migrate-seed-docker
   ```
4. The application will be available at `http://localhost:<port>`

**Docker Migration Commands:**
```bash
make migrate-docker                    # Run migrations in Docker
make migrate-status-docker            # Show migration status in Docker
make migrate-rollback-docker           # Rollback last batch in Docker
make migrate-rollback-batch-docker batch=<number>  # Rollback batch in Docker
make migrate-rollback-all-docker       # Rollback all in Docker
make migrate-create-docker name=<name> # Create migration in Docker
```

### Option 2: Without Docker
1. Configure `.env` with your PostgreSQL credentials:
   ```bash
   DB_HOST=localhost
   DB_USER=postgres
   DB_PASS=your_password
   DB_NAME=your_database
   DB_PORT=5432
   ```
2. Run the application:
   ```bash
   make migrate      # Run migrations
   make seed         # Run seeders (optional)
   make migrate-seed # Run Migrations + Seeder
   make run          # Start the application
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

### Migration Commands
```bash
make migrate                    # Run all pending migrations
make migrate-status            # Show migration status
make migrate-rollback          # Rollback the last batch
make migrate-rollback-batch batch=<number>  # Rollback specific batch
make migrate-rollback-all      # Rollback all migrations
make migrate-create name=<migration_name>  # Create new migration file
```

**Migration Examples:**
```bash
make migrate                                    # Run migrations
make migrate-status                            # Check migration status
make migrate-rollback                           # Rollback last batch
make migrate-rollback-batch batch=2             # Rollback batch 2
make migrate-rollback-all                      # Rollback all migrations
make migrate-create name=create_posts_table     # Create migration with entity
```

**Note:** When creating a migration with format `create_*_table`, the system will automatically:
- Create the entity file in `database/entities/`
- Add the entity to the migration file
- Add the entity to `database/migration.go` AutoMigrate section

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

## Advanced Usage 🔧

### Running Migrations, Seeders, and Scripts
You can run migrations, seed the database, and execute scripts while keeping the application running:

```bash
go run cmd/main.go --migrate:run --seed --run --script:example_script
```

**Available flags:**
- `--migrate` or `--migrate:run`: Apply all pending migrations
- `--migrate:status`: Show migration status
- `--migrate:rollback`: Rollback the last batch
- `--migrate:rollback <batch_number>`: Rollback specific batch
- `--migrate:rollback:all`: Rollback all migrations
- `--migrate:create:<migration_name>`: Create new migration file
- `--seed`: Seed the database with initial data
- `--script:example_script`: Run the specified script (replace `example_script` with your script name)
- `--run`: Keep the application running after executing the commands above

### Individual Commands

#### Database Migration
```bash
go run cmd/main.go --migrate:run              # Run all pending migrations
go run cmd/main.go --migrate:status          # Show migration status
go run cmd/main.go --migrate:rollback         # Rollback last batch
go run cmd/main.go --migrate:rollback 2       # Rollback batch 2
go run cmd/main.go --migrate:rollback:all     # Rollback all migrations
go run cmd/main.go --migrate:create:create_posts_table  # Create migration
```

**Migration System Features:**
- **Batch-based migrations**: Similar to Laravel, migrations are grouped in batches
- **Automatic entity creation**: When creating migration with format `create_*_table`, the system will:
  - Automatically create entity file in `database/entities/`
  - Add entity to migration file's AutoMigrate
  - Add entity to `database/migration.go` AutoMigrate section
- **Rollback support**: Rollback by batch or rollback all migrations
- **Status tracking**: View which migrations have been run and their batch numbers

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

## 🌟 Star History

[![Star History Chart](https://api.star-history.com/svg?repos=Caknoooo/go-gin-clean-starter&type=date&legend=top-left)](https://www.star-history.com/#Caknoooo/go-gin-clean-starter&type=date&legend=top-left)

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Gin Web Framework](https://gin-gonic.com/)
- [GORM](https://gorm.io/)
- [Samber/do](https://github.com/samber/do) for dependency injection
- [Go Playground Validator](https://github.com/go-playground/validator)
- [Testify](https://github.com/stretchr/testify) for testing