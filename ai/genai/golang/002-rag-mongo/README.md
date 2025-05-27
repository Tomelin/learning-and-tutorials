# Go CLI for RAG with Gemini and MongoDB Vector Search

## Overview

This command-line interface (CLI) tool performs Retrieval Augmented Generation (RAG). It uses Google Gemini for language modeling and embedding generation, and MongoDB Atlas Vector Search to retrieve relevant context from your data. You ask a question, the CLI finds relevant documents in your MongoDB collection, and then Gemini uses that information to provide a comprehensive answer.

## Prerequisites

Before you begin, ensure you have the following:

*   **Go:** Version 1.22 or higher.
*   **Google Gemini API Key:** You can obtain one from [Google AI Studio](https://aistudio.google.com/app/apikey).
*   **MongoDB Atlas Account:** A running MongoDB Atlas cluster.
*   **MongoDB Collection with Vector Search Index:**
    *   A specific database and collection in your Atlas cluster.
    *   This collection must have a Vector Search Index configured.
    *   **Index Name:** The CLI currently expects the vector search index to be named `vector_index`.
    *   **Vector Embedding Field:** A field in your documents that stores the vector embeddings (e.g., `plot_embedding`).
    *   **Text Content Field:** A field in your documents that stores the original text content from which embeddings were generated (e.g., `plot`).

## Configuration

Configuration is managed primarily via environment variables.

**Required Environment Variables:**

*   `GEMINI_API_KEY`: Your Google Gemini API key.
*   `MONGO_URI`: Your MongoDB Atlas connection URI (e.g., `mongodb+srv://user:pass@cluster.mongodb.net/...`).
*   `MONGO_DATABASE`: The name of your MongoDB database.
*   `MONGO_COLLECTION`: The name of your MongoDB collection that contains the data and the vector index.
*   `MONGO_VECTOR_FIELD`: The name of the field in your MongoDB documents that stores the vector embeddings (e.g., `plot_embedding`). This field is used by the vector search index.
*   `MONGO_CONTENT_FIELD`: The name of the field in your MongoDB documents that stores the raw text content (e.g., `plot`). This content is retrieved and sent to the LLM.

*(Optional: While primarily configured via environment variables, the application uses Viper, which can also read from a `config.yaml` file if present in the execution directory or specified by the `PATH_CONFIG` environment variable. However, for simplicity, using environment variables is recommended for this CLI.)*

## Setup & Build

1.  **Clone the repository (or navigate to the project directory):**
    If you have the full project:
    ```bash
    git clone <repository-url>
    cd <repository-url>/ai/genai/golang/002-rag-mongo
    ```
    If you are already in the `002-rag-mongo` directory, proceed to the next step.

2.  **Build the CLI:**
    From within the `ai/genai/golang/002-rag-mongo` directory, run:
    ```bash
    go build -o rag-cli ./cmd/cli/main.go
    ```
    This will create an executable file named `rag-cli` in the current directory.

## Running the CLI

To use the CLI, provide your query using the `--query` (or `-q`) flag:

```bash
./rag-cli --query "your search query here"
```

**Example:**

```bash
./rag-cli -q "What are the benefits of using vector databases for AI applications?"
```

The CLI will then output logs indicating its progress (vector search, content generation) and finally print the answer from Gemini.

### Inserting Documents for RAG

The `insert` subcommand allows you to add new documents to your MongoDB collection, making them available for the RAG system. For each input text, the CLI will automatically generate an embedding using Google Gemini and store both the text and its embedding.

**Usage:**

```bash
./rag-cli insert [flags]
```

**Flags:**

*   `-t, --text "your text"`: The text content you want to insert directly.
*   `-f, --file "/path/to/file.txt"`: The path to a plain text file whose content you want to insert.

**Important:** You must provide either the `--text` flag or the `--file` flag, but not both.

**Examples:**

1.  **Insert text directly:**
    ```bash
    ./rag-cli insert --text "The quick brown fox jumps over the lazy dog."
    ```

2.  **Insert content from a file:**
    ```bash
    ./rag-cli insert --file ./documents/my_article.txt
    ```

## How it Works

1.  **User Query:** You provide a question or topic as a query string.
2.  **Query Embedding:** The CLI uses the Gemini API (specifically, an embedding model like `text-embedding-004`) to convert your query into a vector embedding.
3.  **Vector Search:** This query embedding is used to search your MongoDB Atlas collection via the configured vector search index (`vector_index`). The search identifies documents in your collection that are semantically similar to your query.
4.  **Contextual Augmentation:** The text content (from the field specified by `MONGO_CONTENT_FIELD`) of these retrieved documents is collected.
5.  **Content Generation:** The original query and the retrieved text content are combined into a prompt and sent to a Gemini generative model (e.g., `gemini-pro`). The model uses this augmented context to generate a comprehensive and relevant answer.

## (Optional) MongoDB Vector Index Setup Example

To use this CLI, you need a vector search index on your MongoDB collection. The specifics of the index will depend on your data and embedding model.

**Key considerations for your index definition:**

*   **`path`**: Must match the field name you set for `MONGO_VECTOR_FIELD`.
*   **`numDimensions`**: Must match the dimensionality of the embeddings your model generates. For example, Google's `text-embedding-004` model generates 768-dimensional embeddings.
*   **`similarity`**: Common choices are `cosine`, `euclidean`, or `dotProduct`. `cosine` is often recommended for text embeddings.

**Example JSON for an Atlas Vector Search Index (using IVF type):**

```json
{
  "fields": [
    {
      "type": "vector",
      "path": "your_vector_field_name",
      "numDimensions": 768,
      "similarity": "cosine"
    }
  ]
}
```

**Note:** This is a basic example. MongoDB Atlas offers various index types (e.g., IVF, HNSW) and configurations. Always consult the official [MongoDB documentation](https://www.mongodb.com/docs/atlas/atlas-vector-search/create-index/) for the most up-to-date and detailed instructions on creating vector search indexes. The CLI currently expects the index to be named `vector_index`.