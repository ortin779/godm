import { Progress } from "@/components/ui/progress";
import { calcPercent, cn } from "@/lib/utils";
import type { DownloadStatus } from "@/lib/types";

interface DownloadProgressProps {
  downloaded: number;
  totalSize: number;
  status: DownloadStatus;
}

export function DownloadProgress({
  downloaded,
  totalSize,
  status,
}: DownloadProgressProps) {
  const percent = calcPercent(downloaded, totalSize);

  // Active → primary (indigo). At-rest states → muted so they read as paused/waiting.
  const indicatorClass = cn({
    "bg-primary": status === "active",
    "bg-muted-foreground/40": status === "paused" || status === "pending",
    "bg-destructive/70": status === "error",
  });

  return (
    <div className="flex items-center gap-2.5 w-full">
      <Progress
        value={percent}
        className="flex-1 h-1"
        indicatorClassName={indicatorClass}
      />
      <span className="tabular-nums text-[11px] text-muted-foreground w-8 text-right">
        {percent}%
      </span>
    </div>
  );
}
