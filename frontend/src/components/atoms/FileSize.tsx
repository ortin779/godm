import { formatBytes, formatSpeed } from "@/lib/utils";

interface FileSizeProps {
  downloaded: number;
  totalSize: number;
  speed: number;
}

export function FileSize({ downloaded, totalSize, speed }: FileSizeProps) {
  const sizeLabel =
    totalSize > 0
      ? `${formatBytes(downloaded)} / ${formatBytes(totalSize)}`
      : formatBytes(downloaded);

  return (
    <span className="text-xs text-muted-foreground whitespace-nowrap">
      {sizeLabel}
      {speed > 0 && (
        <span className="ml-2 text-primary font-medium">
          {formatSpeed(speed)}
        </span>
      )}
    </span>
  );
}
