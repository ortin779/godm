// Types matching the Go downloader package structs

export type DownloadStatus =
  | "pending"
  | "active"
  | "paused"
  | "completed"
  | "cancelled"
  | "error";

export interface Download {
  id: string;
  url: string;
  fileName: string;
  totalSize: number;
  downloaded: number;
  speed: number; // bytes per second
  status: DownloadStatus;
  parts: number;
  error?: string;
  createdAt: string; // ISO timestamp
}

export interface Config {
  maxConcurrentDownloads: number;
  maxPartsPerDownload: number;
  downloadDir: string;
}

export interface DownloadInfo {
  url: string;
  fileName: string;
  totalSize: number;
  contentType: string;
  resumable: boolean;
}
