import React from 'react';

export type ContentDashboardMetric = {
  label: string;
  value: number | string;
  note?: string;
};

export type ContentDashboardBucket = {
  label: string;
  value: number | string;
  note?: string;
};

export type ContentDashboardSection = {
  title: string;
  note?: string;
  buckets: ContentDashboardBucket[];
  emptyLabel?: string;
};

type ContentDashboardProps = {
  title: string;
  subtitle?: string;
  status?: string;
  loading?: boolean;
  error?: string | null;
  metrics: ContentDashboardMetric[];
  sections: ContentDashboardSection[];
};

const formatValue = (value: number | string) =>
  typeof value === 'number' ? new Intl.NumberFormat().format(value) : value;

export const ContentDashboard = ({
  title,
  subtitle,
  status,
  loading = false,
  error,
  metrics,
  sections,
}: ContentDashboardProps) => {
  return (
    <section
      className="overflow-hidden rounded-2xl border p-5 text-white shadow-sm"
      style={{
        borderColor: '#1e293b',
        background:
          'linear-gradient(135deg, rgb(2, 6, 23) 0%, rgb(15, 23, 42) 55%, rgb(30, 41, 59) 100%)',
      }}
    >
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div className="max-w-3xl">
          <div className="text-xs font-semibold uppercase tracking-widest text-sky-200">
            {title}
          </div>
          {subtitle ? (
            <p className="mt-2 text-sm text-slate-200">{subtitle}</p>
          ) : null}
        </div>
        <div className="flex flex-wrap items-center gap-2 text-xs">
          {status ? (
            <span
              className="rounded-full border px-3 py-1 text-slate-100"
              style={{
                borderColor: 'rgba(255, 255, 255, 0.16)',
                backgroundColor: 'rgba(255, 255, 255, 0.08)',
              }}
            >
              {status}
            </span>
          ) : null}
          {loading ? (
            <span
              className="rounded-full border px-3 py-1 text-sky-100"
              style={{
                borderColor: 'rgba(125, 211, 252, 0.28)',
                backgroundColor: 'rgba(56, 189, 248, 0.14)',
              }}
            >
              Refreshing totals...
            </span>
          ) : null}
          {error ? (
            <span
              className="rounded-full border px-3 py-1 text-rose-100"
              style={{
                borderColor: 'rgba(253, 164, 175, 0.28)',
                backgroundColor: 'rgba(244, 63, 94, 0.14)',
              }}
            >
              Aggregate refresh failed
            </span>
          ) : null}
        </div>
      </div>

      <div className="mt-5 grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
        {metrics.map((metric) => (
          <div
            key={metric.label}
            className="rounded-xl border px-4 py-3"
            style={{
              borderColor: 'rgba(255, 255, 255, 0.12)',
              backgroundColor: 'rgba(255, 255, 255, 0.08)',
            }}
          >
            <div className="text-xs font-medium uppercase tracking-wide text-slate-300">
              {metric.label}
            </div>
            <div className="mt-2 text-3xl font-semibold text-white">
              {formatValue(metric.value)}
            </div>
            {metric.note ? (
              <div className="mt-1 text-xs text-slate-300">{metric.note}</div>
            ) : null}
          </div>
        ))}
      </div>

      <div className="mt-5 grid gap-3 xl:grid-cols-3">
        {sections.map((section) => (
          <div
            key={section.title}
            className="rounded-xl border p-4"
            style={{
              borderColor: 'rgba(255, 255, 255, 0.12)',
              backgroundColor: 'rgba(2, 6, 23, 0.28)',
            }}
          >
            <div className="mb-3">
              <div className="text-sm font-semibold text-white">
                {section.title}
              </div>
              {section.note ? (
                <div className="mt-1 text-xs text-slate-300">
                  {section.note}
                </div>
              ) : null}
            </div>

            {section.buckets.length === 0 ? (
              <div
                className="rounded-lg border border-dashed px-3 py-4 text-sm text-slate-300"
                style={{ borderColor: 'rgba(255, 255, 255, 0.14)' }}
              >
                {section.emptyLabel ?? 'No aggregate data yet.'}
              </div>
            ) : (
              <div className="space-y-2">
                {section.buckets.map((bucket) => (
                  <div
                    key={`${section.title}-${bucket.label}`}
                    className="rounded-lg border px-3 py-2"
                    style={{
                      borderColor: 'rgba(255, 255, 255, 0.1)',
                      backgroundColor: 'rgba(255, 255, 255, 0.06)',
                    }}
                  >
                    <div className="flex items-baseline justify-between gap-3">
                      <div className="min-w-0 text-sm text-slate-100">
                        {bucket.label}
                      </div>
                      <div className="shrink-0 text-lg font-semibold text-white">
                        {formatValue(bucket.value)}
                      </div>
                    </div>
                    {bucket.note ? (
                      <div className="mt-1 text-xs text-slate-300">
                        {bucket.note}
                      </div>
                    ) : null}
                  </div>
                ))}
              </div>
            )}
          </div>
        ))}
      </div>
    </section>
  );
};

export default ContentDashboard;
