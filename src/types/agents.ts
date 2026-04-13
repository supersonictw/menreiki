export type MenreikiAgentConfig = {
  model: string;
  apiKey: string;
  baseURL?: string;
  temperature?: number;
  systemPrompt: string;
};

export interface MenreikiAgent {
  chatWithAI(chatId: string, prompt: string): Promise<string>;
}
