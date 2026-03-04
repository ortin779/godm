import { Pause, Play, X, AlertCircle, FolderOpen } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { DownloadProgress } from "@/components/atoms/DownloadProgress";
import { FileSize } from "@/components/atoms/FileSize";
import type { Download } from "@/lib/types";

interface DownloadItemProps {
  download: Download;
  onPause: (id: string) => void;
  onResume: (id: string) => void;
  onCancel: (id: string) => void;
  onReveal?: (id: string) => void;
  showActions?: boolean;
}

export function DownloadItem({
  download,
  onPause,
  onResume,
  onCancel,
  onReveal,
  showActions = true,
}: DownloadItemProps) {
  const { id, fileName, status, downloaded, totalSize, speed, parts, error } =
    download;

  const isActive = status === "active";
  const isPaused = status === "paused";
  const isCompleted = status === "completed";

  return (
    <div className="flex flex-col gap-2 bg-card border border-border rounded-lg px-4 py-3 shadow-sm hover:shadow-md hover:border-primary/25 transition-all duration-150">
      <div className="flex items-center justify-between gap-3">
        {/* Left: filename + badges */}
        <div className="flex items-center gap-2 min-w-0">
          <span className="text-sm font-medium truncate leading-snug" title={fileName}>
            {fileName}
          </span>
          {parts > 1 && (
            <Badge variant="secondary" className="text-[10px] shrink-0 font-normal">
              {parts} parts
            </Badge>
          )}
          {status === "error" && (
            <Badge variant="destructive" className="text-[10px] shrink-0 gap-1">
              <AlertCircle className="h-3 w-3" />
              Error
            </Badge>
          )}
          {isCompleted && (
            <Badge variant="success" className="text-[10px] shrink-0">
              Done
            </Badge>
          )}
        </div>

        {/* Right: size/speed + action buttons */}
        <div className="flex items-center gap-1 shrink-0">
          <FileSize
            downloaded={downloaded}
            totalSize={totalSize}
            speed={isActive ? speed : 0}
          />

          <div className="flex items-center gap-0.5 ml-1">
            {/* Show in Finder — completed downloads only */}
            {isCompleted && onReveal && (
              <Button
                size="icon"
                variant="ghost"
                className="h-7 w-7 text-muted-foreground hover:text-primary hover:bg-accent transition-all duration-150"
                title="Show in Finder"
                onClick={() => onReveal(id)}
              >
                <FolderOpen className="h-3.5 w-3.5" />
              </Button>
            )}

            {/* Pause / Resume / Cancel — active downloads only */}
            {showActions && !isCompleted && (
              <>
                {(isActive || isPaused) && (
                  <Button
                    size="icon"
                    variant="ghost"
                    className="h-7 w-7 text-muted-foreground hover:text-primary hover:bg-accent transition-all duration-150"
                    title={isActive ? "Pause" : "Resume"}
                    onClick={() => (isActive ? onPause(id) : onResume(id))}
                  >
                    {isActive ? (
                      <Pause className="h-3.5 w-3.5" />
                    ) : (
                      <Play className="h-3.5 w-3.5" />
                    )}
                  </Button>
                )}
                <Button
                  size="icon"
                  variant="ghost"
                  className="h-7 w-7 text-muted-foreground hover:text-destructive hover:bg-destructive/10 transition-all duration-150"
                  title="Cancel"
                  onClick={() => onCancel(id)}
                >
                  <X className="h-3.5 w-3.5" />
                </Button>
              </>
            )}
          </div>
        </div>
      </div>

      {!isCompleted && (
        <DownloadProgress
          downloaded={downloaded}
          totalSize={totalSize}
          status={status}
        />
      )}

      {error && (
        <p className="text-xs text-destructive/80">{error}</p>
      )}
    </div>
  );
}
