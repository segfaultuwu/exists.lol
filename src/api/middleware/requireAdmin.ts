import type { Context, Next } from "hono";
import { config } from "../../config";

export async function requireAdmin(c: Context, next: Next) {
  const auth = c.req.header("Authorization");

  if (auth !== `Bearer ${config.adminToken}`) {
    return c.json(
      {
        ok: false,
        error: "unauthorized",
      },
      401,
    );
  }

  await next();
}
