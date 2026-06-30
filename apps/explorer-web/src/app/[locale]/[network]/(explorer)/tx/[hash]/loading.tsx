import { LoadingCard } from "@/components/common/loading-card";

export default function TransactionDetailLoading() {
  return (
    <div className="space-y-6">
      <div className="space-y-2">
        <div className="bg-muted h-8 w-48 animate-pulse rounded" />
        <div className="bg-muted h-4 w-96 animate-pulse rounded" />
      </div>
      <LoadingCard rows={6} />
      <LoadingCard rows={4} />
    </div>
  );
}
