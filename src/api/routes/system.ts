import { Hono } from "hono";
import { requireAdmin } from "../middleware/requireAdmin";

export const systemRoutes = new Hono();

systemRoutes.get("/health", (c) => {
  return c.json({
    ok: true,
  });
});

systemRoutes.post("/update", requireAdmin, async (c) => {
  const git = Bun.spawn(["git", "pull"]);
  const code = await git.exited;

  return c.json({
    ok: code === 0,
  });
});
