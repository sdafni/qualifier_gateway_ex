# LLM Gateway Router

A Go-based gateway that routes LLM API requests to different providers (OpenAI, Anthropic, DeepSeek) based on virtual API keys.

## Features

- **Virtual Key Routing**: Maps virtual API keys to real provider API keys
- **Multi-Provider Support**: OpenAI, Anthropic (Claude), DeepSeek

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

## Testing with Example Script

```bash
cd examples
pip install -r requirements.txt
# Test router validation (asks providers to identify themselves)
python test_router.py
```