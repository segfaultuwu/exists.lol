type DnsRecordType = "A" | "AAAA" | "CNAME" | "TXT" | "MX";

type CloudflareRecord = {
  id: string;
  type: DnsRecordType;
  name: string;
  content: string;
  proxied?: boolean;
  ttl: number;
  priority?: number;
};

type CloudflareResponse<T> = {
  success: boolean;
  errors: Array<{
    code: number;
    message: string;
  }>;
  messages: unknown[];
  result: T;
};

type RecordInput = {
  subdomain: string;
  type: DnsRecordType;
  content: string;
  proxied?: boolean;
  ttl?: number;
  priority?: number;
};

const CF_API = "https://api.cloudflare.com/client/v4";

function canBeProxied(type: DnsRecordType) {
  return type === "A" || type === "AAAA" || type === "CNAME";
}

function buildRecordBody(input: {
  type: DnsRecordType;
  name: string;
  content: string;
  proxied?: boolean;
  ttl?: number;
  priority?: number;
}) {
  const body: Record<string, unknown> = {
    type: input.type,
    name: input.name,
    content: input.content,
    ttl: input.ttl ?? 1,
  };

  if (canBeProxied(input.type)) {
    body.proxied = input.proxied ?? false;
  }

  if (input.type === "MX") {
    body.priority = input.priority ?? 10;
  }

  return body;
}

export class CloudflareDNS {
  constructor(
    private readonly token: string,
    private readonly zoneId: string,
    private readonly rootDomain: string,
  ) {
    if (!token) {
      throw new Error("Missing Cloudflare API token");
    }

    if (!zoneId) {
      throw new Error("Missing Cloudflare zone ID");
    }

    if (!rootDomain) {
      throw new Error("Missing root domain");
    }
  }

  private async request<T>(
    path: string,
    options: RequestInit = {},
  ): Promise<T> {
    const res = await fetch(`${CF_API}${path}`, {
      ...options,
      headers: {
        Authorization: `Bearer ${this.token}`,
        "Content-Type": "application/json",
        ...options.headers,
      },
    });

    const data = (await res.json()) as CloudflareResponse<T>;

    if (!res.ok || !data.success) {
      const message =
        data.errors?.map((e) => e.message).join(", ") || res.statusText;

      throw new Error(`Cloudflare API error: ${message}`);
    }

    return data.result;
  }

  fqdn(subdomain: string): string {
    return `${subdomain}.${this.rootDomain}`;
  }

  async listByName(name: string): Promise<CloudflareRecord[]> {
    const qs = new URLSearchParams({
      name,
      per_page: "100",
    });

    return await this.request<CloudflareRecord[]>(
      `/zones/${this.zoneId}/dns_records?${qs.toString()}`,
    );
  }

  async createRecord(input: RecordInput): Promise<CloudflareRecord> {
    const name = this.fqdn(input.subdomain);

    return await this.request<CloudflareRecord>(
      `/zones/${this.zoneId}/dns_records`,
      {
        method: "POST",
        body: JSON.stringify(
          buildRecordBody({
            type: input.type,
            name,
            content: input.content,
            ttl: input.ttl,
            proxied: input.proxied,
            priority: input.priority,
          }),
        ),
      },
    );
  }

  async updateRecord(
    recordId: string,
    input: RecordInput,
  ): Promise<CloudflareRecord> {
    const name = this.fqdn(input.subdomain);

    return await this.request<CloudflareRecord>(
      `/zones/${this.zoneId}/dns_records/${recordId}`,
      {
        method: "PATCH",
        body: JSON.stringify(
          buildRecordBody({
            type: input.type,
            name,
            content: input.content,
            ttl: input.ttl,
            proxied: input.proxied,
            priority: input.priority,
          }),
        ),
      },
    );
  }

  async deleteRecord(recordId: string): Promise<void> {
    await this.request(`/zones/${this.zoneId}/dns_records/${recordId}`, {
      method: "DELETE",
    });
  }

  async upsertRecord(input: RecordInput): Promise<CloudflareRecord> {
    const name = this.fqdn(input.subdomain);
    const existing = await this.listByName(name);

    const hasCname = existing.some((record) => record.type === "CNAME");
    const hasAddressRecord = existing.some(
      (record) => record.type === "A" || record.type === "AAAA",
    );

    if (input.type === "CNAME" && existing.length > 0) {
      throw new Error(
        `Cannot create CNAME because ${name} already has DNS records`,
      );
    }

    if ((input.type === "A" || input.type === "AAAA") && hasCname) {
      throw new Error(
        `Cannot create ${input.type} because ${name} already has CNAME`,
      );
    }

    if (input.type === "CNAME" && hasAddressRecord) {
      throw new Error(
        `Cannot create CNAME because ${name} already has A/AAAA records`,
      );
    }

    const sameRecord = existing.find((record) => {
      return record.type === input.type && record.content === input.content;
    });

    if (sameRecord) {
      return await this.updateRecord(sameRecord.id, input);
    }

    const sameType = existing.find((record) => record.type === input.type);

    if (
      sameType &&
      (input.type === "A" || input.type === "AAAA" || input.type === "CNAME")
    ) {
      return await this.updateRecord(sameType.id, input);
    }

    return await this.createRecord(input);
  }
}
