import type { Notification } from '../types';
import { VampireMark } from './VampireMark';

// Full-screen takeover for a GM broadcast. Dismissible by the player; it returns
// only if the GM pushes a new one.
export const NotificationTakeover = ({
  notification,
  onDismiss,
}: {
  notification: Notification;
  onDismiss: () => void;
}) => (
  <div className="fixed inset-0 z-50 flex items-center justify-center px-6 bg-blood-ink">
    <div className="max-w-md text-center">
      <VampireMark className="w-14 h-14 mx-auto mb-5" />
      {notification.title && (
        <h1 className="font-display text-3xl font-bold text-bone mb-4 leading-tight">
          {notification.title}
        </h1>
      )}
      <p className="text-bone/90 text-lg leading-relaxed whitespace-pre-wrap">{notification.body}</p>
      <button
        onClick={onDismiss}
        className="mt-8 px-8 py-3 rounded-md border border-blood text-blood-bright uppercase tracking-[0.2em] text-sm hover:bg-blood hover:text-bone transition-colors"
      >
        Understood
      </button>
    </div>
  </div>
);
