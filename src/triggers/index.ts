import {
  Events,
} from "discord.js";
import {
  client,
} from "../clients/discord.ts";

import messageCreateHandler from "./message_create/index.ts";

export const setupTriggers = (): void => {
  client.on(Events.MessageCreate, messageCreateHandler);
};
