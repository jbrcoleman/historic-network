# Historical Network Visualizer

A Go application for visualizing connections between historical figures using interactive network graphs, enhanced with Wikipedia data scraping and NLP-based relationship analysis.

![App Screenshot](/img/app-screenshot.png)

## Overview

Historical Network Visualizer is a web application that displays relationships between historical figures as an interactive network graph. The application allows users to explore how different historical figures were connected through mentorship, influence, rivalry, and other relationships. With an integrated Wikipedia scraper, the application can automatically gather information about historical figures and analyze relationships between them.

## Features

### Core Features

- Interactive network graph visualization using D3.js
- Historical figures displayed as nodes, relationships as edges
- Color-coded connections based on relationship type
- Filter by era or connection type
- Detailed information panel for selected historical figures
- Zoom and pan capabilities for graph exploration
- REST API for accessing and modifying data

### Advanced Features

- **Wikipedia Scraping**: Automatically extract information about historical figures from Wikipedia
- **Natural Language Processing**: Analyze text to identify relationships between historical figures
- **Entity Recognition**: Extract mentions of historical figures from text
- **Relationship Analysis**: Determine the type and strength of relationships between figures
- **Batch Processing**: Import multiple historical figures at once
- **Interactive UI**: User-friendly interface for all the new features

## Technology Stack

- **Backend**: Go with Gorilla Mux router
- **Frontend**: HTML, CSS, JavaScript with D3.js
- **Data Processing**: Custom-built NLP analyzer for relationship extraction
- **Data Sources**: Wikipedia API integration for historical data
- **Data Storage**: In-memory storage (can be extended to use databases)

## Requirements

- Go 1.16 or higher
- Internet connection (for Wikipedia access)
- Web browser with JavaScript enabled

## Installation

1. Clone the repository:
```bash
git clone https://github.com/jbrcoleman/historic-network.git
cd historic-network
```

2. Install dependencies:
```bash
go mod tidy
```

3. Build the application:
```bash
go build
```

4. Run the application:
```bash
./historical-network-visualizer
```

Alternatively, you can run directly with:
```bash
go run *.go
```

5. Access the application:
   - Main visualization: `http://localhost:8080/`
   - Wikipedia scraper interface: `http://localhost:8080/wikipedia.html`

## Project Structure

```
historic-network/
│
├── main.go                       # Main application file
├── wiki_scraper.go               # Wikipedia scraping functionality
├── nlp_analyzer.go               # NLP analysis for relationships
├── wikipedia_handlers.go         # API handlers for Wikipedia integration
├── go.mod                        # Go module definition
├── go.sum                        # Go module checksums
│
├── static/                       # Static web files
│   ├── index.html                # Main graph visualization interface
│   ├── wikipedia.html            # Wikipedia scraping interface
│   └── img/                      # Images for the application
│       └── app-screenshot.png    # Screenshot of the application
│
└── README.md                     # Project documentation
```

## How to Use

### Main Network Visualization

1. Navigate to `http://localhost:8080/`
2. The main network graph will display historical figures and their relationships
3. Click on a node to see details about a historical figure
4. Use the filters to view specific eras or relationship types
5. Zoom and pan to explore the network

### Using the Wikipedia Scraper

#### Search and Import

1. Navigate to `http://localhost:8080/wikipedia.html`
2. Enter a historical figure's name in the search box
3. Click "Search" to find matching Wikipedia entries
4. Click "Import to Network" to add the figure to your graph

#### Batch Import

1. Go to the "Batch Import" tab
2. Enter multiple names, one per line
3. Click "Import All" to add all figures to your network

#### Text Analysis

1. Go to the "Text Analysis" tab
2. Paste text containing historical figures and their relationships
3. Click "Extract Entities" to identify historical figures in the text
4. Use "Analyze Relationship" to determine connections between two specific figures

## API Endpoints

### Graph Data Endpoints

- `GET /api/graph` - Get the complete graph data (nodes and links)
- `GET /api/people` - Get all historical figures
- `GET /api/people/{id}` - Get details for a specific historical figure
- `GET /api/connections` - Get all connections
- `POST /api/people` - Add a new historical figure
- `POST /api/connections` - Add a new connection

### Wikipedia Integration Endpoints

- `GET /api/wikipedia/search?q={query}` - Search Wikipedia for historical figures
- `POST /api/wikipedia/scrape` - Scrape a historical figure from Wikipedia
- `GET /api/wikipedia/relationships/{id}` - Find relationships for a historical figure
- `POST /api/wikipedia/batch-scrape` - Scrape multiple historical figures
- `POST /api/wikipedia/extract-entities` - Extract historical figures from text
- `POST /api/wikipedia/analyze-relationship` - Analyze a relationship between two figures

## Data Models

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

## Relationship Types

The application can detect various types of relationships between historical figures:

- **Mentor**: Teaching or guiding relationship
- **Student**: Learning from or being mentored by
- **Colleague**: Worked together or collaborated
- **Influenced**: Had an impact on the other's thinking or work
- **Rival**: Competitive or adversarial relationship
- **Friend**: Personal or close relationship
- **Admired**: Respected or looked up to

## Extending the Application

### Adding More Relationship Types

Edit the `initializeCorpus()` function in `nlp_analyzer.go` to add more relationship types and their associated keywords.

### Improving Entity Recognition

The current entity recognition is basic. To improve it, you could:
1. Implement a more sophisticated named entity recognition system
2. Add a curated list of known historical figures
3. Use a machine learning approach for more accurate detection

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

## Troubleshooting

### Rate Limiting Issues

Wikipedia may rate-limit requests if too many are made in a short time. The code includes basic throttling, but you might need to adjust the delay if you encounter issues.

### Missing Relationships

The NLP analysis is based on patterns found in Wikipedia text. If relationships are missing:
1. Check if the person pages exist on Wikipedia
2. Verify that the relationship is mentioned explicitly in the text
3. Consider adding custom relationships via the API

### Performance Considerations

For large networks, you might want to:
1. Add caching for Wikipedia requests
2. Implement pagination for graph visualization
3. Use a more efficient storage solution for large datasets

