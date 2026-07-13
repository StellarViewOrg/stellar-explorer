import { truncateHash } from "./format";
import { formatAssetFromFields, parseSorobanFunctionType } from "./operation-helpers";

type Op = Record<string, unknown>;

function addr(value: unknown): string {
  if (typeof value !== "string") return "unknown";
  return truncateHash(value, 4, 4);
}

function asset(type?: unknown, code?: unknown, issuer?: unknown): string {
  const a = formatAssetFromFields(type as string, code as string, issuer as string);
  return a.code;
}

function amount(value: unknown): string {
  if (typeof value !== "string" && typeof value !== "number") return "";
  const num = parseFloat(String(value));
  if (isNaN(num)) return String(value);
  return num % 1 === 0
    ? num.toLocaleString()
    : num.toLocaleString(undefined, { maximumFractionDigits: 7 }).replace(/\.?0+$/, "");
}

/**
 * Returns a single human-readable sentence for a Horizon operation object.
 * Language is English; i18n strings can be added later.
 */
export function getOperationSummary(op: Op): string {
  const type = op.type as string;

  switch (type) {
    case "payment":
      return `Sent ${amount(op.amount)} ${asset(op.asset_type, op.asset_code, op.asset_issuer)} → ${addr(op.to)}`;

    case "create_account":
      return `Created account ${addr(op.account)} with ${amount(op.starting_balance)} XLM`;

    case "path_payment_strict_send":
      return `Swapped ${amount(op.source_amount)} ${asset(op.source_asset_type, op.source_asset_code, op.source_asset_issuer)} → ${amount(op.amount)} ${asset(op.asset_type, op.asset_code, op.asset_issuer)}`;

    case "path_payment_strict_receive":
      return `Swapped up to ${amount(op.source_max)} ${asset(op.source_asset_type, op.source_asset_code, op.source_asset_issuer)} → ${amount(op.amount)} ${asset(op.asset_type, op.asset_code, op.asset_issuer)}`;

    case "manage_sell_offer": {
      const price = op.price as string;
      const isDelete = parseFloat(String(op.amount)) === 0;
      if (isDelete) return `Cancelled sell offer #${op.offer_id}`;
      return `Sell ${amount(op.amount)} ${asset(op.selling_asset_type, op.selling_asset_code, op.selling_asset_issuer)} @ ${price} ${asset(op.buying_asset_type, op.buying_asset_code, op.buying_asset_issuer)}`;
    }

    case "manage_buy_offer": {
      const isDelete = parseFloat(String(op.amount)) === 0;
      if (isDelete) return `Cancelled buy offer #${op.offer_id}`;
      return `Buy ${amount(op.amount)} ${asset(op.buying_asset_type, op.buying_asset_code, op.buying_asset_issuer)} @ ${op.price} ${asset(op.selling_asset_type, op.selling_asset_code, op.selling_asset_issuer)}`;
    }

    case "create_passive_sell_offer":
      return `Passive sell ${amount(op.amount)} ${asset(op.selling_asset_type, op.selling_asset_code, op.selling_asset_issuer)} @ ${op.price} ${asset(op.buying_asset_type, op.buying_asset_code, op.buying_asset_issuer)}`;

    case "set_options": {
      const parts: string[] = [];
      if (op.inflation_dest) parts.push(`set inflation dest`);
      if (op.home_domain) parts.push(`set domain to ${op.home_domain}`);
      if (op.signer_key) parts.push(`updated signer`);
      if (op.master_key_weight !== undefined) parts.push(`set master weight`);
      return parts.length > 0 ? `Account options: ${parts.join(", ")}` : "Updated account options";
    }

    case "change_trust": {
      const limit = parseFloat(String(op.limit ?? "0"));
      const assetName = asset(op.asset_type, op.asset_code, op.asset_issuer);
      return limit === 0
        ? `Removed trustline for ${assetName}`
        : `Added trustline for ${assetName}`;
    }

    case "allow_trust":
      return `${op.authorize ? "Authorized" : "Revoked"} trustline for ${op.asset_code} on ${addr(op.trustor)}`;

    case "account_merge":
      return `Merged account into ${addr(op.into)}`;

    case "inflation":
      return "Ran inflation";

    case "manage_data": {
      const key = typeof op.name === "string" ? op.name : "entry";
      return op.value ? `Set data entry "${key}"` : `Removed data entry "${key}"`;
    }

    case "bump_sequence":
      return `Bumped sequence to ${op.bump_to}`;

    case "create_claimable_balance":
      return `Created claimable balance: ${amount(op.amount)} ${asset(op.asset_type, op.asset_code, op.asset_issuer)}`;

    case "claim_claimable_balance":
      return `Claimed balance ${addr(op.balance_id as string)}`;

    case "clawback":
      return `Clawed back ${amount(op.amount)} ${asset(op.asset_type, op.asset_code, op.asset_issuer)} from ${addr(op.from)}`;

    case "clawback_claimable_balance":
      return `Clawed back claimable balance`;

    case "set_trust_line_flags":
      return `Updated trustline flags for ${op.asset_code} on ${addr(op.trustor)}`;

    case "liquidity_pool_deposit":
      return `Deposited into liquidity pool`;

    case "liquidity_pool_withdraw":
      return `Withdrew from liquidity pool`;

    case "invoke_host_function": {
      const { displayName } = parseSorobanFunctionType(op.function as string);
      return displayName;
    }

    case "extend_footprint_ttl":
      return `Extended contract TTL to ledger ${op.extend_to}`;

    case "restore_footprint":
      return "Restored contract footprint";

    default:
      return type?.replace(/_/g, " ") ?? "Unknown operation";
  }
}

/**
 * Returns a concise summary of a transaction based on its operations.
 * Used for activity feeds and human-readable lists.
 */
export function getTransactionSummary(operations: Op[]): string {
  if (!operations || operations.length === 0) return "Empty transaction";
  if (operations.length === 1) return getOperationSummary(operations[0]);

  // Group by type for multi-op transactions
  const types = new Set(operations.map((op) => op.type as string));
  if (types.size === 1) {
    const type = [...types][0];
    return `${operations.length}× ${getOperationSummary(operations[0])}`.replace(
      /^(\d+)×/,
      `$1 ${type === "payment" ? "payments" : "operations"}: `
    );
  }

  return `${operations.length} operations`;
}
