# Email Service with Daily Cron Jobs

This Go service handles email sending and includes automated daily email functionality.

## Features

- 📧 Individual email sending via HTTP API
- ⏰ Daily automated emails at 9:00 AM IST
- 🗄️ MongoDB integration for user data
- 🔄 Concurrent email sending with worker pools
- 🎨 Beautiful HTML email templates
- 🚀 Docker containerized
- 💪 Graceful shutdown handling

## Environment Variables

Create a `.env` file in the `go_email` directory:

```env
# Server Configuration
PORT=8080
ENVIRONMENT=development

# MongoDB Configuration
MONGO_URI=mongodb+srv://username:password@cluster.mongodb.net/
DB_NAME=your_database_name

# Email Configuration (SMTP)
MAIL_HOST=smtp.gmail.com
MAIL_USER=your-email@gmail.com
SMTP_MAIL_PASS=your-app-password
```

## API Endpoints

### 1. Send Individual Email

```bash
POST /send-email
Content-Type: application/json

{
  "firstName": "John",
  "lastName": "Doe",
  "email": "john@example.com",
  "subject": "Welcome!",
  "template": "accountCreationTemplate",
  "extraData": {}
}
```

### 2. Health Check

```bash
GET /health
```

### 3. Manual Daily Email Trigger (for testing)

```bash
POST /trigger-daily-email
```

## Cron Job Schedule

- **Daily emails**: Every day at 9:00 AM IST
- **Cron expression**: `0 9 * * *`
- **Timezone**: Asia/Kolkata (IST)

## Running the Service

### Development

```bash
cd go_email
go run main.go
```

### Docker

```bash
docker build -t email-service .
docker run -p 8080:8080 --env-file .env email-service
```

### Docker Compose

```bash
docker-compose up go-mail
```

## Daily Email Content

The daily emails include:

- 👋 Personalized greeting
- 📚 Learning reminders and motivation
- 🔥 Daily learning tips
- 💪 Progress encouragement
- 🎨 Beautiful responsive HTML design

## Logs

The service provides detailed logging:

- ✅ Successful operations
- ❌ Error conditions
- 📊 Email sending statistics
- 🕘 Cron job execution times

## Testing

### Test Individual Email

```bash
curl -X POST http://localhost:8080/send-email \
  -H "Content-Type: application/json" \
  -d '{
    "firstName": "Test",
    "lastName": "User",
    "email": "test@example.com",
    "subject": "Test Email",
    "template": "test"
  }'
```

### Test Daily Email (Manual Trigger)

```bash
curl -X POST http://localhost:8080/trigger-daily-email
```

### Health Check

```bash
curl http://localhost:8080/health
```

## Architecture

```
┌─────────────────────────────────────┐
│         Go Email Service            │
├─────────────────────────────────────┤
│ • HTTP Server (Port 8080)          │
│ • Cron Scheduler (Daily 9 AM)      │
│ • MongoDB Client                   │
│ • SMTP Email Sender                │
│ • Worker Pool (10 concurrent)      │
│ • Graceful Shutdown                │
└─────────────────────────────────────┘
                    ↕️
┌─────────────────────────────────────┐
│            MongoDB                  │
│     (User Data Storage)             │
└─────────────────────────────────────┘
```

## Troubleshooting

### Common Issues

1. **MongoDB Connection Failed**

   - Check MONGO_URI in .env file
   - Verify network connectivity
   - Ensure database exists

2. **SMTP Authentication Failed**

   - Use App Password for Gmail
   - Check MAIL_USER and SMTP_MAIL_PASS
   - Verify SMTP settings

3. **Cron Job Not Running**
   - Check logs for cron scheduler startup
   - Verify timezone settings
   - Ensure MongoDB connection is successful

### Debug Mode

Set environment variable for more verbose logging:

```bash
export LOG_LEVEL=debug
```

## Performance

- **Concurrent Workers**: 10 (configurable)
- **Email Rate**: ~100 emails/minute
- **Memory Usage**: ~50MB
- **CPU Usage**: Low (event-driven)

## Security

- Environment variables for sensitive data
- No hardcoded credentials
- Secure SMTP with TLS
- Input validation on all endpoints
