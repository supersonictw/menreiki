# Menreiki 面霊気

[正體中文](README.zh-TW.md) | English

> 咲き狂え彼岸花 染まる色に嗤え

![avatar](avatar.min.png)

Menreiki is a Discord AI agent built on a multi-agent architecture. It goes beyond simple chatbots by using a reasoning graph (LangGraph) to plan, retrieve, and respond with deliberate, tool-assisted thinking. You bring the LLM. Menreiki brings the structure.

## Features

- Multi-agent reasoning graph powered by LangGraph with Chain of Thought (CoT) support
- Pluggable tool ecosystem: web search via Tavily, system time utilities, and more
- Persistent per-channel chat history with automatic context trimming
- Works in both guild channels and direct messages
- Compatible with any OpenAI-compatible API endpoint (OpenAI, Azure, local models, etc.)
- Horizontally scalable via PM2 cluster mode, containerized for easy deployment

## Architecture Overview

Each incoming Discord message is routed through a trigger layer to a `ThinkingAgent` instance. The agent runs a stateful LangGraph workflow that alternates between model inference and tool execution until the model signals completion or a step/timeout limit is reached. Chat history is maintained in-memory per channel, with a rolling window to manage context length.

```
Discord Message
    └─ Trigger Layer
         └─ ThinkingAgent (LangGraph)
              ├─ agent node  →  LLM (OpenAI-compatible)
              └─ tools node  →  Tavily / Time / ...
```

## Getting Started

### 1. Create a Discord Application

1. Go to the [Discord Developer Portal](https://discord.com/developers/applications).
2. Click **New Application**, give it a name, and save the **Application ID**.
3. Open the **Bot** tab and click **Reset Token** to get your **Bot Token**.
4. Under **Privileged Gateway Intents**, enable **Message Content Intent**.
5. Go to **OAuth2 > URL Generator**, select scopes `bot` and `applications.commands`.
   Add bot permissions: **Send Messages**, **Read Message History**, **Use Slash Commands**.
6. Open the generated URL in your browser to invite the bot to your server.

### 2. Prepare Environment Variables

| Variable | Description |
|---|---|
| `OPENAI_BASE_URL` | Base URL of your OpenAI-compatible API |
| `OPENAI_API_KEY` | API key for the LLM provider |
| `OPENAI_CHAT_MODEL` | Model name to use (e.g. `gpt-4o`) |
| `DISCORD_APP_ID` | Application ID from the Developer Portal |
| `DISCORD_BOT_TOKEN` | Bot token from the Developer Portal |

### 3. Run with Docker

```bash
docker pull ghcr.io/supersonictw/menreiki:main

docker run -d \
  --name menreiki \
  --restart unless-stopped \
  -e OPENAI_BASE_URL="https://api.openai.com/v1" \
  -e OPENAI_API_KEY="your_api_key" \
  -e OPENAI_CHAT_MODEL="gpt-4o" \
  -e DISCORD_APP_ID="your_app_id" \
  -e DISCORD_BOT_TOKEN="your_bot_token" \
  ghcr.io/supersonictw/menreiki:main
```

## License

The software is licensed under [MIT License](LICENSE).

> (c) 2025 [Shiranui](https://shiranui.xyz).
