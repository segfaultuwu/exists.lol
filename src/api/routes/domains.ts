import { Hono } from "hono";
import { z } from "zod";
import { config } from "../../config";
import { CloudflareDNS } from "../../dns/cloudflare";
import { findDomain, loadRegistry } from "../../registry/loader";
import { RecordTypeSchema } from "../../registry/schema";
import { requireAdmin } from "../middleware/requireAdmin";

export const domainsRoutes = new Hono();

export const ValidateSchema = z.object({
  name: z.string(),
  type: RecordTypeSchema,
  value: z.string(),
});

domainsRoutes.get("/domains", async (c) => {
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

domainsRoutes.get("/domains/:name", async (c) => {
  const name = c.req.param("name");

  const entry = await findDomain(config.registryDir, config.baseDomain, name);

  if (!entry) {
    return c.json(
      {
        ok: false,
        error: "domain not found",
      },
      404,
    );
  }

  return c.json({
    ok: true,
    name: entry.name,
    domain: entry.domain,
    owner: entry.file.owner,
    records: entry.file.records,
  });
});

domainsRoutes.post("/domains/:name/publish", requireAdmin, async (c) => {
  const name = c.req.param("name") as string;

  const entry = await findDomain(config.registryDir, config.baseDomain, name);

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

domainsRoutes.delete("/domains/:name/cloudflare", requireAdmin, async (c) => {
  const name = c.req.param("name");

  const cf = new CloudflareDNS(
    config.cloudflare.apiToken,
    config.cloudflare.zoneId,
    config.baseDomain,
  );

  const fqdn = `${name}.${config.baseDomain}`;
  const records = await cf.listByName(fqdn);

  for (const record of records) {
    await cf.deleteRecord(record.id);
  }

  return c.json({
    ok: true,
    domain: fqdn,
    deleted: records.length,
  });
});
