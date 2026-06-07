import { readdir, readFile } from "node:fs/promises";
import { join } from "node:path";
import { normalizeDomainFile, type RegistryEntry } from "./schema";
import { validateDomainRecord, validateSubdomain } from "./validate";

export async function loadRegistry(
  dir: string,
  baseDomain: string,
): Promise<RegistryEntry[]> {
  const files = await readdir(dir).catch(() => []);

  const entries: RegistryEntry[] = [];
  const seen = new Set<string>();

  for (const file of files) {
    if (!file.endsWith(".json")) continue;

    const name = file.replace(/\.json$/, "");
    const path = join(dir, file);

    const subdomainError = validateSubdomain(name);
    if (subdomainError) {
      throw new Error(`${file}: ${subdomainError}`);
    }

    if (seen.has(name)) {
      throw new Error(`${file}: duplicate domain`);
    }

    seen.add(name);

    const raw = await readFile(path, "utf8");
    const json = JSON.parse(raw);
    const parsed = normalizeDomainFile(name, json);

    for (const record of parsed.records) {
      const recordError = validateDomainRecord(record);

      if (recordError) {
        throw new Error(`${file}: ${recordError}`);
      }

      if (record.name !== name) {
        throw new Error(`${file}: record name must match filename`);
      }
    }

    entries.push({
      name,
      domain: `${name}.${baseDomain}`,
      file: parsed,
    });
  }

  return entries;
}

export async function findDomain(
  dir: string,
  baseDomain: string,
  name: string,
): Promise<RegistryEntry | null> {
  const registry = await loadRegistry(dir, baseDomain);
  return registry.find((entry) => entry.name === name) ?? null;
}
