import { z } from "zod";

export const RecordTypeSchema = z.enum(["A", "AAAA", "CNAME", "TXT"]);

const RecordValuesSchema = z.union([z.string(), z.array(z.string())]);

export const RawDomainFileSchema = z.object({
  owner: z.object({
    username: z.string().min(1),
    github_username: z.string().min(1),
    discord_id: z.string().min(1),
  }),
  records: z.object({
    A: RecordValuesSchema.optional(),
    AAAA: RecordValuesSchema.optional(),
    CNAME: RecordValuesSchema.optional(),
    TXT: RecordValuesSchema.optional(),
  }),
});

export type RecordType = z.infer<typeof RecordTypeSchema>;

export type DomainRecord = {
  type: RecordType;
  name: string;
  value: string;
};

export type DomainFile = {
  owner: {
    username: string;
    githubUsername: string;
    discordId: string;
  };
  records: DomainRecord[];
};

export type RegistryEntry = {
  name: string;
  domain: string;
  file: DomainFile;
};

export function normalizeDomainFile(name: string, raw: unknown): DomainFile {
  const parsed = RawDomainFileSchema.parse(raw);

  const records: DomainRecord[] = [];

  for (const [type, values] of Object.entries(parsed.records)) {
    if (!values) continue;

    const list = Array.isArray(values) ? values : [values];

    for (const value of list) {
      records.push({
        type: type as RecordType,
        name,
        value,
      });
    }
  }

  return {
    owner: {
      username: parsed.owner.username,
      githubUsername: parsed.owner.github_username,
      discordId: parsed.owner.discord_id,
    },
    records,
  };
}
