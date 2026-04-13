import {
  type StructuredToolInterface,
} from "@langchain/core/tools";

import {
  currentTimeTool,
} from "./time.ts";
import {
  isEnabled as isTavilyEnabled,
  tavilySearchTool,
} from "./tavily.ts";

export const tools: StructuredToolInterface[] = [
  currentTimeTool,
];

if (isTavilyEnabled) {
  tools.push(tavilySearchTool);
}

export const toolMap: Record<
    string, StructuredToolInterface
> = Object.fromEntries(
  tools.map((toolImpl) => [toolImpl.name, toolImpl]),
);
