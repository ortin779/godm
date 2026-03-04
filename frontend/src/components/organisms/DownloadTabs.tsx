import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import { DownloadList } from "@/components/organisms/DownloadList";
import type { Download } from "@/lib/types";

interface DownloadTabsProps {
  active: Download[];
  completed: Download[];
  onPause: (id: string) => void;
  onResume: (id: string) => void;
  onCancel: (id: string) => void;
  onReveal: (id: string) => void;
}

export function DownloadTabs({
  active,
  completed,
  onPause,
  onResume,
  onCancel,
  onReveal,
}: DownloadTabsProps) {
  return (
    <Tabs defaultValue="downloading" className="flex flex-col h-full">
      <TabsList className="w-full shrink-0">
        <TabsTrigger value="downloading" className="flex-1 gap-1.5">
          All
          {active.length > 0 && (
            <span className="rounded-full bg-primary text-primary-foreground text-[10px] px-1.5 py-0.5 font-semibold">
              {active.length}
            </span>
          )}
        </TabsTrigger>
        <TabsTrigger value="completed" className="flex-1 gap-1.5">
          Completed
          {completed.length > 0 && (
            <span className="rounded-full bg-muted-foreground/20 text-muted-foreground text-[10px] px-1.5 py-0.5 font-semibold">
              {completed.length}
            </span>
          )}
        </TabsTrigger>
      </TabsList>

      {/* relative+absolute gives TabsContent a concrete height regardless of Radix internals */}
      <div className="relative flex-1 mt-2">
        <TabsContent value="downloading" className="absolute inset-0 flex flex-col">
          <DownloadList
            downloads={active}
            showActions
            emptyMessage="No downloads yet. Click New to add one."
            onPause={onPause}
            onResume={onResume}
            onCancel={onCancel}
          />
        </TabsContent>

        <TabsContent value="completed" className="absolute inset-0 flex flex-col">
          <DownloadList
            downloads={completed}
            showActions={false}
            emptyMessage="No completed downloads yet."
            onPause={onPause}
            onResume={onResume}
            onCancel={onCancel}
            onReveal={onReveal}
          />
        </TabsContent>
      </div>
    </Tabs>
  );
}
