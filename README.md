# LLM Gateway Router

A Go-based gateway that routes LLM API requests to different providers (OpenAI, Anthropic, DeepSeek) based on virtual API keys.

## Features

- **Virtual Key Routing**: Maps virtual API keys to real provider API keys
- **Multi-Provider Support**: OpenAI, Anthropic (Claude), DeepSeek
- **Usage Quotas**: Global hourly request limits to prevent abuse
- **Comprehensive Logging**: Pretty-printed JSON logs with full request/response capture

## Configuration

### Environment Variables

- `KEYS_FILE`: Path to virtual keys configuration file (default: `keys.json`)
- `GATEWAY_PORT`: Port to run the gateway server on (default: `8080`)
- `MAX_REQUESTS_PER_HOUR`: Per-key hourly request quota (default: `100`)
- `REQUEST_TIMEOUT`: Timeout for requests to LLM providers (default: `30s`). Use Go duration format (e.g., `30s`, `1m`, `500ms`)

### Virtual Keys Configuration

Create a `keys.json` in the project folder with your virtual key mappings:

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

# With custom configuration
MAX_REQUESTS_PER_HOUR=500  GATEWAY_PORT=3000  KEYS_FILE=/path/to/custom-keys.json go run main.go
```

### Building the Binary

```bash
go build -o gateway
```

### Making Requests

Send requests to `/chat/completions` with a virtual API key in the Authorization header:

#### Example: OpenAI Request
```bash
curl -X POST http://localhost:8080/chat/completions \
  -H "Authorization: Bearer vk_user1_openai" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```
## Testing with Example Script

```bash
cd examples
pip install -r requirements.txt
# Test router validation (asks providers to identify themselves)
python test_router.py
```