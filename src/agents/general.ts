import OpenAI from "openai";

const {
  OPENAI_BASE_URL: baseURL,
  OPENAI_API_KEY: apiKey,
  OPENAI_CHAT_MODEL: chatModel,
} = process.env as {
    OPENAI_BASE_URL?: string;
    OPENAI_API_KEY?: string;
    OPENAI_CHAT_MODEL?: string;
};

const prependPrompts: Array<OpenAI.Chat.ChatCompletionMessageParam> = [];

const client = new OpenAI({ baseURL, apiKey });
const chatHistoryMapper = new Map<string, Array<OpenAI.Chat.ChatCompletionMessageParam>>();

/**
 * Randomly choose an element from an array.
 * @template T
 * @param choices - The array of choices
 * @returns The randomly chosen element
 */
function choose<T>(choices: T[]): T {
  const seed = Math.random();
  const index = Math.floor(seed * choices.length);
  return choices[index];
}

/**
 * Chat with the AI.
 * @param chatId - The chat ID to chat with the AI
 * @param prompt - The prompt to chat with the AI
 * @returns The response from the AI
 */
export async function chatWithAI(
  chatId: string,
  prompt: string,
): Promise<string> {
  if (!chatHistoryMapper.has(chatId)) {
    chatHistoryMapper.set(chatId, []);
  }

  const chatHistory = chatHistoryMapper.get(chatId)!;
  const userPromptMessage: OpenAI.Chat.ChatCompletionUserMessageParam = {
    role: "user",
    content: prompt,
  };

  const response = await client.chat.completions.create({
    model: chatModel as string,
    messages: [
      ...prependPrompts,
      ...chatHistory,
      userPromptMessage,
    ],
  });

  const choice = choose(response.choices);
  const reply = choice.message.content ?? "";
  const assistantMsg: OpenAI.Chat.ChatCompletionAssistantMessageParam = {
    role: "assistant",
    content: reply,
  };

  chatHistory.push(
    userPromptMessage,
    assistantMsg,
  );
  if (chatHistory.length > 30) {
    chatHistory.shift();
    chatHistory.shift();
  }

  return reply;
}

/**
 * Get the OpenAI client instance.
 * @returns The OpenAI client
 */
export function useClient(): OpenAI {
  return client;
}
