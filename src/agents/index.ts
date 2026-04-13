import { GeneralAgent } from "./general.ts";
import { ThinkingAgent } from "./thinking.ts";
import type {
  MenreikiAgentConfig,
  MenreikiAgentModeType,
} from "../types/agents.ts";

const {
  OPENAI_CHAT_MODEL: chatModel,
  OPENAI_API_KEY: apiKey,
  OPENAI_BASE_URL: baseURL,
  AGENT_MODE_TYPE: agentModeType,
  AGENT_SYSTEM_PROMPT: agentSystemPrompt,
} = process.env;

const config: MenreikiAgentConfig = {
  model: chatModel || "gpt-4o",
  apiKey: apiKey || "",
  baseURL: baseURL,
  systemPrompt: agentSystemPrompt || "You are a helpful assistant.",
};

const generalAgent = new GeneralAgent(config);
const thinkingAgent = new ThinkingAgent(config);

/**
 * Fetch an agent by type.
 * @param agentType - The type of agent to fetch
 * @returns The agent
 */
export function fetchAgent(agentType: MenreikiAgentModeType | null = null) {
  if (agentType === null) {
    agentType = agentModeType as MenreikiAgentModeType;
  }
  switch (agentType) {
    case "general":
      return generalAgent;
    case "thinking":
      return thinkingAgent;
    default:
      throw new Error(`Unknown agent type: ${agentType}`);
  }
};

export * from "../types/agents.ts";
export { GeneralAgent } from "./general.ts";
export { ThinkingAgent } from "./thinking.ts";
