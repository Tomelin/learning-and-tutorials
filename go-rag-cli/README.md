# Go RAG CLI

This is a command-line interface (CLI) application built in Go that implements a basic Retrieval Augmented Generation (RAG) pattern. It takes a user's sentence (and optional tags), attempts to find a relevant answer in a MongoDB store, and if not found, queries an OpenAI LLM to get an answer. The new answer is then stored in MongoDB for future retrieval.

## Features

-   Accepts a user sentence and up to 3 optional tags via CLI flags.
-   Retrieves existing answers from a MongoDB collection based on tags.
-   If no relevant answer is found, queries an OpenAI LLM (e.g., GPT-3.5 Turbo).
-   Generates vector embeddings for sentences using OpenAI's embedding models (e.g., `text-embedding-ada-002`).
-   Saves new sentence-answer pairs, their embeddings, and tags to MongoDB.

## Prerequisites

-   Go (version 1.18 or higher recommended)
-   Access to a MongoDB instance.
-   An OpenAI API key.

## Setup & Configuration

1.  **Clone the repository (if applicable) or ensure you have the source code.**

2.  **Install Go dependencies:**
    Navigate to the `go-rag-cli` directory and run:
    ```bash
    go mod tidy 
    ```
    (This will ensure all dependencies listed in `go.mod` like the MongoDB driver and OpenAI library are downloaded.)

3.  **Set Environment Variables:**
    The application requires the following environment variables to be set:
    *   `OPENAI_API_KEY`: Your OpenAI API key.
    *   `MONGODB_URI`: The connection URI for your MongoDB instance (e.g., `mongodb://localhost:27017`).

    You can set them in your shell environment:
    ```bash
    export OPENAI_API_KEY="your_openai_api_key"
    export MONGODB_URI="your_mongodb_uri"
    ```

## Building the Application

Navigate to the `go-rag-cli` directory and run:
```bash
go build -o ragcli main.go store.go llm.go
```
This will create an executable file named `ragcli` (or `ragcli.exe` on Windows) in the current directory.

## Running the Application

Execute the compiled application from the `go-rag-cli` directory:

```bash
./ragcli -sentence "Your query here" -tags "tag1,tag2"
```

**CLI Flags:**

*   `-sentence="<user_query>"`: (Required) The sentence or question you want to ask.
*   `-tags="<tag1,tag2,tag3>"`: (Optional) Up to three comma-separated tags to help refine the search and provide context.

**Example:**

```bash
./ragcli -sentence "What is the capital of France?" -tags "france,capital,geography"
```

## RAG Implementation Notes

-   **MongoDB Storage:**
    -   Database: `ragdb`
    -   Collection: `entries`
    -   Document Structure: `{ sentence: string, vector: []float64, answer: string, tags: []string }`
-   **Retrieval:** The current retrieval mechanism from MongoDB primarily uses tag matching (`$all` operator).
-   **Vector Search:** True vector similarity search against the MongoDB store (using the generated embeddings) is a planned enhancement (TODO in `store.go`). For optimal vector search, a MongoDB Atlas Vector Search index would typically be used, or alternatively, client-side similarity calculations if not using Atlas.

## Development

### Running Tests
To run the unit tests:
```bash
go test ./...
```
