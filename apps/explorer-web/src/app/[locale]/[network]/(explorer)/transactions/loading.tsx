import { TransactionCardSkeleton } from "@/components/cards/transaction-card";
import { LoadingCard } from "@/components/common/loading-card";

export default function TransactionsLoading() {
  return (
    <div className="space-y-6">
      <LoadingCard rows={2} />
      <div className="space-y-2">
        {Array.from({ length: 10 }).map((_, i) => (
          <TransactionCardSkeleton key={i} />
        ))}
      </div>
    </div>
  );
}
