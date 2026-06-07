import { serve } from "bun";
import { api } from "./api";
import { config } from "./config";

serve({
  hostname: config.host,
  port: config.port,
  fetch: api.fetch,
});

console.log(`exists.lol registry running on http://localhost:${config.port}`);
