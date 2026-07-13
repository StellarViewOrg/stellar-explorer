import { truncateHash } from "./format";

type EntityType = "transaction" | "account" | "contract" | "asset" | "ledger";

interface AgentContextData {
  type: EntityType;
  network: string;
  data: Record<string, unknown>;
  baseUrl?: string;
}

/**
 * Generates a concise, LLM-friendly text summary of a Stellar entity.
 * Designed to be copied into an AI assistant context window.
 */
export function generateAgentContext({
  type,
  network,
  data,
  baseUrl = "",
}: AgentContextData): string {
  const net =
    network === "mainnet" ? "Mainnet" : network.charAt(0).toUpperCase() + network.slice(1);
  const lines: string[] = [];

  switch (type) {
    case "account": {
      const id = (data.id ?? data.account_id) as string;
      const balances = (data.balances as Array<Record<string, string>>) ?? [];
      const xlm = balances.find((b) => b.asset_type === "native");
      const otherAssets = balances.filter((b) => b.asset_type !== "native");

      lines.push(`STELLAR ACCOUNT: ${id}`);
      lines.push(`Network: ${net} | Signers: ${(data.signers as unknown[])?.length ?? 1}`);
      if (xlm) lines.push(`XLM Balance: ${parseFloat(xlm.balance).toFixed(7)} XLM`);
      if (otherAssets.length > 0) {
        lines.push(
          `Other Assets (${otherAssets.length}): ${otherAssets.map((b) => `${b.balance} ${b.asset_code ?? "?"}`).join(", ")}`
        );
      }
      if (data.sequence) lines.push(`Sequence: ${data.sequence}`);
      if (data.home_domain) lines.push(`Home Domain: ${data.home_domain}`);
      break;
    }

    case "transaction": {
      const hash = data.hash as string;
      lines.push(`STELLAR TRANSACTION: ${hash}`);
      lines.push(`Network: ${net} | Status: ${data.successful ? "Success" : "Failed"}`);
      lines.push(`Ledger: ${data.ledger} | Created: ${data.created_at}`);
      lines.push(`Operations: ${data.operation_count} | Fee paid: ${data.fee_charged} stroops`);
      if (data.memo_type && data.memo_type !== "none" && data.memo) {
        lines.push(`Memo (${data.memo_type}): ${data.memo}`);
      }
      lines.push(`Source: ${data.source_account}`);
      break;
    }

    case "contract": {
      const id = data.id as string;
      lines.push(`SOROBAN CONTRACT: ${id}`);
      lines.push(`Network: ${net}`);
      if (data.type === "sac") {
        lines.push(`Type: Stellar Asset Contract (SAC) — wraps a native Stellar asset`);
      } else {
        lines.push(`Type: Custom WASM contract`);
        if (data.wasmHash) lines.push(`WASM Hash: ${data.wasmHash}`);
        if (data.codeSize) lines.push(`Code Size: ${data.codeSize} bytes`);
      }
      if (data.balance && parseFloat(data.balance as string) > 0) {
        lines.push(`XLM Balance: ${data.balance} XLM`);
      }
      break;
    }

    case "asset": {
      const code = data.asset_code as string;
      const issuer = data.asset_issuer as string;
      lines.push(`STELLAR ASSET: ${code}`);
      lines.push(`Network: ${net} | Issuer: ${issuer}`);
      const accounts = data.accounts as Record<string, number> | undefined;
      if (accounts) {
        const total =
          (accounts.authorized ?? 0) + (accounts.authorized_to_maintain_liabilities ?? 0);
        lines.push(`Holders: ${total.toLocaleString()}`);
      }
      const balances = data.balances as Record<string, string> | undefined;
      if (balances?.authorized) {
        lines.push(`Total Supply: ${parseFloat(balances.authorized).toLocaleString()} ${code}`);
      }
      if (data.flags) {
        const flags = data.flags as Record<string, boolean>;
        const activeFlags = Object.entries(flags)
          .filter(([, v]) => v)
          .map(([k]) => k);
        if (activeFlags.length > 0) lines.push(`Flags: ${activeFlags.join(", ")}`);
      }
      break;
    }

    case "ledger": {
      lines.push(`STELLAR LEDGER: #${data.sequence}`);
      lines.push(`Network: ${net} | Closed: ${data.closed_at}`);
      lines.push(
        `Transactions: ${data.successful_transaction_count} success, ${data.failed_transaction_count} failed`
      );
      lines.push(`Operations: ${data.operation_count} | Fee pool: ${data.fee_pool} XLM`);
      lines.push(
        `Base fee: ${data.base_fee_in_stroops} stroops | Protocol: ${data.protocol_version}`
      );
      break;
    }
  }

  const entityId = (data.id ?? data.hash ?? data.sequence) as string;
  lines.push(`\nExplorer URL: ${baseUrl}/en/${network}/${type}/${entityId}`);
  lines.push(`Data as of: ${new Date().toISOString()}`);

  return lines.join("\n");
}

export function formatHashForDisplay(hash: string): string {
  return truncateHash(hash, 8, 8);
}
