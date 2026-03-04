import { useState } from "react";
import { Plus, ArrowLeft } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
  DialogFooter,
  DialogDescription,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { formatBytes } from "@/lib/utils";
import type { DownloadInfo } from "@/lib/types";

interface AddDownloadDialogProps {
  onProbe: (url: string) => Promise<DownloadInfo>;
  onAdd: (url: string, fileName: string) => Promise<void>;
}

type Step = "url" | "confirm";

export function AddDownloadDialog({ onProbe, onAdd }: AddDownloadDialogProps) {
  const [open, setOpen] = useState(false);
  const [step, setStep] = useState<Step>("url");
  const [url, setUrl] = useState("");
  const [info, setInfo] = useState<DownloadInfo | null>(null);
  const [fileName, setFileName] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const reset = () => {
    setStep("url");
    setUrl("");
    setInfo(null);
    setFileName("");
    setError("");
    setLoading(false);
  };

  const handleProbe = async () => {
    const trimmed = url.trim();
    if (!trimmed) {
      setError("Please enter a URL");
      return;
    }
    setLoading(true);
    setError("");
    try {
      const result = await onProbe(trimmed);
      setInfo(result);
      setFileName(result.fileName);
      setStep("confirm");
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : "Failed to fetch file info");
    } finally {
      setLoading(false);
    }
  };

  const handleDownload = async () => {
    if (!info) return;
    setLoading(true);
    setError("");
    try {
      await onAdd(info.url, fileName.trim() || info.fileName);
      reset();
      setOpen(false);
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : "Failed to start download");
    } finally {
      setLoading(false);
    }
  };

  return (
    <Dialog
      open={open}
      onOpenChange={(v) => {
        setOpen(v);
        if (!v) reset();
      }}
    >
      <DialogTrigger asChild>
        <Button size="sm" className="gap-1.5">
          <Plus className="h-4 w-4" />
          New
        </Button>
      </DialogTrigger>

      <DialogContent className="sm:max-w-md">
        {step === "url" ? (
          <>
            <DialogHeader>
              <DialogTitle>New Download</DialogTitle>
              <DialogDescription>
                Paste a URL — we'll detect the filename and size before starting.
              </DialogDescription>
            </DialogHeader>

            <div className="flex flex-col gap-2">
              <Input
                placeholder="https://example.com/file.zip"
                value={url}
                onChange={(e) => setUrl(e.target.value)}
                onKeyDown={(e) => e.key === "Enter" && handleProbe()}
                autoFocus
              />
              {error && <p className="text-xs text-destructive">{error}</p>}
            </div>

            <DialogFooter>
              <Button variant="outline" onClick={() => setOpen(false)} disabled={loading}>
                Cancel
              </Button>
              <Button onClick={handleProbe} disabled={loading}>
                {loading ? "Checking…" : "Next"}
              </Button>
            </DialogFooter>
          </>
        ) : (
          <>
            <DialogHeader>
              <DialogTitle>Confirm Download</DialogTitle>
              <DialogDescription className="break-all text-xs truncate">
                {info?.url}
              </DialogDescription>
            </DialogHeader>

            <div className="flex flex-col gap-3">
              <div className="flex flex-col gap-1">
                <label className="text-xs font-medium text-muted-foreground">
                  File name
                </label>
                <Input
                  value={fileName}
                  onChange={(e) => setFileName(e.target.value)}
                  onKeyDown={(e) => e.key === "Enter" && handleDownload()}
                  autoFocus
                />
              </div>

              {info && info.totalSize > 0 && (
                <div className="flex flex-col gap-1">
                  <label className="text-xs font-medium text-muted-foreground">
                    File size
                  </label>
                  <span className="text-sm">{formatBytes(info.totalSize)}</span>
                </div>
              )}

              {error && <p className="text-xs text-destructive">{error}</p>}
            </div>

            <DialogFooter>
              <Button
                variant="outline"
                onClick={() => { setStep("url"); setError(""); }}
                disabled={loading}
                className="gap-1"
              >
                <ArrowLeft className="h-3.5 w-3.5" />
                Back
              </Button>
              <Button onClick={handleDownload} disabled={loading}>
                {loading ? "Starting…" : "Download"}
              </Button>
            </DialogFooter>
          </>
        )}
      </DialogContent>
    </Dialog>
  );
}
