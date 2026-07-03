import type { Horizon } from "@stellar/stellar-sdk";

export type AssetRecordExtended = Horizon.ServerApi.AssetRecord & {
  amount: string;
  num_accounts: number;
  _links?: { toml?: { href?: string }; [key: string]: unknown };
};
