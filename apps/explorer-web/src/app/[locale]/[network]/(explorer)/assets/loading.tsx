import { AssetCardSkeleton } from "@/components/cards/asset-card";
import { LoadingCard } from "@/components/common/loading-card";

export default function AssetsLoading() {
  return (
    <div className="space-y-6">
      <LoadingCard rows={2} />
      <div className="space-y-2">
        {Array.from({ length: 10 }).map((_, i) => (
          <AssetCardSkeleton key={i} />
        ))}
      </div>
    </div>
  );
}
