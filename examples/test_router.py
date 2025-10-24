#!/usr/bin/env python3
"""
Test script to validate that the gateway properly routes to different providers.
Asks each provider to identify themselves to confirm correct routing.
"""

import requests
from openai import OpenAI

GATEWAY_URL = "http://localhost:8080"

def test_deepseek():
    """Test DeepSeek routing using OpenAI-compatible client"""
    print("=" * 60)
    print("Testing DeepSeek Routing")
    print("=" * 60)

    # Configure OpenAI client to use DeepSeek through gateway
    client = OpenAI(
        api_key="vk_user3_deepseek",
        base_url=GATEWAY_URL,
    )

    # Ask DeepSeek to identify itself
    response = client.chat.completions.create(
        model="deepseek-chat",
        messages=[
            {"role": "user", "content": "In one sentence, who are you? What AI assistant/model are you?"}
        ]
    )

    answer = response.choices[0].message.content
    print(f"Virtual Key: vk_user3_deepseek")
    print(f"Expected Provider: DeepSeek")
    print(f"Response: {answer}")
    print()

    # Validation
    if "deepseek" in answer.lower():
        print("✓ SUCCESS: DeepSeek correctly identified itself")
    else:
        print("✗ WARNING: Response doesn't mention DeepSeek")

    return answer


def test_anthropic():
    """Test Anthropic routing using requests"""
    print("\n" + "=" * 60)
    print("Testing Anthropic Routing")
    print("=" * 60)

    headers = {
        "Authorization": "Bearer vk_user2_anthropic",
        "Content-Type": "application/json"
    }

    payload = {
        "model": "claude-3-5-sonnet-20241022",
        "messages": [
            {"role": "user", "content": "In one sentence, who are you? What AI assistant/model are you?"}
        ],
        "max_tokens": 1024
    }

    response = requests.post(
        f"{GATEWAY_URL}/chat/completions",
        headers=headers,
        json=payload,
        timeout=30
    )

    print(f"Virtual Key: vk_user2_anthropic")
    print(f"Expected Provider: Anthropic")
    print(f"Status Code: {response.status_code}")

    if response.status_code == 200:
        data = response.json()
        # Anthropic returns content as a list
        if "content" in data and isinstance(data["content"], list):
            answer = data["content"][0].get("text", "")
            print(f"Response: {answer}")
            print()

            # Validation
            if "claude" in answer.lower() or "anthropic" in answer.lower():
                print("✓ SUCCESS: Anthropic/Claude correctly identified itself")
            else:
                print("✗ WARNING: Response doesn't mention Claude or Anthropic")

            return answer
    else:
        print(f"✗ ERROR: Request failed - {response.text}")
        return None


if __name__ == "__main__":
    print("\n🔄 LLM Gateway Router Validation Test\n")

    try:
        # Test both providers
        test_deepseek()
        test_anthropic()

        print("\n" + "=" * 60)
        print("Test Complete!")
        print("=" * 60)

    except Exception as e:
        print(f"\n✗ ERROR: {e}")
        print("\nMake sure:")
        print("1. The gateway is running (go run main.go)")
        print("2. Real API keys are configured in keys.json")
        print("3. The requests library is installed (pip install requests)")
