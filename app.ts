import "dotenv/config";
import {
  setupTriggers,
} from "./src/triggers/index.ts";

setupTriggers();

const timestamp = new Date().toString();
console.info(`menreiki: "${timestamp}"`);
