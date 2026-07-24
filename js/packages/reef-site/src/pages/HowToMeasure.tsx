// R-8.2: /how-to-measure. R-4.7 says measurement error is the projected top
// cause of returns — this page and each parameter's inline help/diagram
// (SchemaForm) are both part of addressing that, not cosmetic.
export default function HowToMeasure() {
  return (
    <div className="max-w-2xl space-y-6">
      <h1 className="text-2xl font-bold">How to measure your tank</h1>
      <p className="text-reef-ink/80">
        Every configurator parameter that depends on a physical measurement includes a "where to
        measure" diagram link next to the field. A few general tips:
      </p>
      <ul className="list-disc list-inside space-y-2 text-reef-ink/80">
        <li>Use digital calipers where possible — a tape measure can be off by a millimeter or two, which matters for a snug magnetic fit.</li>
        <li>Measure glass thickness at the rim, not the base — some tanks taper.</li>
        <li>For rim width, measure how far the rim extends inward from the tank's outer edge, not the total rim length.</li>
        <li>If your tank matches a model in our verified list, select it from the dropdown — it auto-fills glass and rim dimensions from sourced manufacturer specs.</li>
      </ul>
    </div>
  );
}
