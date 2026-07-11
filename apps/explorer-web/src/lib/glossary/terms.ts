export type GlossaryLevel = "beginner" | "intermediate" | "advanced";

export interface GlossaryTermMeta {
  id: string;
  level: GlossaryLevel;
  category: string;
  relatedTerms?: string[];
  learnMoreUrl?: string;
}

export interface GlossaryEntry extends GlossaryTermMeta {
  term: string;
  shortDefinition: string;
  fullDefinition: string;
}

// Metadata for glossary terms (structure only - translations come from messages files)
export const glossaryTermsMeta: Record<string, GlossaryTermMeta> = {
  stellar: {
    id: "stellar",
    level: "beginner",
    category: "basics",
    relatedTerms: ["xlm", "ledger", "network"],
    learnMoreUrl: "https://stellar.org/learn/intro-to-stellar",
  },
  xlm: {
    id: "xlm",
    level: "beginner",
    category: "basics",
    relatedTerms: ["stellar", "transaction", "fee"],
    learnMoreUrl: "https://stellar.org/lumens",
  },
  account: {
    id: "account",
    level: "beginner",
    category: "basics",
    relatedTerms: ["publicKey", "balance", "signer"],
  },
  publicKey: {
    id: "publicKey",
    level: "beginner",
    category: "basics",
    relatedTerms: ["account", "secretKey", "signer"],
  },
  secretKey: {
    id: "secretKey",
    level: "beginner",
    category: "basics",
    relatedTerms: ["publicKey", "signer", "transaction"],
  },
  ledger: {
    id: "ledger",
    level: "beginner",
    category: "network",
    relatedTerms: ["transaction", "consensus", "stellar"],
  },
  transaction: {
    id: "transaction",
    level: "beginner",
    category: "transactions",
    relatedTerms: ["operation", "fee", "memo", "ledger"],
  },
  operation: {
    id: "operation",
    level: "intermediate",
    category: "transactions",
    relatedTerms: ["transaction", "payment", "trustline"],
  },
  payment: {
    id: "payment",
    level: "beginner",
    category: "operations",
    relatedTerms: ["operation", "asset", "trustline"],
  },
  asset: {
    id: "asset",
    level: "beginner",
    category: "assets",
    relatedTerms: ["xlm", "trustline", "issuer"],
  },
  trustline: {
    id: "trustline",
    level: "intermediate",
    category: "assets",
    relatedTerms: ["asset", "issuer", "balance"],
  },
  issuer: {
    id: "issuer",
    level: "intermediate",
    category: "assets",
    relatedTerms: ["asset", "trustline", "account"],
  },
  fee: {
    id: "fee",
    level: "beginner",
    category: "transactions",
    relatedTerms: ["stroops", "transaction", "xlm"],
  },
  stroops: {
    id: "stroops",
    level: "intermediate",
    category: "basics",
    relatedTerms: ["xlm", "fee"],
  },
  memo: {
    id: "memo",
    level: "beginner",
    category: "transactions",
    relatedTerms: ["transaction"],
  },
  signer: {
    id: "signer",
    level: "intermediate",
    category: "accounts",
    relatedTerms: ["account", "publicKey", "threshold"],
  },
  threshold: {
    id: "threshold",
    level: "intermediate",
    category: "accounts",
    relatedTerms: ["signer", "account", "multisig"],
  },
  multisig: {
    id: "multisig",
    level: "intermediate",
    category: "accounts",
    relatedTerms: ["signer", "threshold", "account"],
  },
  soroban: {
    id: "soroban",
    level: "advanced",
    category: "contracts",
    relatedTerms: ["contract", "wasm", "stellar"],
    learnMoreUrl: "https://soroban.stellar.org",
  },
  contract: {
    id: "contract",
    level: "advanced",
    category: "contracts",
    relatedTerms: ["soroban", "wasm", "storage"],
  },
  wasm: {
    id: "wasm",
    level: "advanced",
    category: "contracts",
    relatedTerms: ["soroban", "contract"],
  },
  xdr: {
    id: "xdr",
    level: "advanced",
    category: "technical",
    relatedTerms: ["transaction", "ledger"],
  },
  horizon: {
    id: "horizon",
    level: "intermediate",
    category: "technical",
    relatedTerms: ["stellar", "api", "stellarCore"],
  },
  stellarCore: {
    id: "stellarCore",
    level: "advanced",
    category: "technical",
    relatedTerms: ["horizon", "consensus", "validator"],
  },
  consensus: {
    id: "consensus",
    level: "advanced",
    category: "network",
    relatedTerms: ["ledger", "validator", "stellarCore"],
  },
  validator: {
    id: "validator",
    level: "advanced",
    category: "network",
    relatedTerms: ["consensus", "stellarCore", "node"],
  },
  dex: {
    id: "dex",
    level: "intermediate",
    category: "trading",
    relatedTerms: ["asset", "offer", "pathPayment"],
  },
  offer: {
    id: "offer",
    level: "intermediate",
    category: "trading",
    relatedTerms: ["dex", "asset", "orderbook"],
  },
  claimableBalance: {
    id: "claimableBalance",
    level: "advanced",
    category: "features",
    relatedTerms: ["asset", "trustline"],
  },
  sponsorship: {
    id: "sponsorship",
    level: "advanced",
    category: "features",
    relatedTerms: ["reserve", "account"],
  },
  reserve: {
    id: "reserve",
    level: "intermediate",
    category: "accounts",
    relatedTerms: ["account", "trustline", "sponsorship"],
  },
};

// Category metadata with icon names
export const glossaryCategoryMeta = [
  { id: "basics", icon: "BookOpen" },
  { id: "network", icon: "Globe" },
  { id: "accounts", icon: "User" },
  { id: "transactions", icon: "ArrowLeftRight" },
  { id: "operations", icon: "Layers" },
  { id: "assets", icon: "Coins" },
  { id: "trading", icon: "TrendingUp" },
  { id: "contracts", icon: "Code" },
  { id: "features", icon: "Puzzle" },
  { id: "technical", icon: "Terminal" },
];

// Helper to get all term IDs
export function getAllTermIds(): string[] {
  return Object.keys(glossaryTermsMeta);
}

// Helper to get terms by category
export function getTermIdsByCategory(category: string): string[] {
  return Object.values(glossaryTermsMeta)
    .filter((term) => term.category === category)
    .map((term) => term.id);
}

// Helper to get terms by level
export function getTermIdsByLevel(level: GlossaryLevel): string[] {
  return Object.values(glossaryTermsMeta)
    .filter((term) => term.level === level)
    .map((term) => term.id);
}

// Helper to get category IDs
export function getCategoryIds(): string[] {
  return glossaryCategoryMeta.map((cat) => cat.id);
}
