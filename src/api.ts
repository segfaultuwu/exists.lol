import { Context, Hono, type Next } from "hono";
import { z } from "zod";
import { config } from "./config";
import { findDomain, loadRegistry } from "./registry/loader";
import { RecordTypeSchema } from "./registry/schema";
import { CloudflareDNS } from "./dns/cloudflare";

export const api = new Hono();

export const ValidateSchema = z.object({
  name: z.string(),
  type: RecordTypeSchema,
  value: z.string(),
});

async function requireAdmin(c: Context, next: Next) {
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

api.get("/", (c) => {
  return c.json({
    name: "exists.lol registry",
    ok: true,
  });
});

api.get("/api/health", (c) => {
  return c.json({
    ok: true,
  });
});

api.get("/api/domains", async (c) => {
  try {
    console.log("GET /api/domains");

    const registry = await loadRegistry(config.registryDir, config.baseDomain);

    console.log("loaded domains:", registry.length);

    return c.json({
      ok: true,
      count: registry.length,
      domains: registry.map((entry) => ({
        name: entry.name,
        domain: entry.domain,
        owner: entry.file.owner,
        records: entry.file.records,
      })),
    });
  } catch (err) {
    console.error("GET /api/domains ERROR:", err);

    return c.json(
      {
        ok: false,
        error: err instanceof Error ? err.message : String(err),
      },
      500,
    );
  }
});

api.get("/api/domains/:name", async (c) => {
  const name = c.req.param("name");
  const entry = await findDomain(config.registryDir, config.baseDomain, name);

  if (!entry) {
    return c.json(
      {
        error: "domain not found",
      },
      404,
    );
  }

  return c.json({
    name: entry.name,
    domain: entry.domain,
    owner: entry.file.owner,
    records: entry.file.records,
  });
});

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

api.post("/api/domains/:name/publish", requireAdmin, async (c) => {
  const name = c.req.param("name");

  const entry = await findDomain(
    config.registryDir,
    config.baseDomain,
    name as string,
  );

  if (!entry) {
    return c.json(
      {
        ok: false,
        error: "domain not found",
      },
      404,
    );
  }

  const cf = new CloudflareDNS(
    config.cloudflare.apiToken,
    config.cloudflare.zoneId,
    config.baseDomain,
  );

  const published = [];

  for (const record of entry.file.records) {
    const cfRecord = await cf.upsertRecord({
      subdomain: entry.name,
      type: record.type,
      content: record.value,
      ttl: 1,
      proxied: false,
    });

    published.push({
      id: cfRecord.id,
      type: cfRecord.type,
      name: cfRecord.name,
      content: cfRecord.content,
    });
  }

  return c.json({
    ok: true,
    domain: entry.domain,
    published,
  });
});
