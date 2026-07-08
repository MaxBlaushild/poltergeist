import { VampireMark } from './VampireMark';

// Full-screen climax shown to every player once the GM triggers the Final Reveal
// act. Copy is a draft assembled from the canonical solution — edit freely.
export const FinalReveal = ({ onDone }: { onDone: () => void }) => (
  <div className="fixed inset-0 z-50 overflow-y-auto bg-blood-ink">
    <div className="min-h-full flex items-center justify-center px-4 py-12">
      <div className="w-full max-w-lg">
        <div className="text-center mb-8">
          <VampireMark className="w-14 h-14 mx-auto mb-4" />
          <p className="text-xs uppercase tracking-[0.4em] text-gold mb-2">The Final Reveal</p>
          <h1 className="font-display text-3xl md:text-4xl font-bold text-bone">
            What Truly Happened to Caspian Vael
          </h1>
        </div>

        <div className="flex flex-col gap-5 text-bone/90 leading-relaxed">
          <Reveal title="The Bound Heir">
            Caspian Vael was never his own. In secret, he was a thrall — bound to Marquess Gruber
            against his knowledge and his will, her blood and a binding agent worked into the
            ceremonial wine he drank at his investiture. Her purpose was simple and cold: a
            compliant heir, a throne ruled through a puppet, a reign that need never end.
          </Reveal>

          <Reveal title="The Rite of Undoing">
            But Caspian discovered the leash upon him, and he would not wear it. He sought the Rite
            of Undoing — an ancient severance that could break the bond and set him free. To reach
            it he leaned, unknowing, on a chain of quiet hands who never grasped what they were
            helping him do.
          </Reveal>

          <Reveal title="The Sabotage">
            He was betrayed before he ever began. A single page of the Rite was swapped — the work
            of Calven Rue, paid for by Ivara Saye. She believed she was stopping a different
            severance entirely; she thought she was protecting the Court. She was wrong. Her coin
            corrupted the ritual, and hers is the hand that knowingly set his death in motion.
          </Reveal>

          <Reveal title="The Mill Road">
            A Rite that draws blood from all five houses releases enormous force. Corrupted, that
            force had nowhere safe to go. It took Caspian at the old mill house — and it took Lena
            Ashford, a mortal woman who should never have been near it, with him. Two dead on the
            mill road, and a cover story that the heir had simply failed to appear.
          </Reveal>

          <Reveal title="What Still Festers">
            Doctor Thorne's work was turned to ends he never intended. Brother Aldric, who came too
            close to the truth, is not missing — he is held, alive, by the Marquess's own people.
            And Caspian was never the only puppet: Mara Voss, too, is bound in secret — her strings
            in the hands of Eiran Vox.
          </Reveal>

          <p className="text-center text-gold/90 italic mt-2">
            The Court survived another Crimson Toast. Whether it deserved to is another question.
          </p>
        </div>

        <button
          onClick={onDone}
          className="mt-10 w-full py-3 rounded-md bg-blood text-bone uppercase tracking-[0.2em] text-sm hover:bg-blood-bright"
        >
          Close
        </button>
      </div>
    </div>
  </div>
);

const Reveal = ({ title, children }: { title: string; children: React.ReactNode }) => (
  <section>
    <h2 className="font-heading text-gold text-sm uppercase tracking-[0.3em] mb-2">{title}</h2>
    <p>{children}</p>
  </section>
);
