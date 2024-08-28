export const scrambleAndObscureWords = (input: string, seed: string): string => {
  const seedInt = parseInt(seed.slice(0, 8), 16); // Convert first 8 hex digits of seed to an integer

  const words = input.split(/\s+/);
  words.forEach((word, i) => {
    const letters = Array.from(word);
    const alphabeticLetters: string[] = [];
    const nonAlphabeticMapping: { [index: number]: string } = {};

    letters.forEach((letter, idx) => {
      if (/[a-zA-Z]/.test(letter)) {
        alphabeticLetters.push(letter);
      } else {
        nonAlphabeticMapping[idx] = letter;
      }
    });

    // Shuffle alphabetic letters
    for (let i = alphabeticLetters.length - 1; i > 0; i--) {
      const j = Math.floor(Math.random() * (i + 1));
      [alphabeticLetters[i], alphabeticLetters[j]] = [alphabeticLetters[j], alphabeticLetters[i]];
    }

    let j = 0;
    for (let k = 0; k < letters.length; k++) {
      if (nonAlphabeticMapping.hasOwnProperty(k)) {
        letters[k] = nonAlphabeticMapping[k];
      } else {
        letters[k] = alphabeticLetters[j];
        j++;
      }
    }

    words[i] = letters.join('');
  });

  return words.join(' ');
}

