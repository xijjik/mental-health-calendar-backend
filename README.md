# Calendar Events API

This is a simple Golang API that uses SQLite to manage calendar events. The API allows users to add new events, read the entire list of events, and modify existing ones.

## Features

- Add new calendar events
- Retrieve all calendar events
- Modify existing calendar events

## Event Structure

Each calendar event consists of:

- Date (type: date)
- Content (type: string)
- Mood (type: string)

## Setup and Running

1. Ensure you have Go installed on your system.
2. Install the required dependencies:
   ```
   go get github.com/mattn/go-sqlite3
   go get github.com/gorilla/mux
   ```
3. Run the application:
   ```
   go run main.go
   ```
4. The API will be available at `http://localhost:8080`

## API Endpoints

- POST /events - Add a new event
- GET /events - Retrieve all events
- PUT /events/{id} - Modify an existing event

## License

This project is open-source and available under the MIT License.
