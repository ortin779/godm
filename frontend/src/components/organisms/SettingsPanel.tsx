import { useState, useEffect } from "react";
import { FolderOpen, X } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import type { Config } from "@/lib/types";

interface SettingsPanelProps {
  config: Config | null;
  onSave: (cfg: Config) => Promise<Config>;
  onClose: () => void;
}

export function SettingsPanel({ config, onSave, onClose }: SettingsPanelProps) {
  const [form, setForm] = useState<Config>({
    maxConcurrentDownloads: 3,
    maxPartsPerDownload: 4,
    downloadDir: "",
  });
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    if (config) setForm(config);
  }, [config]);

  const handleSave = async () => {
    setSaving(true);
    try {
      await onSave(form);
      onClose();
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="flex flex-col h-full">
      {/* Panel header */}
      <div className="flex items-center justify-between px-5 py-4 border-b border-border shrink-0">
        <h2 className="text-sm font-semibold">Settings</h2>
        <Button variant="ghost" size="icon" onClick={onClose} title="Close">
          <X className="h-4 w-4" />
        </Button>
      </div>

      {/* Form fields */}
      <div className="flex-1 overflow-y-auto px-5 py-5 flex flex-col gap-6">
        <div className="flex flex-col gap-1.5">
          <label className="text-sm font-medium">Max Concurrent Downloads</label>
          <Input
            type="number"
            min={1}
            max={10}
            value={form.maxConcurrentDownloads}
            onChange={(e) =>
              setForm((f) => ({
                ...f,
                maxConcurrentDownloads: Number(e.target.value),
              }))
            }
          />
          <p className="text-xs text-muted-foreground">
            How many files download simultaneously (1–10)
          </p>
        </div>

        <div className="flex flex-col gap-1.5">
          <label className="text-sm font-medium">Max Parts Per Download</label>
          <Input
            type="number"
            min={1}
            max={16}
            value={form.maxPartsPerDownload}
            onChange={(e) =>
              setForm((f) => ({
                ...f,
                maxPartsPerDownload: Number(e.target.value),
              }))
            }
          />
          <p className="text-xs text-muted-foreground">
            Parallel chunks per multipart download (1–16)
          </p>
        </div>

        <div className="flex flex-col gap-1.5">
          <label className="text-sm font-medium">Download Directory</label>
          <div className="flex gap-2">
            <Input
              value={form.downloadDir}
              onChange={(e) =>
                setForm((f) => ({ ...f, downloadDir: e.target.value }))
              }
              placeholder="/Users/you/Downloads"
            />
            <Button variant="outline" size="icon" title="Browse folder">
              <FolderOpen className="h-4 w-4" />
            </Button>
          </div>
          <p className="text-xs text-muted-foreground">
            Where downloaded files are saved
          </p>
        </div>
      </div>

      {/* Footer actions */}
      <div className="flex items-center justify-end gap-2 px-5 py-4 border-t border-border shrink-0">
        <Button variant="outline" onClick={onClose} disabled={saving}>
          Cancel
        </Button>
        <Button onClick={handleSave} disabled={saving}>
          {saving ? "Saving..." : "Save"}
        </Button>
      </div>
    </div>
  );
}
