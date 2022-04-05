**Go-Ping: HTTP Uptime Checker in Golang**

Checks periodically (custom defined cron expression) a list of URLs from a MongoDB collection for their HTTP status using goroutines.

When HTTP response has error status, an alert email is sent to a specified user.

How To Use:

 1. Modify .env.example and set the appropriate values. Rename it to .env
 2. Modify the cron expression and SMTP server in main.go if needed.
 3. `go run main.go`

