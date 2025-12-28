import {
  Message,
} from "discord.js";

import {
  client,
} from "../../clients/discord.ts";
import {
  chatWithAI,
} from "../../clients/openai.ts";

/**
 * Message create event handler.
 * @param message - The message object
 */
export default async function messageCreate(message: Message) {
  if (
    message.author.id === client.user?.id ||
    (message.guild && !message.mentions.users.has(client.user?.id ?? ""))
  ) {
    return;
  }

  if ("sendTyping" in message.channel) {
    await message.channel.sendTyping?.();
  }

  const requestContent = message.content;
  if (!requestContent) {
    await message.reply("Sorry, I couldn't understand your message. Please try rephrasing it.");
    return;
  }

  const responseContent = await chatWithAI(message.channel.id, requestContent);
  if (!responseContent) {
    await message.reply("Sorry, I seem to be having trouble thinking. Please try saying it in a different way.");
    return;
  }

  await message.reply(responseContent);
}
