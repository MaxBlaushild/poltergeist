// R-8.4. Every claim on this page is deliberate — R-8.4 requires stating the
// material, that parts are not certified food-grade or laboratory-grade,
// that the buyer is responsible for evaluating suitability for their
// system, and that no product is rated to suspend load over water. No
// claims about livestock safety, and the phrase "reef safe" must never
// appear anywhere on this page.
export default function MaterialsAndCare() {
  return (
    <div className="max-w-2xl space-y-6">
      <h1 className="text-2xl font-bold">Materials &amp; care</h1>

      <section className="space-y-2">
        <h2 className="font-semibold">Material</h2>
        <p className="text-reef-ink/80">
          Every part in this catalog is printed in PETG. It is not certified food-grade and not
          certified laboratory-grade.
        </p>
      </section>

      <section className="space-y-2">
        <h2 className="font-semibold">Your responsibility</h2>
        <p className="text-reef-ink/80">
          You are responsible for evaluating whether a given part is suitable for your specific
          system before installing it — water chemistry, livestock, and equipment setups vary too
          much for us to make that determination for you.
        </p>
      </section>

      <section className="space-y-2">
        <h2 className="font-semibold">Load rating</h2>
        <p className="text-reef-ink/80">
          No product in this catalog is rated to suspend load over water. Frag racks and clips are
          designed to hold their own weight plus the frag plugs or mesh they're built for — nothing
          more.
        </p>
      </section>

      <section className="space-y-2">
        <h2 className="font-semibold">Care</h2>
        <p className="text-reef-ink/80">
          Rinse parts before first use. PETG holds up well to saltwater exposure over time, but
          inspect magnet pockets periodically for corrosion or a loosened fit.
        </p>
      </section>
    </div>
  );
}
