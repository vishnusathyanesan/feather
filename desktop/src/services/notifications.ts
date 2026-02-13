import { isPermissionGranted, requestPermission, sendNotification } from "@tauri-apps/plugin-notification";

let permissionGranted = false;

export async function initNotifications() {
  try {
    permissionGranted = await isPermissionGranted();
    if (!permissionGranted) {
      const permission = await requestPermission();
      permissionGranted = permission === "granted";
    }
  } catch {
    // Not in Tauri context (web dev)
  }
}

export function notify(title: string, body: string) {
  if (permissionGranted && document.hidden) {
    try {
      sendNotification({ title, body });
    } catch {
      // Fallback: use web notification
      if (Notification.permission === "granted") {
        new Notification(title, { body });
      }
    }
  }
}
