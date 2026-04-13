import { GeneralAgent } from "./general.ts";
import { ThinkingAgent } from "./thinking.ts";

const {
  OPENAI_CHAT_MODEL: chatModel,
  OPENAI_API_KEY: apiKey,
  OPENAI_BASE_URL: baseURL,
} = process.env;

export const generalAgent = new GeneralAgent();

export const thinkingAgent = new ThinkingAgent({
  model: chatModel || "gpt-4o",
  apiKey: apiKey || "",
  baseURL: baseURL,
  systemPrompt: "You are a helpful assistant with thinking capabilities.",
});

export * from "../types/agents.ts";
export { GeneralAgent } from "./general.ts";
export { ThinkingAgent } from "./thinking.ts";
