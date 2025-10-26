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

# With custom quota (e.g., 500 requests per hour)
MAX_REQUESTS_PER_HOUR=500 go run main.go
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

## Usage Quotas

The gateway implements per-virtual-key hourly request quotas to prevent abuse:

- **Default**: 100 requests per hour per virtual key
- **Configurable**: Set via `MAX_REQUESTS_PER_HOUR` environment variable
- **Scope**: Per virtual key (each key has its own independent counter)
- **Window**: Hourly (resets independently per key)

### Quota Enforcement

When the quota is exceeded, the gateway returns:
- **Status Code**: `429 Too Many Requests`
- **Error Message**: `"quota exceeded: <N> requests per hour limit reached"`

Example error response:
```bash
curl -X POST http://localhost:8080/chat/completions \
  -H "Authorization: Bearer vk_user1_openai" \
  -H "Content-Type: application/json" \
  -d '{"model": "gpt-4", "messages": [{"role": "user", "content": "Hello!"}]}'

# Response (when quota exceeded):
# HTTP/1.1 429 Too Many Requests
# quota exceeded: 100 requests per hour limit reached
```

Each virtual key's quota counter resets automatically every hour from its first use.

### Timeout Handling

The gateway implements request timeouts to prevent hanging on slow or unresponsive providers:

- **Default Timeout**: 30 seconds
- **Configurable**: Set via `REQUEST_TIMEOUT` environment variable (e.g., `REQUEST_TIMEOUT=60s`)
- **Quota Behavior**: Failed requests (including timeouts) do NOT consume quota
- **Only successful responses** (where the gateway receives a complete response from the provider) count toward the quota limit

Example timeout scenario:
```bash
# If a provider takes longer than the configured timeout:
# 1. Request is terminated
# 2. Gateway returns 502 Bad Gateway
# 3. Quota is NOT incremented
# 4. User can retry without losing quota
```

## Testing with Example Script

```bash
cd examples
pip install -r requirements.txt
# Test router validation (asks providers to identify themselves)
python test_router.py
```