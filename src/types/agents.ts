export interface MenreikiAgent {
  chatWithAI(chatId: string, prompt: string): Promise<string>;
}
