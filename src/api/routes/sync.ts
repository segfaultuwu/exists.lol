import { Hono } from "hono";
import { config } from "../../config";
import { CloudflareDNS } from "../../dns/cloudflare";
import { loadRegistry } from "../../registry/loader";
import { requireAdmin } from "../middleware/requireAdmin";

export const syncRoutes = new Hono();

syncRoutes.post("/sync/cloudflare", requireAdmin, async (c) => {
  const registry = await loadRegistry(config.registryDir, config.baseDomain);

  const cf = new CloudflareDNS(
    config.cloudflare.apiToken,
    config.cloudflare.zoneId,
    config.baseDomain,
  );

  const results = [];

  for (const entry of registry) {
    for (const record of entry.file.records) {
      try {
        const cfRecord = await cf.upsertRecord({
          subdomain: entry.name,
          type: record.type,
          content: record.value,
          ttl: 1,
          proxied: false,
        });

        results.push({
          ok: true,
          domain: entry.domain,
          type: cfRecord.type,
          name: cfRecord.name,
          content: cfRecord.content,
          cloudflareRecordId: cfRecord.id,
        });
      } catch (err) {
        results.push({
          ok: false,
          domain: entry.domain,
          type: record.type,
          value: record.value,
          error: err instanceof Error ? err.message : String(err),
        });
      }
    }
  }

  return c.json({
    ok: true,
    count: results.length,
    results,
  });
});

syncRoutes.post("/sync/cloudflare/dry-run", requireAdmin, async (c) => {
  const registry = await loadRegistry(config.registryDir, config.baseDomain);

  return c.json({
    ok: true,
    count: registry.length,
    changes: registry.flatMap((entry) =>
      entry.file.records.map((record) => ({
        action: "upsert",
        domain: entry.domain,
        type: record.type,
        value: record.value,
      })),
    ),
  });
});
