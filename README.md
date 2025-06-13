# Payment Service IAE

A GraphQL-based payment service built with Go, integrating with Midtrans payment gateway for secure transaction processing.

## 🚀 Features

- **GraphQL API** - Modern API with flexible queries and mutations
- **Midtrans Integration** - Secure payment processing with Midtrans Snap
- **Authentication Middleware** - JWT-based user authentication
- **Health Check** - Service monitoring endpoint
- **Docker Support** - Containerized deployment
- **Environment Configuration** - Flexible configuration management

## 📋 Prerequisites

- Go 1.21 or higher
- PostgreSQL database
- Midtrans account (Sandbox/Production)
- Docker (optional)

## 🛠 Installation

### 1. Clone the Repository

```bash
git clone https://github.com/your-username/payment-service-iae.git
cd payment-service-iae
```

### 2. Install Dependencies

```bash
go mod download
```

### 3. Environment Configuration

Create a `.env` file in the root directory:

```env
PORT=9210
DATABASE_URL=postgres://postgres:1235813@localhost:5433/payment_db?sslmode=disable
AUTH_SERVICE_URL=http://localhost:8081
USER_SERVICE_URL=http://localhost:8082
MIDTRANS_SERVER_KEY=SB-Mid-server-YOUR_SERVER_KEY
MIDTRANS_ENV=sandbox
JWT_SECRET=your_jwt_secret
MIDTRANS_CLIENT_KEY=SB-Mid-client-YOUR_CLIENT_KEY
MIDTRANS_BASE_URL=https://api.sandbox.midtrans.com/v2
```

### 4. Run the Service

```bash
go run main.go
```

The service will be available at `http://localhost:9210`

## 🐳 Docker Deployment

### Build Docker Image

```bash
docker build -t payment-service-iae .
```

### Run Container

```bash
docker run -p 9210:9210 --env-file .env payment-service-iae
```

## 📖 API Documentation

### GraphQL Schema

```graphql
type Query {
  healthCheck: String!
}

type PaymentResponse {
  orderId: String!
  bookId: String!
  customerId: String!
  token: String!
  redirect_url: String!
}

type Mutation {
  createPayment(
    amount: Int!
    bookId: String!
    customerId: String!
  ): PaymentResponse!
}
```

## 🔍 GraphQL Query Examples

### 1. Health Check Query

**Query:**
```graphql
query HealthCheck {
  healthCheck
}
```

**Response:**
```json
{
  "data": {
    "healthCheck": "OK"
  }
}
```

### 2. Create Payment Mutation

**Mutation:**
```graphql
mutation CreatePayment($amount: Int!, $bookId: String!, $customerId: String!) {
  createPayment(amount: $amount, bookId: $bookId, customerId: $customerId) {
    orderId
    bookId
    customerId
    token
    redirect_url
  }
}
```

**Variables:**
```json
{
  "amount": 100000,
  "bookId": "book-12345",
  "customerId": "customer-67890"
}
```

**Response:**
```json
{
  "data": {
    "createPayment": {
      "orderId": "BOOK-book-12345-CUST-customer-67890-1703123456-abc12345",
      "bookId": "book-12345",
      "customerId": "customer-67890",
      "token": "66e4fa55-fdac-4ef9-91b5-733b97d1b862",
      "redirect_url": "https://app.sandbox.midtrans.com/snap/v3/redirection/66e4fa55-fdac-4ef9-91b5-733b97d1b862"
    }
  }
}
```

## 🖥 Using GraphQL Playground

1. Start the service
2. Open your browser and navigate to `http://localhost:9210`
3. You'll see the GraphQL Playground interface
4. Use the examples above to test the API

### Complete Payment Flow Example

```graphql
# Step 1: Create a payment transaction
mutation {
  createPayment(
    amount: 250000
    bookId: "book-programming-101"
    customerId: "user-john-doe"
  ) {
    orderId
    bookId
    customerId
    token
    redirect_url
  }
}

# Step 2: Use the redirect_url to complete payment on Midtrans
# The customer will be redirected to Midtrans payment page
```

## 🧪 Testing with cURL

### Health Check
```bash
curl -X POST http://localhost:9210/query \
  -H "Content-Type: application/json" \
  -d '{"query": "query { healthCheck }"}'
```

### Create Payment
```bash
curl -X POST http://localhost:9210/query \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "query": "mutation CreatePayment($amount: Int!, $bookId: String!, $customerId: String!) { createPayment(amount: $amount, bookId: $bookId, customerId: $customerId) { orderId bookId customerId token redirect_url } }",
    "variables": {
      "amount": 100000,
      "bookId": "book-12345",
      "customerId": "customer-67890"
    }
  }'
```

## 🔐 Authentication

The service uses JWT-based authentication. Include the JWT token in the Authorization header:

```
Authorization: Bearer YOUR_JWT_TOKEN
```

For development, the service currently includes a mock authentication middleware that accepts any request.

## 📁 Project Structure

```
payment-service-iae/
├── config/
│   └── config.go           # Configuration management
├── graph/
│   ├── model/
│   │   └── models_gen.go   # Generated GraphQL models
│   ├── generated.go        # Generated GraphQL code
│   ├── resolver.go         # Resolver dependencies
│   ├── schema.graphqls     # GraphQL schema definition
│   └── schema.resolvers.go # Resolver implementations
├── midtrans/
│   └── client.go          # Midtrans client wrapper
├── .env                   # Environment variables
├── .gitignore
├── Dockerfile
├── go.mod
├── go.sum
├── gqlgen.yml            # GraphQL generator config
├── main.go               # Application entry point
└── README.md
```

## 🌍 Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `PORT` | Server port | `9210` |
| `DATABASE_URL` | PostgreSQL connection string | `postgres://user:pass@localhost/db` |
| `MIDTRANS_SERVER_KEY` | Midtrans server key | `SB-Mid-server-xxx` |
| `MIDTRANS_CLIENT_KEY` | Midtrans client key | `SB-Mid-client-xxx` |
| `MIDTRANS_ENV` | Midtrans environment | `sandbox` or `production` |
| `JWT_SECRET` | JWT signing secret | `your-secret-key` |

## 🔧 Development

### Generate GraphQL Code

After modifying the schema, regenerate the GraphQL code:

```bash
go run github.com/99designs/gqlgen generate
```

### Add New Resolvers

1. Update `graph/schema.graphqls`
2. Run code generation
3. Implement resolvers in `graph/schema.resolvers.go`

## 📊 Monitoring

### Health Check Endpoint

```bash
curl http://localhost:9210/query -d '{"query": "{ healthCheck }"}'
```

Expected response: `{"data": {"healthCheck": "OK"}}`

## 🛡 Security Considerations

- Always use HTTPS in production
- Implement proper JWT validation
- Use environment variables for sensitive data
- Enable CORS appropriately
- Implement rate limiting
- Validate all input parameters

## 🚀 Production Deployment

1. Set `MIDTRANS_ENV=production`
2. Use production Midtrans credentials
3. Configure proper database connection
4. Set up SSL/TLS certificates
5. Implement proper logging and monitoring
6. Use a reverse proxy (nginx/Apache)

## 📝 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## 📞 Support

For support and questions:
- Create an issue in the repository
- Contact the development team
- Check the documentation

---

**Note:** This is a development version. Ensure proper security measures are implemented before deploying to production.