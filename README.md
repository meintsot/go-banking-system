# Banking System

A Go-based banking system with SQLite database storage and a REST API supporting both JSON and XML formats.

## Features

- Customer management (create, read, update, delete)
- Account management with different account types
- Transaction handling (deposits, withdrawals, transfers)
- Persistent storage using SQLite
- RESTful API with JSON and XML support

## Getting Started

### Prerequisites

- Go 1.16 or higher
- SQLite 3

### Installation

1. Clone the repository
2. Run `go build` to compile the application

### Running the Application

```
./BankingSystem -db ./banking.db -port 8080
```

Options:
- `-db`: Path to SQLite database file (default: "./banking.db")
- `-port`: API server port (default: 8080)

## API Documentation

The API supports both JSON and XML formats:
- For JSON: Set `Content-Type: application/json` and `Accept: application/json` headers
- For XML: Set `Content-Type: application/xml` and `Accept: application/xml` headers

### Customer Endpoints

#### List All Customers

```
GET /customers
```

#### Get a Customer

```
GET /customers/{id}
```

#### Create a Customer

```
POST /customers

JSON body:
{
  "first_name": "John",
  "last_name": "Doe",
  "email": "john.doe@example.com",
  "phone": "123-456-7890",
  "address": "123 Main St"
}

XML body:
<customer>
  <first_name>John</first_name>
  <last_name>Doe</last_name>
  <email>john.doe@example.com</email>
  <phone>123-456-7890</phone>
  <address>123 Main St</address>
</customer>
```

#### Update a Customer

```
PUT /customers/{id}

JSON body (fields to update):
{
  "phone": "555-123-4567",
  "address": "456 Oak Ave"
}

XML body (fields to update):
<customer>
  <phone>555-123-4567</phone>
  <address>456 Oak Ave</address>
</customer>
```

#### Delete a Customer

```
DELETE /customers/{id}
```

### Account Endpoints

#### Get an Account

```
GET /accounts/{id}
```

#### Get Customer Accounts

```
GET /customers/{id}/accounts
```

#### Create an Account

```
POST /accounts

JSON body:
{
  "customer_id": "customer-uuid",
  "initial_balance": 1000.00,
  "account_type": "CHECKING"
}

XML body:
<account>
  <customer_id>customer-uuid</customer_id>
  <initial_balance>1000.00</initial_balance>
  <account_type>CHECKING</account_type>
</account>
```

#### Delete an Account

```
DELETE /accounts/{id}
```

### Transaction Endpoints

#### Deposit to an Account

```
POST /accounts/{id}/deposit

JSON body:
{
  "amount": 500.00
}

XML body:
<money>
  <amount>500.00</amount>
</money>
```

#### Withdraw from an Account

```
POST /accounts/{id}/withdraw

JSON body:
{
  "amount": 200.00
}

XML body:
<money>
  <amount>200.00</amount>
</money>
```

#### Transfer Between Accounts

```
POST /transfers?from={source_account_id}

JSON body:
{
  "destination_account_id": "destination-account-id",
  "amount": 300.00
}

XML body:
<money>
  <destination_account_id>destination-account-id</destination_account_id>
  <amount>300.00</amount>
</money>
```

#### Get Account Transactions

```
GET /accounts/{id}/transactions
```

## Error Handling

The API returns appropriate HTTP status codes and error messages:

- 200: Success
- 201: Created
- 204: No Content (successful delete)
- 400: Bad Request (validation error)
- 404: Not Found
- 409: Conflict (e.g., trying to delete a customer with active accounts)
- 415: Unsupported Media Type
- 500: Internal Server Error