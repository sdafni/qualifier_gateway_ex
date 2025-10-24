# LLM Gateway Router

A production-ready Go-based gateway that routes LLM API requests to different providers (OpenAI, Anthropic, DeepSeek) based on virtual API keys.

## Features

- **Virtual Key Routing**: Maps virtual API keys to real provider API keys
- **Multi-Provider Support**: OpenAI, Anthropic (Claude), DeepSeek
- **Unified Endpoint**: Single `/chat/completions` endpoint for all providers
- **Comprehensive Logging**: Pretty-printed JSON logs with full request/response capture
- **Modular Architecture**: Clean separation of concerns using internal packages
- **Easy to Extend**: Add new providers by implementing a simple interface

## Configuration

### Environment Variables

- `KEYS_FILE`: Path to virtual keys configuration file (default: `keys.json`)
- `GATEWAY_PORT`: Port to run the gateway server on (default: `8080`)

### Virtual Keys Configuration

Create a `keys.json` file with your virtual key mappings:

```json
{
  "virtual_keys": {
    "vk_user1_openai": {
      "provider": "openai",
      "api_key": "sk-real-openai-key-123"
    },
    "vk_user2_anthropic": {
      "provider": "anthropic",
      "api_key": "sk-ant-real-anthropic-key-456"
    },
    "vk_user3_deepseek": {
      "provider": "deepseek",
      "api_key": "sk-deepseek-key-abc123"
    }
  }
}
```

**Note**: `keys.json` is gitignored by default to protect your API keys.

## Usage

### Running the Server

```bash
# With default configuration (uses keys.json)
go run main.go

# With custom keys file
KEYS_FILE=/path/to/custom-keys.json go run main.go

# With custom port
GATEWAY_PORT=3000 go run main.go
```

### Building the Binary

```bash
go build -o gateway
./gateway
```

### Making Requests

Send requests to `/chat/completions` with a virtual API key in the Authorization header:

#### OpenAI Request
```bash
curl -X POST http://localhost:8080/chat/completions \
  -H "Authorization: Bearer vk_user1_openai" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

#### Anthropic Request
```bash
curl -X POST http://localhost:8080/chat/completions \
  -H "Authorization: Bearer vk_user2_anthropic" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-5-sonnet-20241022",
    "messages": [{"role": "user", "content": "Hello!"}],
    "max_tokens": 1024
  }'
```

#### DeepSeek Request
```bash
curl -X POST http://localhost:8080/chat/completions \
  -H "Authorization: Bearer vk_user3_deepseek" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "deepseek-chat",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

## Testing with Example Scripts

The `examples/` directory contains Python test scripts:

```bash
cd examples
pip install -r requirements.txt

# Test OpenAI routing
python test_openai.py "Tell me a joke"

# Test Anthropic routing
python test_anthropic.py "Explain quantum computing"

# Test DeepSeek routing
python test_deepseek.py "Write a haiku"

# Test router validation (asks providers to identify themselves)
python test_router.py
```

## Logging

All LLM interactions are logged to both console and file with complete request/response details:

- **Console**: Pretty-printed JSON for easy reading
- **File**: `logs/llm_interactions.json` - Pretty-printed JSON format

### Log Format

```json
{
  "timestamp": "2025-10-24T14:30:00Z",
  "virtual_key": "vk_user1_openai",
  "provider": "openai",
  "method": "POST",
  "status": 200,
  "duration_ms": 1250,
  "request": {
    "model": "gpt-4",
    "messages": [...]
  },
  "response": {
    "choices": [...]
  }
}
```

## Error Handling

- `404 Not Found`: Invalid endpoint (only `/chat/completions` supported)
- `401 Unauthorized`: Missing or invalid virtual key
- `405 Method Not Allowed`: Non-POST requests
- `502 Bad Gateway`: Provider API unreachable
- `500 Internal Server Error`: Configuration or provider errors

## Architecture

The gateway uses a modular internal package structure:

```
internal/
├── config/         # Configuration loading
├── logger/         # Structured JSON logging
├── virtualkey/     # Virtual key validation & lookup
├── provider/       # Provider interface & implementations
└── gateway/        # HTTP routing logic
```

See individual package documentation for details.

## Adding New Providers

1. Create a new file in `internal/provider/` (e.g., `gemini.go`)
2. Implement the `Provider` interface:
   ```go
   type Gemini struct{}
   func (g *Gemini) GetEndpoint() string { ... }
   func (g *Gemini) SetAuthHeaders(req, key) { ... }
   func (g *Gemini) GetName() string { return "gemini" }
   ```
3. Register in `NewRegistry()` in `provider.go`
4. Add virtual keys to `keys.json`

That's it! No changes needed to main.go or gateway code.
