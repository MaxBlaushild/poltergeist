import React from 'react';

type ContentMapMarkersMovedNoticeProps = {
  subject: string;
};

export const ContentMapMarkersMovedNotice = ({
  subject,
}: ContentMapMarkersMovedNoticeProps) => {
  return (
    <section className="mb-6 rounded-lg border border-amber-200 bg-amber-50 p-4 shadow-sm">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h2 className="text-lg font-semibold text-amber-950">
            Marker Management Moved
          </h2>
          <p className="mt-1 max-w-3xl text-sm text-amber-900">
            {subject} now live on the dedicated Content Map Markers page, where
            defaults and zone-kind overrides are managed together.
          </p>
        </div>
        <a
          href="/content-map-markers"
          className="rounded-md bg-amber-900 px-3 py-2 text-sm font-medium text-white"
        >
          Open Content Map Markers
        </a>
      </div>
    </section>
  );
};

export default ContentMapMarkersMovedNotice;
