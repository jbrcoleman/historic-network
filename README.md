# Historic-network
Historic Network Analyzer that shows relationships of historical figures
# Historical Network Visualizer

A Go application for visualizing connections between historical figures using interactive network graphs.

## Overview

Historical Network Visualizer is a web application that displays relationships between historical figures as an interactive network graph. The application allows users to explore how different historical figures were connected through mentorship, influence, rivalry, and other relationships.

## Features

- Interactive network graph visualization using D3.js
- Historical figures displayed as nodes, relationships as edges
- Color-coded connections based on relationship type
- Filter by era or connection type
- Detailed information panel for selected historical figures
- Zoom and pan capabilities for graph exploration
- REST API for accessing and modifying data

## Screenshots

*![Alt text](/img/app-screenshot.png?raw=true "App Screenshot")*

## Technology Stack

- **Backend**: Go with Gorilla Mux router
- **Frontend**: HTML, CSS, JavaScript with D3.js
- **Data**: In-memory storage (can be extended to use databases)

## Requirements

- Go 1.16 or higher
- Web browser with JavaScript enabled

## Installation

1. Clone the repository:
```bash
git clone https://github.com/yourusername/historical-network-visualizer.git
cd historical-network-visualizer
```

2. Install dependencies:
```bash
go get github.com/gorilla/mux
```

3. Build and run the application:
```bash
go build
./historical-network-visualizer
```

Alternatively, you can run directly with:
```bash
go run main.go
```

4. Access the application:
Open your web browser and navigate to `http://localhost:8080`

## API Endpoints

The application provides the following REST API endpoints:

- `GET /api/graph` - Get the complete graph data (nodes and links)
- `GET /api/people` - Get all historical figures
- `GET /api/people/{id}` - Get details for a specific historical figure
- `GET /api/connections` - Get all connections
- `POST /api/people` - Add a new historical figure
- `POST /api/connections` - Add a new connection

## Data Structure

### Person

```json
{
  "id": "unique-identifier",
  "name": "Person Name",
  "era": "Modern",
  "profession": "Physicist",
  "yearBirth": 1879,
  "yearDeath": 1955,
  "country": "Germany/USA",
  "info": "Biographical information",
  "group": 1
}
```

### Connection

```json
{
  "source": "person-id-1",
  "target": "person-id-2",
  "type": "mentor",
  "strength": 8,
  "description": "Detailed description of the relationship"
}
```

## Extending the Application

### Adding More Historical Figures and Connections

Edit the `initSampleData()` function in `main.go` to add more data.

### Implementing Database Storage

Replace the in-memory storage with a database solution:
1. Create a database schema matching the data structures
2. Update the handlers to use the database
3. Implement proper error handling and connection management

### Adding Authentication

Implement user authentication to allow collaborative editing:
1. Add user management endpoints
2. Implement JWT or session-based authentication
3. Add permission checks to the API endpoints
