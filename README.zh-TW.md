# Menreiki 面霊氣

> 咲き狂え彼岸花 染まる色に嗤え

![avatar](avatar.min.png)

Menreiki 是一個基於多代理架構（multi-agent architecture）的 Discord AI 代理（agent）。它不僅僅是一個簡單的聊天機器人，還利用推理圖（LangGraph）進行計畫、檢索，並透過深思熟慮的工具輔助思考來提供回覆。您提供大語言模型（LLM），Menreiki 則提供架構。

## 特色

- 由 LangGraph 驅動的多代理推理圖，支援思考鏈（Chain of Thought, CoT）
- 可插拔的工具生態系統：透過 Tavily 進行網路搜尋、系統時間工具等
- 持久的每頻道聊天紀錄，具備自動上下文修剪功能
- 支援伺服器（Guild）頻道與私訊（Direct Messages）
- 相容於任何符合 OpenAI 規範的 API 端點（OpenAI, Azure, 本地模型等）
- 可透過 PM2 叢集模式進行水平擴充，並已容器化以便部署

## 架構概覽

每條傳入的 Discord 訊息都會經由觸發層路由到 `ThinkingAgent` 實例。該代理運行一個有狀態的 LangGraph 工作流，在模型推理與工具執行之間切換，直到模型發出完成訊號或達到步數/超時限制。聊天紀錄按頻道維護於記憶體中，並使用滾動視窗來管理上下文長度。

```
Discord 訊息
    └─ 觸發層
         └─ ThinkingAgent (LangGraph)
              ├─ 代理節點 (agent node)  →  LLM (OpenAI 相容)
              └─ 工具節點 (tools node)  →  Tavily / Time / ...
```

## 開始使用

### 1. 建立 Discord 應用程式

1. 前往 [Discord Developer Portal](https://discord.com/developers/applications)。
2. 點擊 **New Application**，為其命名並儲存 **Application ID**。
3. 開啟 **Bot** 分頁，點擊 **Reset Token** 以取得您的 **Bot Token**。
4. 在 **Privileged Gateway Intents** 下，啟用 **Message Content Intent**。
5. 前往 **OAuth2 > URL Generator**，勾選 `bot` 與 `applications.commands` 範圍（Scopes）。
   新增機器人權限：**Send Messages**、**Read Message History**、**Use Slash Commands**。
6. 在瀏覽器中開啟生成的 URL，將機器人邀請至您的伺服器。

### 2. 準備環境變數

| 變數 | 描述 |
|---|---|
| `OPENAI_BASE_URL` | 符合 OpenAI 規範的 API 基礎網址 |
| `OPENAI_API_KEY` | LLM 提供者的 API 金鑰 |
| `OPENAI_CHAT_MODEL` | 使用的模型名稱（例如 `gpt-4o`） |
| `DISCORD_APP_ID` | 來自開發者門戶的應用程式 ID |
| `DISCORD_BOT_TOKEN` | 來自開發者門戶的機器人權鑰（Token） |

### 3. 使用 Docker 執行

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

## 授權條款

本軟體採用 [MIT 授權條款](LICENSE) 進行授權。

> (c) 2025 [不知火 Shiranui](https://shiranui.xyz).
