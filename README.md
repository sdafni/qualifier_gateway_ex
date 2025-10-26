# LLM Gateway Router

A Go-based gateway that routes LLM API requests to different providers (OpenAI, Anthropic, DeepSeek) based on virtual API keys.

## Features

- **Virtual Key Routing**: Maps virtual API keys to real provider API keys
- **Multi-Provider Support**: OpenAI, Anthropic (Claude), DeepSeek
- **Usage Quotas**: Global hourly request limits to prevent abuse
- **Comprehensive Logging**: Pretty-printed JSON logs with full request/response capture


## Limitations:
  No HTTPS Support
  No Request Size Limits
  Usage tracking is basic and inaccurate

## Configuration

### Environment Variables

- `KEYS_FILE`: Path to virtual keys configuration file (default: `keys.json`)
- `GATEWAY_PORT`: Port to run the gateway server on (default: `8080`)
- `MAX_REQUESTS_PER_HOUR`: Per-key hourly request quota (default: `100`)
- `REQUEST_TIMEOUT`: Timeout for requests to LLM providers (default: `30s`). Use Go duration format (e.g., `30s`, `1m`, `500ms`)

### Virtual Keys Configuration

Create a `keys.json` file with your virtual key mappings:

```bash
# Copy the example template
cp keys.json.example keys.json

# Edit with your real API keys
nano keys.json  # or use your preferred editor
```

Example `keys.json` structure:

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

### Run directly

#### Build the Binary

```bash
go build -o gateway
```

#### Run

```bash
# With default configuration (uses keys.json)
go run main.go

# Or with custom configuration
MAX_REQUESTS_PER_HOUR=500  GATEWAY_PORT=3000  KEYS_FILE=/path/to/custom-keys.json go run main.go
```




### Docker

#### Using Docker Compose (Recommended)
```bash
# Build the image
docker build -t llm-gateway .

```bash
# 1. Create your keys.json file (see Virtual Keys Configuration above)
cp keys.json.example keys.json
# Edit keys.json with your real API keys

# 2. Start the gateway (uses ./keys.json by default)
docker-compose up -d

# Or specify a custom keys file location:
KEYS_FILE_PATH=/path/to/your/keys.json docker-compose up -d

# Customize other settings:
KEYS_FILE_PATH=/path/to/keys.json \
MAX_REQUESTS_PER_HOUR=500 \
REQUEST_TIMEOUT=60s \
GATEWAY_PORT=3000 \
docker-compose up -d

```

#### Using Docker directly

```bash
# Build the image
docker build -t llm-gateway .

docker run -d \
  -p 8080:8080 \
  -v /path/to/your/keys.json:/app/keys.json:ro \
  -v $(pwd)/logs:/app/logs \
  -e MAX_REQUESTS_PER_HOUR=100 \
  -e REQUEST_TIMEOUT=30s \
  --name llm-gateway \
  llm-gateway

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