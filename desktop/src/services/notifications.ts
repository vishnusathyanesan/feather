let useTauri = false;
let permissionGranted = false;

export async function initNotifications() {
  // Try Tauri native notifications first
  try {
    const tauri = await import("@tauri-apps/plugin-notification");
    permissionGranted = await tauri.isPermissionGranted();
    if (!permissionGranted) {
      const permission = await tauri.requestPermission();
      permissionGranted = permission === "granted";
    }
    useTauri = true;
    return;
  } catch {
    // Not in Tauri context — fall through to browser API
  }

  // Browser Notification API fallback
  if ("Notification" in window) {
    if (Notification.permission === "granted") {
      permissionGranted = true;
    } else if (Notification.permission !== "denied") {
      const permission = await Notification.requestPermission();
      permissionGranted = permission === "granted";
    }
  }
}

export function notify(title: string, body: string) {
  if (!permissionGranted) return;

  if (useTauri) {
    import("@tauri-apps/plugin-notification")
      .then((tauri) => tauri.sendNotification({ title, body }))
      .catch(() => showBrowserNotification(title, body));
  } else {
    showBrowserNotification(title, body);
  }
}

function showBrowserNotification(title: string, body: string) {
  if ("Notification" in window && Notification.permission === "granted") {
    new Notification(title, { body });
  }
}
