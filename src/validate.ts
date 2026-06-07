import { config } from "./config";
import { loadRegistry } from "./registry/loader";

try {
  const registry = await loadRegistry(config.registryDir, config.baseDomain);

  console.log(`✅ Registry valid`);
  console.log(`Loaded domains: ${registry.length}`);

  for (const entry of registry) {
    console.log(`- ${entry.domain}`);
  }
} catch (err) {
  console.error("❌ Registry invalid");

  if (err instanceof Error) {
    console.error(err.message);
  } else {
    console.error(err);
  }

  process.exit(1);
}
