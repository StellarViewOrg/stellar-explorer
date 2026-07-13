import { TransactionBuilder } from "@stellar/stellar-sdk";
import { NETWORKS } from "@/lib/constants/networks";
import type { NetworkKey } from "@/types";

// Convert camelCase operation types from SDK to snake_case translation keys
function toSnakeCase(str: string): string {
  return str.replace(/([A-Z])/g, "_$1").toLowerCase();
}

export interface DecodedTransaction {
  sourceAccount: string;
  operationType: string;
  memo: string | null;
  memoType: string;
  amount: string | null;
  asset: string | null;
  destination: string | null;
}

export function decodeTransactionEnvelope(
  envelopeXdr: string,
  network: NetworkKey = "mainnet"
): DecodedTransaction | null {
  if (!envelopeXdr) return null;

  try {
    const tx = TransactionBuilder.fromXDR(envelopeXdr, NETWORKS[network].passphrase);

    const sourceAccount =
      "source" in tx
        ? (tx.source as string)
        : "feeSource" in tx
          ? String((tx as unknown as { feeSource: string }).feeSource)
          : "";
    let memo: string | null = null;
    let memoType = "none";

    if ("memo" in tx && tx.memo) {
      const m = tx.memo as { type: string; value?: string | Buffer };
      memoType = m.type ?? "none";
      if (m.value) {
        memo = typeof m.value === "string" ? m.value : m.value.toString("hex");
      }
    }

    const ops =
      "operations" in tx ? (tx.operations as unknown as Array<Record<string, unknown>>) : [];
    const firstOp = ops[0];

    if (!firstOp) {
      return {
        sourceAccount,
        operationType: "unknown",
        memo,
        memoType,
        amount: null,
        asset: null,
        destination: null,
      };
    }

    const operationType = toSnakeCase(firstOp.type as string);
    let amount: string | null = null;
    let asset: string | null = null;
    let destination: string | null = null;

    if ("amount" in firstOp && firstOp.amount) {
      amount = String(firstOp.amount);
    }
    if ("startingBalance" in firstOp && firstOp.startingBalance) {
      amount = String(firstOp.startingBalance);
    }
    if ("asset" in firstOp && firstOp.asset) {
      const a = firstOp.asset as { code?: string; isNative?: () => boolean };
      asset = typeof a.isNative === "function" && a.isNative() ? "XLM" : (a.code ?? null);
    }
    if ("destAsset" in firstOp && firstOp.destAsset) {
      const a = firstOp.destAsset as { code?: string; isNative?: () => boolean };
      asset = typeof a.isNative === "function" && a.isNative() ? "XLM" : (a.code ?? null);
    }
    if ("destination" in firstOp && firstOp.destination) {
      destination = String(firstOp.destination);
    }

    return { sourceAccount, operationType, memo, memoType, amount, asset, destination };
  } catch {
    return null;
  }
}
