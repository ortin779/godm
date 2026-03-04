import { Download as DownloadIcon } from "lucide-react";
import { DownloadItem } from "@/components/molecules/DownloadItem";
import type { Download } from "@/lib/types";

interface DownloadListProps {
  downloads: Download[];
  showActions?: boolean;
  emptyMessage?: string;
  onPause: (id: string) => void;
  onResume: (id: string) => void;
  onCancel: (id: string) => void;
  onReveal?: (id: string) => void;
}

export function DownloadList({
  downloads,
  showActions = true,
  emptyMessage = "No downloads",
  onPause,
  onResume,
  onCancel,
  onReveal,
}: DownloadListProps) {
  if (downloads.length === 0) {
    return (
      <div className="h-full flex flex-col items-center justify-center text-muted-foreground gap-3">
        <DownloadIcon className="h-10 w-10 opacity-30" />
        <p className="text-sm">{emptyMessage}</p>
      </div>
    );
  }

  return (
    <div className="h-full overflow-y-auto flex flex-col gap-2 py-1">
      {downloads.map((dl) => (
        <DownloadItem
          key={dl.id}
          download={dl}
          showActions={showActions}
          onPause={onPause}
          onResume={onResume}
          onCancel={onCancel}
          onReveal={onReveal}
        />
      ))}
    </div>
  );
}
