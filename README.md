# File Service API

A Go Fiber-based file service application that provides secure file upload and download functionality with PostgreSQL database integration.

## Features

- **Secure File Upload**: Upload files using signed URLs that expire and can only be used once
- **Individual File Download**: Download specific files by their ID
- **Share Download**: Download entire shares as individual files or compressed ZIP archives
- **Database Analytics**: Track download and visit analytics
- **PostgreSQL Integration**: Uses GORM for database operations
- **Configurable Storage**: Environment-based configuration for file storage paths
- **Modular Architecture**: Clean separation of concerns with organized packages

## Project Structure

```
pss-fs/
├── main.go                    # Application entry point
├── config/
│   └── config.go             # Configuration management
├── database/
│   └── database.go           # Database connection and migrations
├── handlers/
│   ├── upload.go             # File upload handler
│   ├── download.go           # Download handlers (file & share)
│   └── health.go             # Health check handler
├── models/
│   └── models.go             # Database models/structs
├── utils/
│   └── utils.go              # Utility functions
├── config.env.template       # Environment configuration template
├── schema.ts                 # TypeScript schema reference
└── README.md                 # Documentation
```

## Architecture

The application follows Go best practices with a modular architecture:

- **`config/`**: Centralized configuration management with environment variable loading
- **`database/`**: Database connection, initialization, and migrations
- **`handlers/`**: HTTP request handlers organized by functionality
- **`models/`**: Database models that match the TypeScript Drizzle schema
- **`utils/`**: Shared utility functions
- **`main.go`**: Clean entry point that orchestrates the application startup

## API Endpoints

### 1. Upload Files

```
POST /up/{signature}
```

- Validates and invalidates the provided signature
- Accepts multipart form data with files under the "files" field
- Registers each uploaded file in the database
- Updates share statistics

### 2. Download Individual File

```
GET /d/f/{fileID}
```

- Downloads a specific file by its UUID
- Logs download analytics
- Updates share download count

### 3. Download Share

```
GET /d/s/{shareID}
```

- Downloads all files in a share
- Single file: serves directly
- Multiple files: creates and serves a ZIP archive
- Logs download analytics

### 4. Health Check

```
GET /
```

- Returns service status

## Database Schema

The application uses the following main tables (based on your TypeScript schema):

- `ps_files`: File records with metadata and storage paths
- `ps_shares`: Share information and statistics
- `ps_upload_signatures`: One-time upload signatures with expiry
- `ps_download_analytics`: Download tracking data
- `ps_visit_analytics`: Visit tracking data

## Setup Instructions

### 1. Prerequisites

- Go 1.22.0 or higher
- PostgreSQL database
- Git

### 2. Clone and Setup

```bash
git clone <your-repo-url>
cd pss-fs
go mod tidy
```

### 3. Environment Configuration

Copy the configuration template and modify it:

```bash
cp config.env.template .env
```

Edit `.env` with your actual configuration:

```env
# Database Configuration
DB_HOST=localhost
DB_USER=postgres
DB_PASSWORD=your_actual_password
DB_NAME=your_database_name
DB_PORT=5432
DB_SSLMODE=disable
DB_TIMEZONE=UTC

# Server Configuration
PORT=3000

# File Storage Configuration
FILES_DIRECTORY=./files
```

### 4. Database Setup

Ensure your PostgreSQL database is running and accessible with the credentials in your `.env` file. The application will automatically create the required tables on startup.

### 5. Run the Application

```bash
go run main.go
```

The server will start on the configured port (default: 3000).

## Usage Examples

### Upload Files

```bash
# First, you need a valid upload signature in your database
curl -X POST "http://localhost:3000/up/your-signature-here" \
  -F "files=@/path/to/file1.txt" \
  -F "files=@/path/to/file2.pdf"
```

### Download Individual File

```bash
curl -o downloaded_file "http://localhost:3000/d/f/file-uuid-here"
```

### Download Share

```bash
# Single file share - downloads the file directly
curl -o downloaded_file "http://localhost:3000/d/s/share-uuid-here"

# Multiple file share - downloads as ZIP
curl -o share_archive.zip "http://localhost:3000/d/s/share-uuid-here"
```

## Configuration Options

| Environment Variable | Description              | Default   |
| -------------------- | ------------------------ | --------- |
| `DB_HOST`            | PostgreSQL host          | localhost |
| `DB_USER`            | PostgreSQL username      | postgres  |
| `DB_PASSWORD`        | PostgreSQL password      | -         |
| `DB_NAME`            | PostgreSQL database name | -         |
| `DB_PORT`            | PostgreSQL port          | 5432      |
| `DB_SSLMODE`         | PostgreSQL SSL mode      | disable   |
| `DB_TIMEZONE`        | Database timezone        | UTC       |
| `PORT`               | Server port              | 3000      |
| `FILES_DIRECTORY`    | Local file storage path  | ./files   |

## Security Features

- **Signature Validation**: Upload signatures are validated and can only be used once
- **Expiry Checking**: Signatures have expiration timestamps
- **UUID-based IDs**: All file and share IDs use UUIDs for security
- **Soft Deletion**: Files support soft deletion (deleted_at timestamp)

## Analytics and Tracking

The application automatically tracks:

- Download events with IP addresses and user agents
- File access patterns
- Share popularity metrics
- Visit analytics for shares

## File Storage

- Files are stored locally in the configured directory
- Each uploaded file gets a unique filename (UUID + original name)
- Storage URLs are saved in the database for efficient retrieval
- ZIP archives for multi-file shares are created dynamically and cleaned up after serving

## Error Handling

The API returns appropriate HTTP status codes:

- `200`: Success
- `400`: Bad request (invalid parameters)
- `401`: Unauthorized (invalid signature)
- `404`: Resource not found
- `500`: Internal server error

## Development Notes

- The application uses GORM for database operations
- Fiber framework provides the HTTP server functionality
- UUID v4 is used for all entity identifiers
- File uploads support multipart form data
- ZIP compression is handled in-memory for optimal performance

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

## License

[Add your license information here]
