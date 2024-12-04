## How to run
1. Clone repository
2. Go to created folder
3. `go mod tidy` - initialize and download dependencies
4. `go test ./... -v` - execute all tests  
5. `go build -o server .` - create an executable named `server` in the project directory
6. `./server` - run server

## Usage
1. Connect to the server using telnet or a custom TCP client:
2. Send commands such as: `PAYMENT|150` or simulate long-running requests `PAYMENT|5000`
3. To test shutdown - initiate the server shutdown using an interrupt signal (e.g., Ctrl+C)


## Notes
By default, the server listens on port :8080.
Grace period is configurable in the code. Update the gracePeriod value (default: 3 seconds) as needed.
Logs will provide more info about server activity, including request handling and shutdown progress.