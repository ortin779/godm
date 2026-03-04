import { useState } from "react";
import { Settings } from "lucide-react";
import { useDownloads } from "@/hooks/useDownloads";
import { SearchBar } from "@/components/molecules/SearchBar";
import { AddDownloadDialog } from "@/components/molecules/AddDownloadDialog";
import { DownloadTabs } from "@/components/organisms/DownloadTabs";
import { SettingsPanel } from "@/components/organisms/SettingsPanel";
import { Button } from "@/components/ui/button";

function App() {
  const [settingsOpen, setSettingsOpen] = useState(false);

  const {
    active,
    completed,
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
  } = useDownloads();

  return (
    <div className="flex flex-col h-screen bg-background">
      {/* Header */}
      <header className="flex items-center gap-3 px-6 py-4 border-b border-border shrink-0">
        <SearchBar value={search} onChange={setSearch} />
        <AddDownloadDialog
          onProbe={probeDownload}
          onAdd={addDownload}
        />
        <Button
          variant={settingsOpen ? "secondary" : "outline"}
          size="icon"
          title="Settings"
          onClick={() => setSettingsOpen((v) => !v)}
        >
          <Settings className="h-4 w-4" />
        </Button>
      </header>

      {/* Body — relative so the overlay is scoped here, below the header */}
      <div className="relative flex flex-1 overflow-hidden">
        <main className="flex flex-col flex-1 overflow-hidden px-6 py-4">
          <DownloadTabs
            active={active}
            completed={completed}
            onPause={pause}
            onResume={resume}
            onCancel={cancel}
            onReveal={revealFile}
          />
        </main>

        {/* Backdrop */}
        <div
          className={`absolute inset-0 z-40 bg-black/20 transition-opacity duration-300 ${
            settingsOpen ? "opacity-100" : "opacity-0 pointer-events-none"
          }`}
          onClick={() => setSettingsOpen(false)}
        />

        {/* Settings slide-over panel */}
        <aside
          className={`absolute right-0 top-0 h-full w-1/3 z-50 bg-card border-l border-border shadow-2xl flex flex-col transition-transform duration-300 ease-in-out ${
            settingsOpen ? "translate-x-0" : "translate-x-full"
          }`}
        >
          <SettingsPanel
            config={config}
            onSave={saveConfig}
            onClose={() => setSettingsOpen(false)}
          />
        </aside>
      </div>
    </div>
  );
}

export default App;
