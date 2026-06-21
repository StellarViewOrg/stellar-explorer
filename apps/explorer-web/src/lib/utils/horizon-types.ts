import type { Horizon } from "@stellar/stellar-sdk";

/**
 * Horizon operation records are discriminated by `type` but the SDK's union type
 * doesn't satisfy `Record<string, unknown>`, so components that render arbitrary
 * operation fields (OperationDetails, OperationSummary) use this bridge type.
 */
export type HorizonOperation = Record<string, unknown> & { type: string };

/** Cast a Horizon SDK operation record to the bridge type for generic rendering. */
export function asHorizonOperation(op: Horizon.ServerApi.OperationRecord): HorizonOperation {
  return op as unknown as HorizonOperation;
}

/** Type guard: true when the operation is an account_merge and exposes `.account`. */
export function isAccountMergeOp(
  op: Horizon.ServerApi.OperationRecord
): op is Horizon.ServerApi.OperationRecord & { account: string } {
  return op.type === "account_merge";
}

/** Cast any Horizon response to a plain record for display/serialization purposes. */
export function asRecord(value: unknown): Record<string, unknown> {
  return value as Record<string, unknown>;
}
