import { useState, useEffect, useCallback } from "react";
import { Download, Config, DownloadInfo } from "@/lib/types";
import { useWailsEvent } from "./useWailsEvents";
import {
  AddDownload,
  ProbeDownload,
  GetDownloads,
  GetCompletedDownloads,
  PauseDownload,
  ResumeDownload,
  CancelDownload,
  RevealFile,
  GetConfig,
  UpdateConfig,
} from "../../wailsjs/go/main/App";

// Cast Wails-generated types to our stricter types (status is string in generated code)
function asDownload(d: unknown): Download {
  return d as Download;
}
function asDownloads(ds: unknown[] | null | undefined): Download[] {
  return (ds ?? []).map(asDownload);
}
function asConfig(c: unknown): Config {
  return c as Config;
}
function asDownloadInfo(i: unknown): DownloadInfo {
  return i as DownloadInfo;
}

export function useDownloads() {
  const [active, setActive] = useState<Download[]>([]);
  const [completed, setCompleted] = useState<Download[]>([]);
  const [config, setConfig] = useState<Config | null>(null);
  const [search, setSearch] = useState("");

  // Initial load
  useEffect(() => {
    GetDownloads().then((ds) => setActive(asDownloads(ds)));
    GetCompletedDownloads().then((ds) => setCompleted(asDownloads(ds)));
    GetConfig().then((c) => setConfig(asConfig(c)));
  }, []);

  // Real-time progress updates from Go via Wails events
  const handleProgress = useCallback((updated: Download) => {
    if (updated.status === "completed") {
      setActive((prev) => prev.filter((d) => d.id !== updated.id));
      setCompleted((prev) => {
        const exists = prev.some((d) => d.id === updated.id);
        return exists ? prev : [updated, ...prev];
      });
    } else {
      setActive((prev) =>
        prev.map((d) => {
          if (d.id !== updated.id) return d;
          // The progress loop emits every 500ms. A stale "active" snapshot
          // emitted just before a pause was processed by Go must not revert
          // a "paused" status that was set by an explicit user action.
          if (d.status === "paused" && updated.status === "active") return d;
          return updated;
        })
      );
    }
  }, []);

  useWailsEvent<Download>("download:progress", handleProgress);

  const probeDownload = useCallback(async (url: string): Promise<DownloadInfo> => {
    const info = await ProbeDownload(url);
    return asDownloadInfo(info);
  }, []);

  const addDownload = useCallback(async (url: string, fileName: string) => {
    await AddDownload(url, fileName);
    // Refresh both lists from Go so we get the real status.
    // Fast downloads may already be in 'completed' before this line runs.
    const [actives, completeds] = await Promise.all([
      GetDownloads(),
      GetCompletedDownloads(),
    ]);
    setActive(asDownloads(actives));
    setCompleted(asDownloads(completeds));
  }, []);

  const pause = useCallback(async (id: string) => {
    try {
      await PauseDownload(id);
      setActive((prev) =>
        prev.map((d) => (d.id === id ? { ...d, status: "paused" as const } : d))
      );
    } catch (err) {
      console.error("Pause failed:", err);
    }
  }, []);

  const resume = useCallback(async (id: string) => {
    try {
      await ResumeDownload(id);
      setActive((prev) =>
        prev.map((d) => (d.id === id ? { ...d, status: "active" as const } : d))
      );
    } catch (err) {
      console.error("Resume failed:", err);
    }
  }, []);

  const cancel = useCallback(async (id: string) => {
    try {
      await CancelDownload(id);
      setActive((prev) => prev.filter((d) => d.id !== id));
    } catch (err) {
      console.error("Cancel failed:", err);
    }
  }, []);

  const revealFile = useCallback(async (id: string) => {
    await RevealFile(id);
  }, []);

  const saveConfig = useCallback(async (cfg: Config) => {
    const saved = asConfig(await UpdateConfig(cfg as never));
    setConfig(saved);
    return saved;
  }, []);

  const filteredActive = active.filter((d) =>
    d.fileName.toLowerCase().includes(search.toLowerCase())
  );
  const filteredCompleted = completed.filter((d) =>
    d.fileName.toLowerCase().includes(search.toLowerCase())
  );

  return {
    active: filteredActive,
    completed: filteredCompleted,
    config,
    search,
    setSearch,
    probeDownload,
    addDownload,
    pause,
    resume,
    cancel,
    revealFile,
    saveConfig,
  };
}
