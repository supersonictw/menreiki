import OpenAI from "openai";
import type { MenreikiAgent } from "../types/agents.ts";

const {
  OPENAI_BASE_URL: baseURL,
  OPENAI_API_KEY: apiKey,
  OPENAI_CHAT_MODEL: chatModel,
} = process.env as {
    OPENAI_BASE_URL?: string;
    OPENAI_API_KEY?: string;
    OPENAI_CHAT_MODEL?: string;
};

/**
 * General AI agent based on OpenAI Chat Completion API.
 */
export class GeneralAgent implements MenreikiAgent {
  private readonly client: OpenAI;
  private readonly chatHistoryMapper: Map<string, Array<OpenAI.Chat.ChatCompletionMessageParam>>;
  private readonly prependPrompts: Array<OpenAI.Chat.ChatCompletionMessageParam> = [];

  /**
   * Initialize the GeneralAgent.
   */
  constructor() {
    this.client = new OpenAI({ baseURL, apiKey });
    this.chatHistoryMapper = new Map<string, Array<OpenAI.Chat.ChatCompletionMessageParam>>();
  }

  /**
   * Randomly choose an element from an array.
   * @template T
   * @param choices - The array of choices
   * @returns The randomly chosen element
   */
  private choose<T>(choices: T[]): T {
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
  public async chatWithAI(
    chatId: string,
    prompt: string,
  ): Promise<string> {
    if (!this.chatHistoryMapper.has(chatId)) {
      this.chatHistoryMapper.set(chatId, []);
    }

    const chatHistory = this.chatHistoryMapper.get(chatId)!;
    const userPromptMessage: OpenAI.Chat.ChatCompletionUserMessageParam = {
      role: "user",
      content: prompt,
    };

    const response = await this.client.chat.completions.create({
      model: chatModel as string,
      messages: [
        ...this.prependPrompts,
        ...chatHistory,
        userPromptMessage,
      ],
    });

    const choice = this.choose(response.choices);
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
  public useClient(): OpenAI {
    return this.client;
  }
}
