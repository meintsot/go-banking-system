# Banking System Simulation

This project simulates a banking system to demonstrate various Go programming concepts and design patterns without implementing actual REST calls or database communication.

## Project Overview

The banking system simulation includes the following features:
- Customer management
- Account management (checking, savings, business)
- Transaction processing (deposits, withdrawals, transfers)
- External system integration via adapters
- Notification system with adapter pattern implementation
- Generic repository pattern using Go generics

## Key Go Concepts Demonstrated

1. **Structs and Interfaces**
   - Modeling domain entities with structs (Customer, Account, Transaction)
   - Interface-based design for loose coupling
   - Interface implementations for different behaviors

2. **Pointers and References**
   - Pointer receivers for methods that modify state
   - Passing pointers to repositories to ensure proper data manipulation

3. **Package Organization**
   - Modular code organization with domain-specific packages
   - Clean separation of concerns across packages

4. **Generics**
   - Generic repository implementation for type-safe data access
   - Type constraints with interfaces

5. **Design Patterns**
   - **Adapter Pattern**: For external system communication and notifications
   - **Repository Pattern**: For data access abstraction
   - **Dependency Injection**: For service composition

## Project Structure

```
bankingsystem/
├── main.go                    # Application entry point
└── pkg/                       # Package directory
    ├── models/                # Domain models
    │   ├── account.go         # Account entity
    │   ├── customer.go        # Customer entity
    │   └── transaction.go     # Transaction entity
    ├── services/              # Business logic
    │   ├── banking_service.go # Core banking functionality
    │   └── interfaces.go      # Service interfaces
    ├── adapters/              # External system adapters
    │   ├── external_systems.go # External system integration
    │   ├── generic_repository.go # Generic repository implementation
    │   └── memory_repository.go # In-memory data storage
    └── utils/                 # Utility functions
        └── id_generator.go    # UUID generation
```

## Running the Application

To run the banking system simulation:

```bash
go run main.go
```

This will execute the simulation which demonstrates:
1. Creating customers and accounts
2. Performing banking operations (deposits, withdrawals, transfers)
3. Using the adapter pattern to switch between notification methods
4. Demonstrating external system integration
5. Using the generic repository pattern

## Design Patterns Explained

### Adapter Pattern

The adapter pattern is implemented in two ways:
1. `NotificationAdapter`: Allows the system to switch between email and SMS notifications while providing a unified interface
2. `ReportingAdapter`: Allows the system to interact with external reporting systems

### Repository Pattern

The project implements both specific and generic repositories:
1. Type-specific repositories: `InMemoryCustomerRepository`, `InMemoryAccountRepository`, etc.
2. Generic repository: `GenericRepository[T Entity]` which works with any type that implements the `Entity` interface

### Dependency Injection

Services are composed by injecting their dependencies, making them testable and modular.