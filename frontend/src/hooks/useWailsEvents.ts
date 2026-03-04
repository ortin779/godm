import { useEffect } from "react";
import { EventsOn } from "../../wailsjs/runtime/runtime";

type UnsubscribeFn = () => void;

/**
 * Subscribe to a Wails runtime event. Automatically unsubscribes on unmount.
 */
export function useWailsEvent<T>(
  eventName: string,
  handler: (data: T) => void
) {
  useEffect(() => {
    const unsubscribe = EventsOn(eventName, handler) as unknown as UnsubscribeFn;
    return () => {
      if (typeof unsubscribe === "function") unsubscribe();
    };
  }, [eventName, handler]);
}
