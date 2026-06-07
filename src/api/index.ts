import { Hono } from "hono";
import { domainsRoutes } from "./routes/domains";
import { syncRoutes } from "./routes/sync";
import { systemRoutes } from "./routes/system";

export const api = new Hono();

api.get("/", (c) => {
  return c.json({
    name: "exists.lol registry",
    ok: true,
  });
});

api.route("/api", systemRoutes);
api.route("/api", domainsRoutes);
api.route("/api", syncRoutes);

api.onError((err, c) => {
  console.error("API ERROR:", err);

  return c.json(
    {
      ok: false,
      error: err.message,
    },
    500,
  );
});
